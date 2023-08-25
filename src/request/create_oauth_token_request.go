package request

type CreateOAuthTokenRequest struct {
	RedirectURL string `json:"redirectURL" validate:"required"`
	AccessType  string `json:"accessType" validate:"oneof=rw r"`
}

func NewCreateOAuthTokenRequest() *CreateOAuthTokenRequest {
	return &CreateOAuthTokenRequest{}
}
