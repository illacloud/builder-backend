package request

type CreateOAuthTokenRequest struct {
	RedirectURL string `json:"redirectURL" validate:"required"`
	AccessType  string `json:"accessType" validate:"oneof=rw r"`
}

const (
	OAUTH_TOKEN_ACCESS_TYPE_READ_AND_WRITE = "rw"
	OAUTH_TOKEN_ACCESS_TYPE_READ_ONLY      = "r"
)

func NewCreateOAuthTokenRequest() *CreateOAuthTokenRequest {
	return &CreateOAuthTokenRequest{}
}

func (req *CreateOAuthTokenRequest) IsReadAndWrite() bool {
	return req.AccessType == OAUTH_TOKEN_ACCESS_TYPE_READ_AND_WRITE
}

func (req *CreateOAuthTokenRequest) IsReadOnly() bool {
	return req.AccessType == OAUTH_TOKEN_ACCESS_TYPE_READ_ONLY
}

func (req *CreateOAuthTokenRequest) ExportRedirectURL() string {
	return req.RedirectURL
}
