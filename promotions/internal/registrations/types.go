package registrations

type RegisterPromotionEvent struct {
	Email      string   `json:"email"`
	ClientID   int64    `json:"clientId"`
	Categories []string `json:"categories"`
}

type UnregisterPromotionEvent struct {
	Email    string `json:"email"`
	ClientID int64  `json:"clientId"`
}
