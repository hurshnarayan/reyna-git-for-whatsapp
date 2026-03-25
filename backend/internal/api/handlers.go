package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/reyna-bot/reyna-backend/internal/auth"
	"github.com/reyna-bot/reyna-backend/internal/config"
	"github.com/reyna-bot/reyna-backend/internal/db"
	"github.com/reyna-bot/reyna-backend/internal/gdrive"
	"github.com/reyna-bot/reyna-backend/internal/models"
	"github.com/reyna-bot/reyna-backend/internal/reyna"
)

type Server struct {
	cfg    *config.Config
	store  *db.Store
	drive  *gdrive.Service
	reyna  *reyna.Reyna
	mux    *http.ServeMux
}

func NewServer(cfg *config.Config, store *db.Store, drive *gdrive.Service) *Server {
	s := &Server{
		cfg:   cfg,
		store: store,
		drive: drive,
		reyna: reyna.New(),
		mux:   http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// CORS + JSON middleware wrapper
	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Content-Type", "application/json")
			if r.Method == "OPTIONS" {
				w.WriteHeader(200)
				return
			}
			h(w, r)
		}
	}

	// Auth-required middleware
	protected := func(h http.HandlerFunc) http.HandlerFunc {
		return wrap(func(w http.ResponseWriter, r *http.Request) {
			authMid := auth.Middleware(s.cfg.JWTSecret)
			authMid(http.HandlerFunc(h)).ServeHTTP(w, r)
		})
	}

	// ── Public routes ──
	s.mux.HandleFunc("/api/health", wrap(s.handleHealth))
	s.mux.HandleFunc("/api/auth/login", wrap(s.handleLogin))
	s.mux.HandleFunc("/api/auth/register", wrap(s.handleRegister))
	s.mux.HandleFunc("/api/auth/google", wrap(s.handleGoogleAuthStart))
	s.mux.HandleFunc("/api/auth/google/callback", wrap(s.handleGoogleAuthCallback))
	s.mux.HandleFunc("/api/waitlist", wrap(s.handleWaitlist))

	// ── WhatsApp Bot API (used by bot service) ──
	s.mux.HandleFunc("/api/bot/command", wrap(s.handleBotCommand))
	s.mux.HandleFunc("/api/bot/upload", wrap(s.handleBotUpload))

	// ── Protected routes (require JWT) ──
	s.mux.HandleFunc("/api/me", protected(s.handleMe))
	s.mux.HandleFunc("/api/auth/google/status", protected(s.handleGoogleStatus))
	s.mux.HandleFunc("/api/auth/google/connect", protected(s.handleGoogleConnect))
	s.mux.HandleFunc("/api/dashboard", protected(s.handleDashboard))
	s.mux.HandleFunc("/api/groups", protected(s.handleGroups))
	s.mux.HandleFunc("/api/files", protected(s.handleFiles))
	s.mux.HandleFunc("/api/files/search", protected(s.handleSearchFiles))
	s.mux.HandleFunc("/api/files/versions", protected(s.handleFileVersions))
	s.mux.HandleFunc("/api/files/upload", protected(s.handleUploadFile))
	s.mux.HandleFunc("/api/activity", protected(s.handleActivity))
}

// ── Health ──

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"service":   "reyna-backend",
		"version":   "0.1.0",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ── Auth ──

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, 405)
		return
	}
	var req struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, 400)
		return
	}
	if req.Phone == "" {
		http.Error(w, `{"error":"phone required"}`, 400)
		return
	}

	user, err := s.store.UpsertUser(req.Phone, req.Name)
	if err != nil {
		log.Printf("register error: %v", err)
		http.Error(w, `{"error":"registration failed"}`, 500)
		return
	}

	// Create Drive root folder
	rootID, _ := s.drive.CreateUserRoot(user.ID)
	s.store.UpdateUserGoogle(user.ID, user.Email, "", "", rootID)

	// Auto-link to any existing groups where this phone appeared
	s.store.AutoLinkUserToGroups(user.ID, req.Phone)

	token, err := auth.GenerateToken(user.ID, s.cfg.JWTSecret)
	if err != nil {
		http.Error(w, `{"error":"token generation failed"}`, 500)
		return
	}

	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, 405)
		return
	}
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" {
		http.Error(w, `{"error":"phone required"}`, 400)
		return
	}

	user, err := s.store.GetUserByPhone(req.Phone)
	if err != nil {
		http.Error(w, `{"error":"user not found, register first"}`, 404)
		return
	}

	// Auto-link: find all groups where this phone has shared files and add as member
	s.store.AutoLinkUserToGroups(user.ID, req.Phone)

	token, _ := auth.GenerateToken(user.ID, s.cfg.JWTSecret)
	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r)
	user, err := s.store.GetUserByID(uid)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, 404)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// ── Dashboard ──

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r)
	stats, err := s.store.GetDashboardStats(uid)
	if err != nil {
		http.Error(w, `{"error":"stats failed"}`, 500)
		return
	}
	storageUsed := s.drive.GetStorageUsed(uid)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats":        stats,
		"storage_used": storageUsed,
		"storage_limit": 15 * 1024 * 1024 * 1024, // 15 GB
	})
}

// ── Groups ──

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r)
	switch r.Method {
	case "GET":
		groups, err := s.store.GetUserGroups(uid)
		if err != nil {
			groups = []models.Group{}
		}
		json.NewEncoder(w).Encode(groups)

	case "POST":
		var req struct {
			WAID string `json:"wa_id"`
			Name string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.WAID == "" || req.Name == "" {
			http.Error(w, `{"error":"wa_id and name required"}`, 400)
			return
		}
		group, err := s.store.UpsertGroup(req.WAID, req.Name, uid)
		if err != nil {
			http.Error(w, `{"error":"create group failed"}`, 500)
			return
		}
		s.store.AddGroupMember(group.ID, uid, "", "admin")
		json.NewEncoder(w).Encode(group)
	default:
		http.Error(w, `{"error":"method not allowed"}`, 405)
	}
}

// ── Files ──

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r)
	groupIDStr := r.URL.Query().Get("group_id")
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 100
	}

	var files []models.File

	if groupIDStr != "" {
		gid, _ := strconv.ParseInt(groupIDStr, 10, 64)
		files, _ = s.store.GetGroupFiles(gid, limit)
	} else {
		// Get files from ALL groups the user is in
		groupIDs := s.store.GetUserGroupIDs(uid)
		if len(groupIDs) > 0 {
			files, _ = s.store.GetGroupsFiles(groupIDs, limit)
		}
		// Fallback: also include files user personally stored
		if len(files) == 0 {
			files, _ = s.store.GetUserFiles(uid, limit)
		}
	}

	if files == nil {
		files = []models.File{}
	}
	json.NewEncoder(w).Encode(files)
}

func (s *Server) handleSearchFiles(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r)
	query := r.URL.Query().Get("q")
	groupIDStr := r.URL.Query().Get("group_id")

	if query == "" {
		http.Error(w, `{"error":"query 'q' required"}`, 400)
		return
	}

	var files []models.File

	if groupIDStr != "" {
		gid, _ := strconv.ParseInt(groupIDStr, 10, 64)
		files, _ = s.store.FindFiles(gid, query)
	} else {
		// Search across all user's groups
		groupIDs := s.store.GetUserGroupIDs(uid)
		for _, gid := range groupIDs {
			gFiles, _ := s.store.FindFiles(gid, query)
			files = append(files, gFiles...)
		}
	}

	if files == nil {
		files = []models.File{}
	}
	json.NewEncoder(w).Encode(files)
}

func (s *Server) handleFileVersions(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("file_id")
	if fileIDStr == "" {
		http.Error(w, `{"error":"file_id required"}`, 400)
		return
	}
	fid, _ := strconv.ParseInt(fileIDStr, 10, 64)
	versions, err := s.store.GetFileVersions(fid)
	if err != nil || versions == nil {
		versions = []models.FileVersion{}
	}
	json.NewEncoder(w).Encode(versions)
}

func (s *Server) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, 405)
		return
	}
	uid := auth.GetUserID(r)

	var req models.AddFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, 400)
		return
	}

	// Resolve group
	group, err := s.store.GetGroupByWAID(req.GroupWAID)
	if err != nil {
		http.Error(w, `{"error":"group not found"}`, 404)
		return
	}

	// Upload to drive (real Drive if connected, local fallback)
	user, _ := s.store.GetUserByID(uid)
	accessToken := ""
	driveRootID := ""
	if user != nil {
		accessToken = user.GoogleToken
		driveRootID = user.DriveRootID
		if user.GoogleRefresh != "" {
			if t, err := s.drive.GetValidToken(user.GoogleToken, user.GoogleRefresh, 0); err == nil {
				accessToken = t
			}
		}
	}

	fileData := []byte(req.FileData)

	driveID, folderID, err := s.drive.SmartUpload(accessToken, driveRootID, uid, req.Subject, req.FileName, req.MimeType, fileData)
	if err != nil {
		log.Printf("upload error: %v", err)
		http.Error(w, `{"error":"upload failed"}`, 500)
		return
	}

	file := &models.File{
		GroupID:       group.ID,
		UserID:        uid,
		SharedByPhone: req.SharedByPhone,
		SharedByName:  req.SharedByName,
		FileName:      req.FileName,
		FileSize:      req.FileSize,
		MimeType:      req.MimeType,
		DriveFileID:   driveID,
		DriveFolderID: folderID,
		Subject:       req.Subject,
		WAMessageID:   req.WAMessageID,
	}

	saved, err := s.store.AddFile(file)
	if err != nil {
		http.Error(w, `{"error":"save file failed"}`, 500)
		return
	}

	s.store.LogActivity(group.ID, uid, "add", "/reyna add "+req.FileName, "stored")
	json.NewEncoder(w).Encode(saved)
}

// ── Activity ──

func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.URL.Query().Get("group_id")
	if groupIDStr == "" {
		http.Error(w, `{"error":"group_id required"}`, 400)
		return
	}
	gid, _ := strconv.ParseInt(groupIDStr, 10, 64)
	logs, err := s.store.GetActivityLog(gid, 50)
	if err != nil || logs == nil {
		logs = []models.ActivityLog{}
	}
	json.NewEncoder(w).Encode(logs)
}

// ── WhatsApp Bot Command Handler ──

func (s *Server) handleBotCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, 405)
		return
	}

	var req models.CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, 400)
		return
	}

	// Parse the command
	action, args := s.reyna.ProcessCommand(req.Command)

	// Get or create user from phone
	userName := req.UserName
	if userName == "" {
		userName = req.UserPhone
	}
	user, _ := s.store.UpsertUser(req.UserPhone, userName)

	// Get or create group
	group, err := s.store.GetGroupByWAID(req.GroupWAID)
	if err != nil {
		group, _ = s.store.UpsertGroup(req.GroupWAID, "WhatsApp Group", user.ID)
	}
	// Always ensure user is a member of this group
	if group != nil {
		s.store.AddGroupMember(group.ID, user.ID, req.UserPhone, "member")
	}

	groupID := int64(0)
	if group != nil {
		groupID = group.ID
	}

	resp := models.CommandResponse{}

	switch action {
	case "add":
		fileName := args
		if fileName == "" || fileName == "." {
			fileName = req.FileName // sent from bot with file info
		}
		if fileName == "" {
			fileName = "unknown_file"
		}

		// Actually insert the file into the database
		file := &models.File{
			GroupID:       groupID,
			UserID:        user.ID,
			SharedByPhone: req.UserPhone,
			SharedByName:  req.UserName,
			FileName:      fileName,
			FileSize:      req.FileSize,
			MimeType:      req.MimeType,
			Subject:       req.Subject,
			DriveFileID:   fmt.Sprintf("wa_%d_%d", user.ID, time.Now().UnixNano()),
		}

		saved, err := s.store.AddFile(file)
		if err != nil {
			log.Printf("Failed to save file: %v", err)
		}

		total := s.store.CountGroupFiles(groupID)
		version := 1
		if saved != nil {
			version = saved.Version
		}
		resp.Reply = s.reyna.AddResponse(fileName, version, total)
		s.store.LogActivity(groupID, user.ID, "add", req.Command, "stored")

	case "find":
		query := args
		if query == "" {
			resp.Reply = "Find kya? Query toh de bhai. Usage: /reyna find \"DSA notes\" 🤦‍♀️"
		} else {
			files, _ := s.store.FindFiles(groupID, query)
			if files == nil {
				files = []models.File{}
			}
			resp.Files = files
			resp.Reply = s.reyna.FindResponse(query, files)
		}
		s.store.LogActivity(groupID, user.ID, "find", req.Command, resp.Reply)

	case "log":
		files, _ := s.store.GetGroupFiles(groupID, 10)
		if files == nil {
			files = []models.File{}
		}
		total := s.store.CountGroupFiles(groupID)
		resp.Files = files
		resp.LogCount = total
		resp.Reply = s.reyna.LogResponse(files, total)
		s.store.LogActivity(groupID, user.ID, "log", req.Command, "shown")

	case "status":
		since := time.Now().Add(-24 * time.Hour) // last 24h
		newFiles, _ := s.store.GetNewFilesSince(groupID, since)
		if newFiles == nil {
			newFiles = []models.File{}
		}
		total := s.store.CountGroupFiles(groupID)
		resp.Files = newFiles
		resp.Reply = s.reyna.StatusResponse(newFiles, total)
		s.store.LogActivity(groupID, user.ID, "status", req.Command, "shown")

	case "help":
		resp.Reply = s.reyna.HelpResponse()
		s.store.LogActivity(groupID, user.ID, "help", req.Command, "shown")

	case "commit":
		// Commit staged files AND upload to Google Drive
		staged, _ := s.store.GetStagedFiles(groupID)
		if len(staged) == 0 {
			resp.Reply = s.reyna.CommitEmptyResponse()
		} else {
			// Find ANY user in this group who has Google Drive connected
			driveUser := s.store.FindDriveConnectedUser(groupID)
			uploaded := 0
			driveSkipped := 0
			for _, f := range staged {
				if args != "" && !strings.Contains(strings.ToLower(f.FileName), strings.ToLower(args)) {
					continue
				}
				if driveUser != nil && driveUser.GoogleRefresh != "" && s.drive.IsConfigured() {
					token, terr := s.drive.GetValidToken(driveUser.GoogleToken, driveUser.GoogleRefresh, 0)
					if terr != nil {
						log.Printf("Token refresh failed: %v", terr)
						driveSkipped++
						continue
					}
					if driveUser.DriveRootID == "" {
						log.Printf("No Drive root folder for user %d", driveUser.ID)
						driveSkipped++
						continue
					}
					folderID, ferr := s.drive.EnsureSubjectFolder(token, driveUser.DriveRootID, f.Subject)
					if ferr != nil {
						log.Printf("Folder error: %v", ferr)
						driveSkipped++
						continue
					}
					fileData, rerr := s.drive.GetLocalFileData(f.ID)
					if rerr != nil || len(fileData) == 0 {
						log.Printf("No local data for file %d: %v (len=%d)", f.ID, rerr, len(fileData))
						driveSkipped++
						continue
					}
					driveID, uerr := s.drive.UploadFileToDrive(token, folderID, f.FileName, f.MimeType, fileData)
					if uerr != nil {
						log.Printf("Drive upload failed: %v", uerr)
						driveSkipped++
						continue
					}
					if driveID != "" {
						s.store.UpdateFileDriveID(f.ID, driveID, folderID)
						uploaded++
						log.Printf("☁️ Uploaded to Drive: %s → %s", f.FileName, driveID)
					}
				} else {
					driveSkipped++
				}
			}
			count, _ := s.store.CommitFiles(groupID)
			if args != "" {
				s.store.CommitFileByName(groupID, args)
				resp.Reply = s.reyna.CommitFileResponse(args)
			} else {
				resp.Reply = s.reyna.CommitAllResponse(int(count))
			}
			if uploaded > 0 {
				resp.Reply += fmt.Sprintf("\n\n☁️ %d file(s) pushed to Google Drive!", uploaded)
			}
			if driveSkipped > 0 && uploaded == 0 {
				if driveUser == nil {
					resp.Reply += "\n\n⚠️ No one in this group has connected Google Drive yet. Files committed locally. Connect Drive on the web dashboard to sync."
				} else {
					resp.Reply += "\n\n⚠️ Drive upload failed. Check if files were properly staged with /reyna staged"
				}
			}
		}
		s.store.LogActivity(groupID, user.ID, "commit", req.Command, "committed")

	case "rm":
		if args == "" || args == "." {
			// Remove all staged
			count, _ := s.store.RemoveAllStaged(groupID)
			if count == 0 {
				resp.Reply = s.reyna.RmEmptyResponse()
			} else {
				resp.Reply = s.reyna.RmAllResponse(int(count))
			}
		} else {
			removed, _ := s.store.RemoveStagedFile(groupID, args)
			if removed {
				resp.Reply = s.reyna.RmFileResponse(args)
			} else {
				resp.Reply = s.reyna.RmNotFoundResponse(args)
			}
		}
		s.store.LogActivity(groupID, user.ID, "rm", req.Command, "removed")

	case "staged":
		staged, _ := s.store.GetStagedFiles(groupID)
		if staged == nil {
			staged = []models.File{}
		}
		resp.Files = staged
		resp.Reply = s.reyna.StagedResponse(staged)
		s.store.LogActivity(groupID, user.ID, "staged", req.Command, "shown")

	default:
		resp.Reply = s.reyna.GenericResponse()
	}

	json.NewEncoder(w).Encode(resp)
}

// ── Bot File Upload (receives binary data from WhatsApp bot) ──

func (s *Server) handleBotUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, 405)
		return
	}

	// Allow up to 50MB uploads
	r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024)

	var req struct {
		GroupWAID string `json:"group_wa_id"`
		UserPhone string `json:"user_phone"`
		UserName  string `json:"user_name"`
		FileName  string `json:"file_name"`
		FileSize  int64  `json:"file_size"`
		MimeType  string `json:"mime_type"`
		Subject   string `json:"subject"`
		FileData  string `json:"file_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Upload decode error: %v", err)
		http.Error(w, `{"error":"invalid body: `+err.Error()+`"}`, 400)
		return
	}

	log.Printf("📥 Received upload request: %s (base64 len: %d) from %s", req.FileName, len(req.FileData), req.UserName)

	// Decode base64
	fileBytes, err := decodeBase64(req.FileData)
	if err != nil {
		log.Printf("❌ Base64 decode error: %v (first 100 chars: %s)", err, req.FileData[:min64(100, len(req.FileData))])
		http.Error(w, `{"error":"base64 decode failed"}`, 400)
		return
	}

	if len(fileBytes) == 0 {
		log.Printf("❌ Decoded to 0 bytes (base64 len was %d)", len(req.FileData))
		http.Error(w, `{"error":"empty file data"}`, 400)
		return
	}

	log.Printf("📤 Bot upload: %s (%d bytes decoded from %d base64 chars) from %s",
		req.FileName, len(fileBytes), len(req.FileData), req.UserName)

	// Get or create user
	userName := req.UserName
	if userName == "" {
		userName = req.UserPhone
	}
	user, _ := s.store.UpsertUser(req.UserPhone, userName)

	// Get or create group
	group, err := s.store.GetGroupByWAID(req.GroupWAID)
	if err != nil {
		group, _ = s.store.UpsertGroup(req.GroupWAID, "WhatsApp Group", user.ID)
	}
	if group != nil {
		s.store.AddGroupMember(group.ID, user.ID, req.UserPhone, "member")
	}

	groupID := int64(0)
	if group != nil {
		groupID = group.ID
	}

	// Save file locally first
	localID, folderID, _ := s.drive.SmartUpload("", "", user.ID, req.Subject, req.FileName, req.MimeType, fileBytes)

	// Store in database as staged
	file := &models.File{
		GroupID:       groupID,
		UserID:        user.ID,
		SharedByPhone: req.UserPhone,
		SharedByName:  req.UserName,
		FileName:      req.FileName,
		FileSize:      int64(len(fileBytes)),
		MimeType:      req.MimeType,
		Subject:       req.Subject,
		DriveFileID:   localID,
		DriveFolderID: folderID,
		Status:        "staged",
	}

	saved, err := s.store.AddFile(file)
	if err != nil {
		log.Printf("DB save error: %v", err)
		http.Error(w, `{"error":"save failed"}`, 500)
		return
	}

	// Also save the raw bytes indexed by file ID for later Drive upload on commit
	s.drive.SaveLocalFileData(saved.ID, fileBytes)

	total := s.store.CountGroupFiles(groupID)
	version := 1
	if saved != nil {
		version = saved.Version
	}
	reply := s.reyna.AddResponse(req.FileName, version, total)

	s.store.LogActivity(groupID, user.ID, "add", "/reyna add "+req.FileName, "staged")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"reply":   reply,
		"file_id": saved.ID,
		"status":  "staged",
	})
}

func decodeBase64(data string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty base64 data")
	}
	// Try standard base64 first
	b, err := base64.StdEncoding.DecodeString(data)
	if err == nil && len(b) > 0 {
		return b, nil
	}
	// Try URL-safe base64
	b, err = base64.URLEncoding.DecodeString(data)
	if err == nil && len(b) > 0 {
		return b, nil
	}
	// Try raw (no padding) variants
	b, err = base64.RawStdEncoding.DecodeString(data)
	if err == nil && len(b) > 0 {
		return b, nil
	}
	b, err = base64.RawURLEncoding.DecodeString(data)
	if err == nil && len(b) > 0 {
		return b, nil
	}
	return nil, fmt.Errorf("all base64 decode attempts failed")
}

func min64(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Waitlist ──

func (s *Server) handleWaitlist(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var req models.WaitlistRequest
		json.NewDecoder(r.Body).Decode(&req)
		contact := strings.TrimSpace(req.Contact)
		if contact == "" {
			http.Error(w, `{"error":"contact required"}`, 400)
			return
		}
		s.store.AddWaitlist(contact)
		count := s.store.CountWaitlist()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "You're in! Reyna will ping you. (With attitude, obviously.) 💅",
			"position": count,
		})
	case "GET":
		count := s.store.CountWaitlist()
		json.NewEncoder(w).Encode(map[string]int{"count": count})
	default:
		http.Error(w, `{"error":"method not allowed"}`, 405)
	}
}

// ── Google OAuth ──

func (s *Server) handleGoogleAuthStart(w http.ResponseWriter, r *http.Request) {
	if !s.drive.IsConfigured() {
		json.NewEncoder(w).Encode(map[string]interface{}{"configured": false, "message": "Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET."})
		return
	}
	state := r.URL.Query().Get("token")
	if state == "" {
		http.Error(w, `{"error":"token param required"}`, 400)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"url": s.drive.GetAuthURL(state)})
}

func (s *Server) handleGoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		http.Error(w, "Missing code", 400)
		return
	}
	tokenInfo, email, err := s.drive.ExchangeCode(code)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><script>window.opener.postMessage({type:"google_auth_error",error:"%s"},"*");window.close();</script></body></html>`, err.Error())
		return
	}
	userID := int64(0)
	if state != "" {
		if uid, err := auth.ValidateToken(state, s.cfg.JWTSecret); err == nil {
			userID = uid
		}
	}
	if userID > 0 {
		rootID, _ := s.drive.CreateUserRootFolder(tokenInfo.AccessToken)
		s.store.UpdateUserGoogle(userID, email, tokenInfo.AccessToken, tokenInfo.RefreshToken, rootID)
		s.store.UpdateUserGoogleExpiry(userID, tokenInfo.ExpiresAt)
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html><body><script>window.opener.postMessage({type:"google_auth_success",email:"%s"},"*");window.close();</script><p>Connected! Close this window.</p></body></html>`, email)
}

func (s *Server) handleGoogleStatus(w http.ResponseWriter, r *http.Request) {
	uid := auth.GetUserID(r)
	user, _ := s.store.GetUserByID(uid)
	if user == nil {
		http.Error(w, `{"error":"user not found"}`, 404)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected": user.GoogleRefresh != "", "email": user.Email,
		"drive_root": user.DriveRootID, "configured": s.drive.IsConfigured(),
	})
}

func (s *Server) handleGoogleConnect(w http.ResponseWriter, r *http.Request) {
	if !s.drive.IsConfigured() {
		json.NewEncoder(w).Encode(map[string]interface{}{"configured": false})
		return
	}
	uid := auth.GetUserID(r)
	token, _ := auth.GenerateToken(uid, s.cfg.JWTSecret)
	json.NewEncoder(w).Encode(map[string]string{"url": s.drive.GetAuthURL(token)})
}