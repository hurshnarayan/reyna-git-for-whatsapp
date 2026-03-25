package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/reyna-bot/reyna-backend/internal/models"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	database, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	database.SetMaxOpenConns(5)
	s := &Store{db: database}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		phone TEXT UNIQUE NOT NULL,
		name TEXT DEFAULT '',
		email TEXT DEFAULT '',
		google_token TEXT DEFAULT '',
		google_refresh TEXT DEFAULT '',
		drive_root_id TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS groups_ (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		wa_id TEXT UNIQUE NOT NULL,
		name TEXT DEFAULT '',
		member_count INTEGER DEFAULT 0,
		created_by INTEGER REFERENCES users(id),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS group_members (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER REFERENCES groups_(id),
		user_id INTEGER REFERENCES users(id),
		phone TEXT NOT NULL,
		role TEXT DEFAULT 'member',
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(group_id, phone)
	);
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER REFERENCES groups_(id),
		user_id INTEGER REFERENCES users(id),
		shared_by_phone TEXT DEFAULT '',
		shared_by_name TEXT DEFAULT '',
		file_name TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		mime_type TEXT DEFAULT '',
		drive_file_id TEXT DEFAULT '',
		drive_folder_id TEXT DEFAULT '',
		subject TEXT DEFAULT '',
		tags TEXT DEFAULT '',
		version INTEGER DEFAULT 1,
		parent_file_id INTEGER DEFAULT 0,
		wa_message_id TEXT DEFAULT '',
		status TEXT DEFAULT 'staged',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS file_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER REFERENCES files(id),
		version INTEGER NOT NULL,
		drive_file_id TEXT DEFAULT '',
		file_size INTEGER DEFAULT 0,
		changed_by INTEGER REFERENCES users(id),
		change_note TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS activity_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER DEFAULT 0,
		user_id INTEGER DEFAULT 0,
		action TEXT NOT NULL,
		command TEXT DEFAULT '',
		result TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS waitlist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		contact TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_files_group ON files(group_id);
	CREATE INDEX IF NOT EXISTS idx_files_user ON files(user_id);
	CREATE INDEX IF NOT EXISTS idx_files_name ON files(file_name);
	CREATE INDEX IF NOT EXISTS idx_files_subject ON files(subject);
	CREATE INDEX IF NOT EXISTS idx_activity_group ON activity_log(group_id);
	`
	_, err := s.db.Exec(schema)
	return err
}

// ── User Operations ──

func (s *Store) UpsertUser(phone, name string) (*models.User, error) {
	_, err := s.db.Exec(
		`INSERT INTO users (phone, name, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(phone) DO UPDATE SET name=COALESCE(NULLIF(excluded.name,''), name), updated_at=CURRENT_TIMESTAMP`,
		phone, name,
	)
	if err != nil {
		return nil, err
	}
	return s.GetUserByPhone(phone)
}

func (s *Store) GetUserByPhone(phone string) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow(
		`SELECT id, phone, name, email, google_token, google_refresh, drive_root_id, created_at, updated_at FROM users WHERE phone=?`,
		phone,
	).Scan(&u.ID, &u.Phone, &u.Name, &u.Email, &u.GoogleToken, &u.GoogleRefresh, &u.DriveRootID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByID(id int64) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow(
		`SELECT id, phone, name, email, google_token, google_refresh, drive_root_id, created_at, updated_at FROM users WHERE id=?`,
		id,
	).Scan(&u.ID, &u.Phone, &u.Name, &u.Email, &u.GoogleToken, &u.GoogleRefresh, &u.DriveRootID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) UpdateUserGoogle(userID int64, email, token, refresh, rootID string) error {
	_, err := s.db.Exec(
		`UPDATE users SET email=?, google_token=?, google_refresh=?, drive_root_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		email, token, refresh, rootID, userID,
	)
	return err
}

func (s *Store) UpdateUserGoogleExpiry(userID int64, expiresAt int64) error {
	_, err := s.db.Exec(`UPDATE users SET updated_at=CURRENT_TIMESTAMP WHERE id=?`, userID)
	return err
}

// ── Group Operations ──

func (s *Store) UpsertGroup(waID, name string, createdBy int64) (*models.Group, error) {
	_, err := s.db.Exec(
		`INSERT INTO groups_ (wa_id, name, created_by, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(wa_id) DO UPDATE SET name=excluded.name, updated_at=CURRENT_TIMESTAMP`,
		waID, name, createdBy,
	)
	if err != nil {
		return nil, err
	}
	return s.GetGroupByWAID(waID)
}

func (s *Store) GetGroupByWAID(waID string) (*models.Group, error) {
	g := &models.Group{}
	err := s.db.QueryRow(
		`SELECT id, wa_id, name, member_count, created_by, created_at, updated_at FROM groups_ WHERE wa_id=?`,
		waID,
	).Scan(&g.ID, &g.WAID, &g.Name, &g.MemberCount, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt)
	return g, err
}

func (s *Store) GetUserGroups(userID int64) ([]models.Group, error) {
	rows, err := s.db.Query(
		`SELECT g.id, g.wa_id, g.name, g.member_count, g.created_by, g.created_at, g.updated_at
		 FROM groups_ g JOIN group_members gm ON g.id = gm.group_id WHERE gm.user_id=? ORDER BY g.updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []models.Group
	for rows.Next() {
		var g models.Group
		rows.Scan(&g.ID, &g.WAID, &g.Name, &g.MemberCount, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt)
		groups = append(groups, g)
	}
	return groups, nil
}

func (s *Store) AddGroupMember(groupID, userID int64, phone, role string) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO group_members (group_id, user_id, phone, role) VALUES (?, ?, ?, ?)`,
		groupID, userID, phone, role,
	)
	return err
}

// AutoLinkUserToGroups finds all groups where files were shared by this phone
// and adds the user as a member. This syncs web login with WhatsApp bot activity.
func (s *Store) AutoLinkUserToGroups(userID int64, phone string) {
	var groupIDs []int64
	seen := make(map[int64]bool)

	// 1. Groups where this phone shared files
	rows, err := s.db.Query(`SELECT DISTINCT group_id FROM files WHERE shared_by_phone=?`, phone)
	if err == nil {
		for rows.Next() {
			var gid int64
			rows.Scan(&gid)
			if !seen[gid] {
				groupIDs = append(groupIDs, gid)
				seen[gid] = true
			}
		}
		rows.Close()
	}

	// 2. Groups where this phone is already a member (from another user record)
	rows2, err := s.db.Query(`SELECT DISTINCT group_id FROM group_members WHERE phone=?`, phone)
	if err == nil {
		for rows2.Next() {
			var gid int64
			rows2.Scan(&gid)
			if !seen[gid] {
				groupIDs = append(groupIDs, gid)
				seen[gid] = true
			}
		}
		rows2.Close()
	}

	// 3. All existing groups (for hackathon demo — every user sees all data)
	rows3, err := s.db.Query(`SELECT id FROM groups_`)
	if err == nil {
		for rows3.Next() {
			var gid int64
			rows3.Scan(&gid)
			if !seen[gid] {
				groupIDs = append(groupIDs, gid)
				seen[gid] = true
			}
		}
		rows3.Close()
	}

	// Now insert memberships (all rows closed, no lock contention)
	for _, gid := range groupIDs {
		s.AddGroupMember(gid, userID, phone, "member")
	}
}

// ── File Operations ──

func (s *Store) AddFile(f *models.File) (*models.File, error) {
	var existingID int64
	var existingVersion int
	err := s.db.QueryRow(
		`SELECT id, version FROM files WHERE group_id=? AND file_name=? ORDER BY version DESC LIMIT 1`,
		f.GroupID, f.FileName,
	).Scan(&existingID, &existingVersion)

	if err == nil {
		f.Version = existingVersion + 1
		f.ParentFileID = existingID
	} else {
		f.Version = 1
	}

	if f.Status == "" {
		f.Status = "staged"
	}

	res, err := s.db.Exec(
		`INSERT INTO files (group_id, user_id, shared_by_phone, shared_by_name, file_name, file_size,
		  mime_type, drive_file_id, drive_folder_id, subject, tags, version, parent_file_id, wa_message_id, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		f.GroupID, f.UserID, f.SharedByPhone, f.SharedByName, f.FileName, f.FileSize,
		f.MimeType, f.DriveFileID, f.DriveFolderID, f.Subject, f.Tags, f.Version, f.ParentFileID, f.WAMessageID, f.Status,
	)
	if err != nil {
		return nil, err
	}
	f.ID, _ = res.LastInsertId()
	f.CreatedAt = time.Now()

	s.db.Exec(
		`INSERT INTO file_versions (file_id, version, drive_file_id, file_size, changed_by) VALUES (?, ?, ?, ?, ?)`,
		f.ID, f.Version, f.DriveFileID, f.FileSize, f.UserID,
	)

	return f, nil
}

// CommitFiles marks staged files as committed in a group
func (s *Store) CommitFiles(groupID int64) (int64, error) {
	res, err := s.db.Exec(
		`UPDATE files SET status='committed' WHERE group_id=? AND status='staged'`,
		groupID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// CommitFileByName marks a specific staged file as committed
func (s *Store) CommitFileByName(groupID int64, fileName string) error {
	_, err := s.db.Exec(
		`UPDATE files SET status='committed' WHERE group_id=? AND file_name LIKE ? AND status='staged'`,
		groupID, "%"+fileName+"%",
	)
	return err
}

// RemoveStagedFile removes a staged (uncommitted) file
func (s *Store) RemoveStagedFile(groupID int64, fileName string) (bool, error) {
	res, err := s.db.Exec(
		`DELETE FROM files WHERE group_id=? AND file_name LIKE ? AND status='staged'`,
		groupID, "%"+fileName+"%",
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// RemoveAllStaged removes all staged files in a group
func (s *Store) RemoveAllStaged(groupID int64) (int64, error) {
	res, err := s.db.Exec(
		`DELETE FROM files WHERE group_id=? AND status='staged'`,
		groupID,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// GetStagedFiles returns all staged (uncommitted) files in a group
func (s *Store) GetStagedFiles(groupID int64) ([]models.File, error) {
	rows, err := s.db.Query(
		`SELECT id, group_id, user_id, shared_by_phone, shared_by_name, file_name, file_size,
		  mime_type, drive_file_id, drive_folder_id, subject, tags, version, parent_file_id, wa_message_id, status, created_at
		 FROM files WHERE group_id=? AND status='staged' ORDER BY created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFiles(rows)
}

// CountStagedFiles returns count of staged files in a group
func (s *Store) CountStagedFiles(groupID int64) int {
	var c int
	s.db.QueryRow(`SELECT COUNT(*) FROM files WHERE group_id=? AND status='staged'`, groupID).Scan(&c)
	return c
}

func (s *Store) FindFiles(groupID int64, query string) ([]models.File, error) {
	rows, err := s.db.Query(
		`SELECT id, group_id, user_id, shared_by_phone, shared_by_name, file_name, file_size,
		  mime_type, drive_file_id, drive_folder_id, subject, tags, version, parent_file_id, wa_message_id, status, created_at
		 FROM files WHERE group_id=? AND (file_name LIKE ? OR subject LIKE ? OR tags LIKE ?)
		 ORDER BY created_at DESC LIMIT 20`,
		groupID, "%"+query+"%", "%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFiles(rows)
}

func (s *Store) GetGroupFiles(groupID int64, limit int) ([]models.File, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT id, group_id, user_id, shared_by_phone, shared_by_name, file_name, file_size,
		  mime_type, drive_file_id, drive_folder_id, subject, tags, version, parent_file_id, wa_message_id, status, created_at
		 FROM files WHERE group_id=? ORDER BY created_at DESC LIMIT ?`,
		groupID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFiles(rows)
}

func (s *Store) GetUserFiles(userID int64, limit int) ([]models.File, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT id, group_id, user_id, shared_by_phone, shared_by_name, file_name, file_size,
		  mime_type, drive_file_id, drive_folder_id, subject, tags, version, parent_file_id, wa_message_id, status, created_at
		 FROM files WHERE user_id=? ORDER BY created_at DESC LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFiles(rows)
}

func (s *Store) GetFileVersions(fileID int64) ([]models.FileVersion, error) {
	// Get the original file name, find all versions
	var fileName string
	var groupID int64
	s.db.QueryRow(`SELECT file_name, group_id FROM files WHERE id=?`, fileID).Scan(&fileName, &groupID)

	rows, err := s.db.Query(
		`SELECT f.id, f.version, f.drive_file_id, f.file_size, f.user_id, f.created_at
		 FROM files f WHERE f.group_id=? AND f.file_name=? ORDER BY f.version DESC`,
		groupID, fileName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var versions []models.FileVersion
	for rows.Next() {
		var v models.FileVersion
		rows.Scan(&v.ID, &v.Version, &v.DriveFileID, &v.FileSize, &v.ChangedBy, &v.CreatedAt)
		v.FileID = fileID
		versions = append(versions, v)
	}
	return versions, nil
}

func (s *Store) CountGroupFiles(groupID int64) int {
	var c int
	s.db.QueryRow(`SELECT COUNT(*) FROM files WHERE group_id=?`, groupID).Scan(&c)
	return c
}

func (s *Store) CountUserFiles(userID int64) int {
	var c int
	s.db.QueryRow(`SELECT COUNT(*) FROM files WHERE user_id=?`, userID).Scan(&c)
	return c
}

func (s *Store) GetNewFilesSince(groupID int64, since time.Time) ([]models.File, error) {
	rows, err := s.db.Query(
		`SELECT id, group_id, user_id, shared_by_phone, shared_by_name, file_name, file_size,
		  mime_type, drive_file_id, drive_folder_id, subject, tags, version, parent_file_id, wa_message_id, status, created_at
		 FROM files WHERE group_id=? AND created_at > ? ORDER BY created_at DESC`,
		groupID, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFiles(rows)
}

// ── Dashboard Stats ──

func (s *Store) GetDashboardStats(userID int64) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{
		SubjectBreak: make(map[string]int),
	}

	// Get all group IDs the user is a member of
	groupIDs := s.GetUserGroupIDs(userID)

	if len(groupIDs) == 0 {
		// No groups — show files user personally stored as fallback
		s.db.QueryRow(`SELECT COUNT(*) FROM files WHERE user_id=?`, userID).Scan(&stats.TotalFiles)
		s.db.QueryRow(`SELECT COALESCE(SUM(file_size),0) FROM files WHERE user_id=?`, userID).Scan(&stats.TotalSize)
		stats.RecentFiles, _ = s.GetUserFiles(userID, 10)
		return stats, nil
	}

	// Build IN clause
	placeholders, args := buildInClause(groupIDs)

	s.db.QueryRow(`SELECT COUNT(*) FROM files WHERE group_id IN (`+placeholders+`)`, args...).Scan(&stats.TotalFiles)
	s.db.QueryRow(`SELECT COUNT(DISTINCT group_id) FROM files WHERE group_id IN (`+placeholders+`)`, args...).Scan(&stats.TotalGroups)
	s.db.QueryRow(`SELECT COALESCE(SUM(file_size),0) FROM files WHERE group_id IN (`+placeholders+`)`, args...).Scan(&stats.TotalSize)

	// Recent files from all groups
	stats.RecentFiles, _ = s.GetGroupsFiles(groupIDs, 10)

	// Subject breakdown across all groups
	rows, _ := s.db.Query(
		`SELECT COALESCE(NULLIF(subject,''),'Uncategorized'), COUNT(*) FROM files WHERE group_id IN (`+placeholders+`) GROUP BY subject`,
		args...,
	)
	if rows != nil {
		for rows.Next() {
			var sub string
			var cnt int
			rows.Scan(&sub, &cnt)
			stats.SubjectBreak[sub] = cnt
		}
		rows.Close()
	}

	// Top contributors across all groups
	contribRows, _ := s.db.Query(
		`SELECT shared_by_name, shared_by_phone, COUNT(*) as cnt FROM files
		 WHERE group_id IN (`+placeholders+`) AND shared_by_phone != ''
		 GROUP BY shared_by_phone ORDER BY cnt DESC LIMIT 5`,
		args...,
	)
	if contribRows != nil {
		for contribRows.Next() {
			var c models.Contributor
			contribRows.Scan(&c.Name, &c.Phone, &c.Count)
			stats.TopContributors = append(stats.TopContributors, c)
		}
		contribRows.Close()
	}

	return stats, nil
}

// GetUserGroupIDs returns all group IDs a user belongs to
func (s *Store) GetUserGroupIDs(userID int64) []int64 {
	var ids []int64
	seen := make(map[int64]bool)

	rows, err := s.db.Query(`SELECT group_id FROM group_members WHERE user_id=?`, userID)
	if err == nil {
		for rows.Next() {
			var id int64
			rows.Scan(&id)
			if !seen[id] {
				ids = append(ids, id)
				seen[id] = true
			}
		}
		rows.Close()
	}

	// Also include groups the user created
	rows2, err := s.db.Query(`SELECT id FROM groups_ WHERE created_by=?`, userID)
	if err == nil {
		for rows2.Next() {
			var id int64
			rows2.Scan(&id)
			if !seen[id] {
				ids = append(ids, id)
				seen[id] = true
			}
		}
		rows2.Close()
	}

	return ids
}

// GetGroupsFiles returns files from multiple groups
func (s *Store) GetGroupsFiles(groupIDs []int64, limit int) ([]models.File, error) {
	if len(groupIDs) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	placeholders, args := buildInClause(groupIDs)
	args = append(args, interface{}(limit))
	rows, err := s.db.Query(
		`SELECT id, group_id, user_id, shared_by_phone, shared_by_name, file_name, file_size,
		  mime_type, drive_file_id, drive_folder_id, subject, tags, version, parent_file_id, wa_message_id, status, created_at
		 FROM files WHERE group_id IN (`+placeholders+`) ORDER BY created_at DESC LIMIT ?`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFiles(rows)
}

func buildInClause(ids []int64) (string, []interface{}) {
	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}
	return placeholders, args
}

// ── Activity Log ──

func (s *Store) LogActivity(groupID, userID int64, action, command, result string) {
	s.db.Exec(
		`INSERT INTO activity_log (group_id, user_id, action, command, result) VALUES (?, ?, ?, ?, ?)`,
		groupID, userID, action, command, result,
	)
}

func (s *Store) GetActivityLog(groupID int64, limit int) ([]models.ActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT id, group_id, user_id, action, command, result, created_at
		 FROM activity_log WHERE group_id=? ORDER BY created_at DESC LIMIT ?`,
		groupID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []models.ActivityLog
	for rows.Next() {
		var l models.ActivityLog
		rows.Scan(&l.ID, &l.GroupID, &l.UserID, &l.Action, &l.Command, &l.Result, &l.CreatedAt)
		logs = append(logs, l)
	}
	return logs, nil
}

// ── Waitlist ──

func (s *Store) AddWaitlist(contact string) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO waitlist (contact) VALUES (?)`, contact)
	return err
}

func (s *Store) CountWaitlist() int {
	var c int
	s.db.QueryRow(`SELECT COUNT(*) FROM waitlist`).Scan(&c)
	return c
}

// ── Helpers ──

func scanFiles(rows *sql.Rows) ([]models.File, error) {
	var files []models.File
	for rows.Next() {
		var f models.File
		err := rows.Scan(&f.ID, &f.GroupID, &f.UserID, &f.SharedByPhone, &f.SharedByName,
			&f.FileName, &f.FileSize, &f.MimeType, &f.DriveFileID, &f.DriveFolderID,
			&f.Subject, &f.Tags, &f.Version, &f.ParentFileID, &f.WAMessageID, &f.Status, &f.CreatedAt)
		if err != nil {
			continue
		}
		files = append(files, f)
	}
	return files, nil
}

// UpdateFileDriveID updates the drive file ID after uploading to Google Drive
func (s *Store) UpdateFileDriveID(fileID int64, driveFileID, driveFolderID string) error {
	_, err := s.db.Exec(
		`UPDATE files SET drive_file_id=?, drive_folder_id=? WHERE id=?`,
		driveFileID, driveFolderID, fileID,
	)
	return err
}

// FindDriveConnectedUser finds any user in a group who has Google Drive connected
func (s *Store) FindDriveConnectedUser(groupID int64) *models.User {
	u := &models.User{}
	err := s.db.QueryRow(
		`SELECT u.id, u.phone, u.name, u.email, u.google_token, u.google_refresh, u.drive_root_id, u.created_at, u.updated_at
		 FROM users u JOIN group_members gm ON u.id = gm.user_id
		 WHERE gm.group_id=? AND u.google_refresh != '' AND u.drive_root_id != ''
		 LIMIT 1`,
		groupID,
	).Scan(&u.ID, &u.Phone, &u.Name, &u.Email, &u.GoogleToken, &u.GoogleRefresh, &u.DriveRootID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		// Fallback: try ANY user with Drive connected (for hackathon demo)
		err = s.db.QueryRow(
			`SELECT id, phone, name, email, google_token, google_refresh, drive_root_id, created_at, updated_at
			 FROM users WHERE google_refresh != '' AND drive_root_id != '' LIMIT 1`,
		).Scan(&u.ID, &u.Phone, &u.Name, &u.Email, &u.GoogleToken, &u.GoogleRefresh, &u.DriveRootID, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil
		}
	}
	return u
}
