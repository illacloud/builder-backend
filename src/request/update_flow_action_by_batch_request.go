package request

type UpdateFlowActionByBatchRequest struct {
	FlowActions []*UpdateFlowActionRequest `json:"flowActions"`
}

func NewUpdateFlowActionByBatchRequest() *UpdateFlowActionByBatchRequest {
	return &UpdateFlowActionByBatchRequest{}
}

func (req *UpdateFlowActionByBatchRequest) ExportFlowActions() []*UpdateFlowActionRequest {
	return req.FlowActions
}
