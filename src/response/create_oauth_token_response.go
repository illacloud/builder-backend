package response

type CreateOAuthTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

func NewCreateOAuthTokenResponse(token string) *CreateOAuthTokenResponse {
	return &CreateOAuthTokenResponse{
		AccessToken: token,
	}
}

func (resp *CreateOAuthTokenResponse) ExportForFeedback() interface{} {
	return resp
}
