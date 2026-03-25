package models

import "time"

// User represents a registered Reyna user
type User struct {
	ID             int64     `json:"id"`
	Phone          string    `json:"phone"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	GoogleToken    string    `json:"-"`
	GoogleRefresh  string    `json:"-"`
	DriveRootID    string    `json:"drive_root_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Group represents a WhatsApp group connected to Reyna
type Group struct {
	ID          int64     `json:"id"`
	WAID        string    `json:"wa_id"`         // WhatsApp group JID
	Name        string    `json:"name"`
	MemberCount int       `json:"member_count"`
	CreatedBy   int64     `json:"created_by"`    // user ID
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupMember links users to groups
type GroupMember struct {
	ID      int64     `json:"id"`
	GroupID int64     `json:"group_id"`
	UserID  int64     `json:"user_id"`
	Phone   string    `json:"phone"`
	Role    string    `json:"role"` // "admin" | "member"
	JoinedAt time.Time `json:"joined_at"`
}

// File represents a stored file in a user's Drive repo
type File struct {
	ID            int64     `json:"id"`
	GroupID       int64     `json:"group_id"`
	UserID        int64     `json:"user_id"`
	SharedByPhone string    `json:"shared_by_phone"`
	SharedByName  string    `json:"shared_by_name"`
	FileName      string    `json:"file_name"`
	FileSize      int64     `json:"file_size"`
	MimeType      string    `json:"mime_type"`
	DriveFileID   string    `json:"drive_file_id"`
	DriveFolderID string    `json:"drive_folder_id"`
	Subject       string    `json:"subject"`
	Tags          string    `json:"tags"`
	Version       int       `json:"version"`
	ParentFileID  int64     `json:"parent_file_id"`
	WAMessageID   string    `json:"wa_message_id"`
	Status        string    `json:"status"` // "staged" | "committed"
	CreatedAt     time.Time `json:"created_at"`
}

// FileVersion tracks version history
type FileVersion struct {
	ID          int64     `json:"id"`
	FileID      int64     `json:"file_id"`
	Version     int       `json:"version"`
	DriveFileID string    `json:"drive_file_id"`
	FileSize    int64     `json:"file_size"`
	ChangedBy   int64     `json:"changed_by"`
	ChangeNote  string    `json:"change_note"`
	CreatedAt   time.Time `json:"created_at"`
}

// ActivityLog tracks all bot interactions
type ActivityLog struct {
	ID        int64     `json:"id"`
	GroupID   int64     `json:"group_id"`
	UserID    int64     `json:"user_id"`
	Action    string    `json:"action"` // "add", "find", "log", "status"
	Command   string    `json:"command"`
	Result    string    `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

// WaitlistEntry for the landing page
type WaitlistEntry struct {
	ID        int64     `json:"id"`
	Contact   string    `json:"contact"` // phone or email
	CreatedAt time.Time `json:"created_at"`
}

// --- API Request/Response types ---

type LoginRequest struct {
	Phone string `json:"phone"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type AddFileRequest struct {
	GroupWAID     string `json:"group_wa_id"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	MimeType      string `json:"mime_type"`
	FileData      string `json:"file_data"` // base64
	SharedByPhone string `json:"shared_by_phone"`
	SharedByName  string `json:"shared_by_name"`
	Subject       string `json:"subject"`
	WAMessageID   string `json:"wa_message_id"`
}

type FindRequest struct {
	GroupWAID string `json:"group_wa_id"`
	Query     string `json:"query"`
}

type CommandRequest struct {
	GroupWAID     string `json:"group_wa_id"`
	Command       string `json:"command"`
	Args          string `json:"args"`
	UserPhone     string `json:"user_phone"`
	UserName      string `json:"user_name"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	MimeType      string `json:"mime_type"`
	Subject       string `json:"subject"`
}

type CommandResponse struct {
	Reply    string  `json:"reply"`
	Files    []File  `json:"files,omitempty"`
	LogCount int     `json:"log_count,omitempty"`
}

type DashboardStats struct {
	TotalFiles     int            `json:"total_files"`
	TotalGroups    int            `json:"total_groups"`
	TotalSize      int64          `json:"total_size"`
	RecentFiles    []File         `json:"recent_files"`
	SubjectBreak   map[string]int `json:"subject_breakdown"`
	TopContributors []Contributor `json:"top_contributors"`
}

type Contributor struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Count int    `json:"count"`
}

type WaitlistRequest struct {
	Contact string `json:"contact"`
}
