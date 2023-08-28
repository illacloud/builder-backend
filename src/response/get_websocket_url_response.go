package response

type WSURLResponse struct {
	WSURL string `json:"wsURL"`
}

func NewWSURLResponse(url string) *WSURLResponse {
	return &WSURLResponse{
		WSURL: url,
	}
}

func (resp *WSURLResponse) ExportForFeedback() interface{} {
	return resp
}
