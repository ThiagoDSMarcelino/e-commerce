package promotions

import "time"

type listRegistredEmailQuery struct {
	Size     int32 `form:"size,default=10" binding:"min=1,max=100"`
	Page     int32 `form:"page,default=1" binding:"min=1"`
	ClientID int64 `form:"clientId" binding:"required,gt=0"`
}

type registerSubscriptionRequest struct {
	Email      string   `json:"email" binding:"required,email"`
	ClientID   int64    `json:"clientId" binding:"required,gt=0"`
	Categories []string `json:"categories" binding:"required,min=1,dive,required"`
}

type cancelSubscriptionRequest struct {
	Email    string `json:"email" binding:"required,email"`
	ClientID int64  `json:"clientId" binding:"required,gt=0"`
}

type registrationResponse struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type registrationListResponse struct {
	Items []registrationResponse `json:"items"`
	Total int64                  `json:"total"`
	Page  int32                  `json:"page"`
}
