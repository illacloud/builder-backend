package illacapacity

import (
	"strconv"

	"github.com/google/uuid"
)

type UpdateCapacityRequest struct {
	TeamID         int       `json:"teamID"         validate:"required"`
	TransactionUID uuid.UUID `json:"transactionUID" validate:"required"`
	ModifiedTarget int       `json:"modifiedTarget" validate:"required"`
	ModifiedMethod int       `json:"modifiedMethod" validate:"required"`
	ModifiedValue  int64     `json:"modifiedValue"  validate:"required"`
	InstanceType   int       `json:"instanceType"   validate:"required"`
	InstanceID     int       `json:"instanceID"     validate:"required"`
}

func NewUpdateCapacityRequest() *UpdateCapacityRequest {
	return &UpdateCapacityRequest{}
}

func NewUpdateCapacityRequestByParam(teamID int, transactionUID string, modifiedTarget int, modifiedMethod int, modifiedValue int64, instanceType int, instanceID int) *UpdateCapacityRequest {
	txuid, _ := uuid.Parse(transactionUID)
	return &UpdateCapacityRequest{
		TeamID:         teamID,
		TransactionUID: txuid,
		ModifiedTarget: modifiedTarget,
		ModifiedMethod: modifiedMethod,
		ModifiedValue:  modifiedValue,
		InstanceType:   instanceType,
		InstanceID:     instanceID,
	}
}

func (u *UpdateCapacityRequest) ExportForRequestTokenValidator() []string {
	return []string{
		strconv.Itoa(u.TeamID),
		u.TransactionUID.String(),
		strconv.Itoa(u.ModifiedTarget),
		strconv.Itoa(u.ModifiedMethod),
		strconv.FormatInt(u.ModifiedValue, 10),
		strconv.Itoa(u.InstanceType),
		strconv.Itoa(u.InstanceID),
	}
}

func (u *UpdateCapacityRequest) ExportModifiedTarget() int {
	return u.ModifiedTarget
}

func (u *UpdateCapacityRequest) ExportModifiedMethod() int {
	return u.ModifiedMethod
}
