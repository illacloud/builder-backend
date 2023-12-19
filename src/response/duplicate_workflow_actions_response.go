package response

type DuplicateWorkflowActionsResponse struct {
	IDMap map[int]int `json:"idMap"`
}

func NewDuplicateWorkflowActionsResponse(idMap map[int]int) *DuplicateWorkflowActionsResponse {
	return &DuplicateWorkflowActionsResponse{
		IDMap: idMap,
	}
}

func (resp *DuplicateWorkflowActionsResponse) ExportForFeedback() interface{} {
	return resp
}
