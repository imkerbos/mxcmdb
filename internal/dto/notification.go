package dto

// NotificationResponse 通知响应
type NotificationResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	SourceID  uint   `json:"source_id"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

// NotificationUnreadCount 未读数响应
type NotificationUnreadCount struct {
	Count int64 `json:"count"`
}
