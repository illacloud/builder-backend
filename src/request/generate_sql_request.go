package request

import (
	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

const GENERATE_SQL_ACTION_SELECT = 1
const GENERATE_SQL_ACTION_INSERT = 2
const GENERATE_SQL_ACTION_UPDATE = 3
const GENERATE_SQL_ACTION_DELETE = 4

var ACTION_MAP = map[int]string{
	GENERATE_SQL_ACTION_SELECT: "SELECT",
	GENERATE_SQL_ACTION_INSERT: "INSERT",
	GENERATE_SQL_ACTION_UPDATE: "UPDATE",
	GENERATE_SQL_ACTION_DELETE: "DELETE",
}

type GenerateSQLRequest struct {
	Description string `json:"description" validate:"required"`
	ResourceID  string `json:"resourceID" validate:"required"`
	Action      int    `json:"action" validate:"required"`
}

func NewGenerateSQLRequest() *GenerateSQLRequest {
	return &GenerateSQLRequest{}
}

func (req *GenerateSQLRequest) GetActionInString() string {
	return ACTION_MAP[req.Action]
}

func (req *GenerateSQLRequest) ExportResourceIDInInt() int {
	return idconvertor.ConvertStringToInt(req.ResourceID)
}
