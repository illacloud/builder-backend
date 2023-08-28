package illacloudperipheralapisdk

type GenerateSQLFeedback struct {
	Payload string `json:"payload"`
}

func (resp *GenerateSQLFeedback) ExportForFeedback() interface{} {
	return resp
}
