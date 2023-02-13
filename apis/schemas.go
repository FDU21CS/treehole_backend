package apis

import "time"

/* account */

type EmailModel struct {
	Email string `json:"email" query:"email" validate:"required,email"`
}

type LoginRequest struct {
	EmailModel
	Password string `json:"password" minLength:"8" validate:"required,min=8"`
}

type TokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
	Message string `json:"message"`
}

type RegisterRequest struct {
	LoginRequest
	Verification string `json:"verification" minLength:"6" maxLength:"6" validate:"len=6"`
}

type VerifyResponse struct {
	Message string `json:"message"`
	Scope   string `json:"scope" enums:"register,reset"`
}

/* Hole */

type ListHoleRequest struct {
	StartTime time.Time `json:"start_time" query:"start_time"`
	Size      int       `json:"size" query:"size"`
	Order     string    `json:"order" query:"order" validate:"omitempty,oneof=created_at updated_at" default:"created_at"`
}

type CreateHoleRequest struct {
	Content    string `json:"content" validate:"max=15000"`
	SpecialTag string `json:"special_tag"`
}

/* Floor */

type ListFloorRequest struct {
	StartFloor int `json:"start_floor" query:"start_floor"`
	Size       int `json:"size" query:"size"`
}

type CreateFloorRequest struct {
	Content    string `json:"content" validate:"max=15000"`
	SpecialTag string `json:"special_tag"`
}

type ModifyFloorRequest struct {
	Content    *string `json:"content" validate:"max=15000"`
	SpecialTag string  `json:"special_tag"`
}

type DeleteFloorRequest struct {
	Reason string `json:"reason" validate:"required"`
}
