package dto

// BackupRequest 创建备份请求
type BackupRequest struct {
	Description string   `json:"description"`
	Collections []string `json:"collections"` // 为空则备份全部
}

// BackupResponse 备份记录响应
type BackupResponse struct {
	ID          string `json:"id"`
	FilePath    string `json:"file_path"`
	FileSize    int64  `json:"file_size"`
	Description string `json:"description"`
	Status      string `json:"status"` // completed / failed
	CreatedAt   string `json:"created_at"`
	ErrorMsg    string `json:"error_msg,omitempty"`
}

// RestoreRequest 恢复请求
type RestoreRequest struct {
	BackupID string `json:"backup_id" binding:"required"`
}

// RestoreResponse 恢复结果
type RestoreResponse struct {
	Success      bool     `json:"success"`
	BackupID     string   `json:"backup_id"`
	RestoredCols []string `json:"restored_collections"`
	Note         string   `json:"note"`
}
