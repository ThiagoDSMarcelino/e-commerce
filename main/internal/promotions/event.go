package promotions

type EventAction string

type registerPromotionEvent struct {
	Email      string   `json:"email"`
	ClientID   int64    `json:"clientId"`
	Categories []string `json:"categories"`
}

type unregisterPromotionEvent struct {
	Email    string `json:"email"`
	ClientID int64  `json:"clientId"`
}
