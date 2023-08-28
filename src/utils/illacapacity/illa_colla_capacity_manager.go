package illacapacity

type ILLACollaCapacityManager struct {
	API    *IllaCapacityRestAPI
	TeamID int
	TXUID  string
}

func NewILLACollaCapacityManager(teamID int) (*ILLACollaCapacityManager, error) {
	// new API
	api, errorInNewAPI := NewIllaCapacityRestAPI()
	if errorInNewAPI != nil {
		return nil, errorInNewAPI
	}
	// start tx
	txuid, errInStartTransaction := api.StartTransaction(teamID)
	if errInStartTransaction != nil {
		return nil, errInStartTransaction
	}
	// feedback
	return &ILLACollaCapacityManager{
		API:    api,
		TXUID:  txuid,
		TeamID: teamID,
	}, nil
}

func (i *ILLACollaCapacityManager) CostColla(num int64) error {
	req := NewUpdateCapacityRequestByParam(
		i.TeamID,
		i.TXUID,
		TRANSACTION_SERIAL_MODIFY_TARGET_CAPACITY_BALANCE,
		TRANSACTION_SERIAL_MODIFY_METHOD_DEDUCT,
		num,
		CAPACITY_INSTANCE_TYPE_COLLA,
		COLLA_DEFAULT_INSTANCE_ID,
	)
	errInReq := i.API.UpdateCapacityWithNegativeValue(req)
	if errInReq != nil {
		i.API.CancelTransaction(i.TXUID)
		return errInReq
	}
	i.API.CommitTransaction(i.TXUID)
	return nil
}

func (i *ILLACollaCapacityManager) TestColla(num int64) error {
	req := NewUpdateCapacityRequestByParam(
		i.TeamID,
		i.TXUID,
		TRANSACTION_SERIAL_MODIFY_TARGET_CAPACITY_BALANCE,
		TRANSACTION_SERIAL_MODIFY_METHOD_DEDUCT,
		num,
		CAPACITY_INSTANCE_TYPE_COLLA,
		COLLA_DEFAULT_INSTANCE_ID,
	)
	errInReq := i.API.TestCapacity(req)
	if errInReq != nil {
		return errInReq
	}
	return nil
}
