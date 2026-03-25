package gdrive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type Service struct {
	clientID     string
	clientSecret string
	redirectURL  string
	localPath    string
}

func New(clientID, clientSecret, redirectURL, localPath string) *Service {
	os.MkdirAll(localPath, 0755)
	return &Service{clientID: clientID, clientSecret: clientSecret, redirectURL: redirectURL, localPath: localPath}
}

func (s *Service) IsConfigured() bool {
	return s.clientID != "" && s.clientSecret != ""
}

// ── OAuth ──

func (s *Service) GetAuthURL(state string) string {
	p := url.Values{
		"client_id": {s.clientID}, "redirect_uri": {s.redirectURL},
		"response_type": {"code"}, "access_type": {"offline"}, "prompt": {"consent"},
		"state": {state},
		"scope": {"https://www.googleapis.com/auth/drive.file https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"},
	}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + p.Encode()
}

func (s *Service) ExchangeCode(code string) (*TokenInfo, string, error) {
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"code": {code}, "client_id": {s.clientID}, "client_secret": {s.clientSecret},
		"redirect_uri": {s.redirectURL}, "grant_type": {"authorization_code"},
	})
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	var r struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	if r.Error != "" {
		return nil, "", fmt.Errorf("%s: %s", r.Error, r.ErrorDesc)
	}
	email, _ := s.getUserEmail(r.AccessToken)
	return &TokenInfo{
		AccessToken: r.AccessToken, RefreshToken: r.RefreshToken,
		ExpiresAt: time.Now().Add(time.Duration(r.ExpiresIn) * time.Second).Unix(),
	}, email, nil
}

func (s *Service) RefreshAccessToken(refreshToken string) (*TokenInfo, error) {
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"refresh_token": {refreshToken}, "client_id": {s.clientID},
		"client_secret": {s.clientSecret}, "grant_type": {"refresh_token"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	if r.Error != "" {
		return nil, fmt.Errorf("refresh: %s", r.Error)
	}
	return &TokenInfo{
		AccessToken: r.AccessToken, RefreshToken: refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(r.ExpiresIn) * time.Second).Unix(),
	}, nil
}

func (s *Service) getUserEmail(token string) (string, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var info struct{ Email string `json:"email"` }
	json.NewDecoder(resp.Body).Decode(&info)
	return info.Email, nil
}

func (s *Service) GetValidToken(access, refresh string, expiresAt int64) (string, error) {
	if time.Now().Unix() < expiresAt-60 {
		return access, nil
	}
	t, err := s.RefreshAccessToken(refresh)
	if err != nil {
		return "", err
	}
	return t.AccessToken, nil
}

// ── Drive API ──

func (s *Service) CreateDriveFolder(token, name, parentID string) (string, error) {
	meta := map[string]interface{}{"name": name, "mimeType": "application/vnd.google-apps.folder"}
	if parentID != "" {
		meta["parents"] = []string{parentID}
	}
	body, _ := json.Marshal(meta)
	req, _ := http.NewRequest("POST", "https://www.googleapis.com/drive/v3/files", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	return r.ID, nil
}

func (s *Service) CreateUserRootFolder(token string) (string, error) {
	return s.CreateDriveFolder(token, "Reyna", "")
}

func (s *Service) EnsureSubjectFolder(token, rootID, subject string) (string, error) {
	if subject == "" {
		subject = "General"
	}
	q := fmt.Sprintf("name='%s' and '%s' in parents and mimeType='application/vnd.google-apps.folder' and trashed=false", subject, rootID)
	u := fmt.Sprintf("https://www.googleapis.com/drive/v3/files?q=%s&fields=files(id)", url.QueryEscape(q))
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r struct {
		Files []struct{ ID string `json:"id"` } `json:"files"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	if len(r.Files) > 0 {
		return r.Files[0].ID, nil
	}
	return s.CreateDriveFolder(token, subject, rootID)
}

func (s *Service) UploadFileToDrive(token, folderID, fileName, mimeType string, data []byte) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	meta, _ := w.CreatePart(map[string][]string{"Content-Type": {"application/json; charset=UTF-8"}})
	meta.Write([]byte(fmt.Sprintf(`{"name":"%s","parents":["%s"]}`, fileName, folderID)))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	file, _ := w.CreatePart(map[string][]string{"Content-Type": {mimeType}})
	file.Write(data)
	w.Close()
	req, _ := http.NewRequest("POST", "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&fields=id", &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "multipart/related; boundary="+w.Boundary())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r struct{ ID string `json:"id"` }
	json.NewDecoder(resp.Body).Decode(&r)
	return r.ID, nil
}

func (s *Service) DownloadFromDrive(token, fileID string) ([]byte, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?alt=media", fileID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *Service) GetDriveStorageUsed(token string) (int64, int64) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/drive/v3/about?fields=storageQuota", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 15 << 30
	}
	defer resp.Body.Close()
	var r struct {
		StorageQuota struct{ Usage, Limit string } `json:"storageQuota"`
	}
	json.NewDecoder(resp.Body).Decode(&r)
	var u, l int64
	fmt.Sscanf(r.StorageQuota.Usage, "%d", &u)
	fmt.Sscanf(r.StorageQuota.Limit, "%d", &l)
	if l == 0 {
		l = 15 << 30
	}
	return u, l
}

// ── Smart Upload (Drive if connected, local fallback) ──

func (s *Service) SmartUpload(token, driveRootID string, userID int64, subject, fileName, mimeType string, data []byte) (string, string, error) {
	if token != "" && driveRootID != "" && s.IsConfigured() {
		fid, err := s.EnsureSubjectFolder(token, driveRootID, subject)
		if err == nil {
			did, err := s.UploadFileToDrive(token, fid, fileName, mimeType, data)
			if err == nil {
				return did, fid, nil
			}
		}
	}
	lid, err := s.UploadFileLocal(userID, subject, fileName, data)
	fid, _ := s.CreateFolder(userID, subject)
	return lid, fid, err
}

// ── Local fallback ──

func (s *Service) CreateUserRoot(uid int64) (string, error) {
	os.MkdirAll(filepath.Join(s.localPath, fmt.Sprintf("user_%d", uid), "Reyna"), 0755)
	return fmt.Sprintf("local_root_%d", uid), nil
}

func (s *Service) CreateFolder(uid int64, name string) (string, error) {
	os.MkdirAll(filepath.Join(s.localPath, fmt.Sprintf("user_%d", uid), "Reyna", name), 0755)
	return fmt.Sprintf("local_%s", name), nil
}

func (s *Service) UploadFileLocal(uid int64, subject, name string, data []byte) (string, error) {
	folder := "General"
	if subject != "" {
		folder = subject
	}
	dir := filepath.Join(s.localPath, fmt.Sprintf("user_%d", uid), "Reyna", folder)
	os.MkdirAll(dir, 0755)
	p := filepath.Join(dir, name)
	if _, err := os.Stat(p); err == nil {
		ext := filepath.Ext(name)
		p = filepath.Join(dir, fmt.Sprintf("%s_v%d%s", strings.TrimSuffix(name, ext), time.Now().Unix(), ext))
	}
	os.WriteFile(p, data, 0644)
	return fmt.Sprintf("local_%d_%d", uid, time.Now().UnixNano()), nil
}

func (s *Service) GetStorageUsed(uid int64) int64 {
	var total int64
	filepath.Walk(filepath.Join(s.localPath, fmt.Sprintf("user_%d", uid), "Reyna"), func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

// SaveLocalFileData saves raw file bytes indexed by file ID for later Drive upload
func (s *Service) SaveLocalFileData(fileID int64, data []byte) error {
	dir := filepath.Join(s.localPath, "_raw")
	os.MkdirAll(dir, 0755)
	return os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d", fileID)), data, 0644)
}

// GetLocalFileData retrieves raw bytes for a file ID
func (s *Service) GetLocalFileData(fileID int64) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.localPath, "_raw", fmt.Sprintf("%d", fileID)))
}
