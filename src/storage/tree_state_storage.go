package storage

import (
	"fmt"

	"github.com/illacloud/builder-backend/src/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TreeStateStorage struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewTreeStateStorage(logger *zap.SugaredLogger, db *gorm.DB) *TreeStateStorage {
	return &TreeStateStorage{
		logger: logger,
		db:     db,
	}
}

func (impl *TreeStateStorage) Create(treeState *model.TreeState) (int, error) {
	fmt.Printf("Createing tree_state: uid: %v, team_id: %v, app_id: %v, name: %v. \n", treeState.UID, treeState.TeamID, treeState.AppRefID, treeState.Name)
	if err := impl.db.Create(treeState).Error; err != nil {
		return 0, err
	}
	return treeState.ID, nil
}

// insert tree state by component tree
func (impl *TreeStateStorage) CreateComponentTree(app *model.App, parentNodeID int, componentNodeTree *model.ComponentNode) error {
	// convert ComponentNode to TreeState
	currentNode, errInNewTreeState := model.NewTreeStateByAppAndComponentState(app, model.TREE_STATE_TYPE_COMPONENTS, componentNodeTree)
	if errInNewTreeState != nil {
		return errInNewTreeState
	}

	// get parentNode
	parentTreeState := model.NewTreeState()
	var errInRetrieveParentTreeState error
	isSummitNode := true

	// process parentNode
	if parentNodeID != 0 || currentNode.ParentNode == model.TREE_STATE_SUMMIT_NAME {
		// parentNode is in database
		isSummitNode = false
		parentTreeState, errInRetrieveParentTreeState = impl.RetrieveByID(app.ExportTeamID(), parentNodeID)
		if errInRetrieveParentTreeState != nil {
			return errInRetrieveParentTreeState
		}
	} else if componentNodeTree.ParentNode != "" && componentNodeTree.ParentNode != model.TREE_STATE_SUMMIT_NAME {
		//  parentNode not in database but parentNode exists
		isSummitNode = false
		parentTreeState, errInRetrieveParentTreeState = impl.RetrieveEditVersionByAppAndName(app.ExportTeamID(), currentNode.AppRefID, currentNode.StateType, componentNodeTree.ParentNode)
		if errInRetrieveParentTreeState != nil {
			return errInRetrieveParentTreeState
		}
	}

	// no parentNode, currentNode is tree summit
	if isSummitNode && currentNode.Name != model.TREE_STATE_SUMMIT_NAME {
		// get root node
		parentTreeState, errInRetrieveParentTreeState = impl.RetrieveEditVersionByAppAndName(app.ExportTeamID(), currentNode.AppRefID, currentNode.StateType, model.TREE_STATE_SUMMIT_NAME)
		if errInRetrieveParentTreeState != nil {
			return errInRetrieveParentTreeState
		}
	}

	// hook parent node for current node
	currentNode.ParentNodeRefID = parentTreeState.ID

	// insert currentNode and get id
	_, errInCreateTreeState := impl.Create(currentNode)
	if errInCreateTreeState != nil {
		return errInCreateTreeState
	}

	// fill currentNode id into parentNode.ChildrenNodeRefIDs and update it
	if currentNode.Name != model.TREE_STATE_SUMMIT_NAME {
		parentTreeState.AppendChildrenNodeRefIDs(currentNode.ID)
		// update parentNode
		errInUpdateTreeState := impl.Update(parentTreeState)
		if errInUpdateTreeState != nil {
			return errInUpdateTreeState
		}
	}

	// create currentNode.ChildrenNode
	for _, childrenComponentNode := range componentNodeTree.ChildrenNode {
		if err := impl.CreateComponentTree(app, currentNode.ID, childrenComponentNode); err != nil {
			return err
		}
	}
	return nil
}

func (impl *TreeStateStorage) DeleteComponentTree(currentNode *model.TreeState) error {
	// get currentTreeState by displayName from database
	currentTreeState, errInRetrieveCurrentNode := impl.RetrieveEditVersionByAppAndName(currentNode.TeamID, currentNode.AppRefID, currentNode.StateType, currentNode.Name)
	if errInRetrieveCurrentNode != nil {
		return errInRetrieveCurrentNode
	}

	// unlink parentNode
	if currentTreeState.ParentNodeRefID != 0 { // parentNode is in database
		// get parentNode
		parentTreeState, errInRetrieveParentTreeState := impl.RetrieveByID(currentNode.TeamID, currentTreeState.ID)
		if errInRetrieveParentTreeState != nil {
			return errInRetrieveParentTreeState
		}
		// update parentNode for unlink
		parentTreeState.RemoveChildrenNodeRefIDs(currentTreeState.ID)
		errInUpdateParentTreeState := impl.Update(parentTreeState)
		if errInUpdateParentTreeState != nil {
			return errInUpdateParentTreeState
		}
	}

	// get all children nodes recursive
	childrenNodes := []*model.TreeState{}
	errInRetrieveChildrenNodes := impl.retrieveChildrenNodes(currentTreeState, &childrenNodes)
	if errInRetrieveChildrenNodes != nil {
		return errInRetrieveChildrenNodes
	}

	// put current node into children nodes slice for delete them all
	childrenNodes = append(childrenNodes, currentTreeState)

	// delete all children nodes
	var childrenNodeIDs []int
	for _, node := range childrenNodes {
		childrenNodeIDs = append(childrenNodeIDs, node.ID)
	}
	errInDeleteChildrenNodes := impl.DeleteByIDs(childrenNodeIDs)
	if errInDeleteChildrenNodes != nil {
		return errInDeleteChildrenNodes
	}

	return nil
}

func (impl *TreeStateStorage) retrieveChildrenNodes(treeState *model.TreeState, childrenNodes *[]*model.TreeState) error {
	ids, err := treeState.ExportChildrenNodeRefIDs()
	if err != nil {
		return err
	}
	nodes, errInRetrieveByIds := impl.RetrieveByIDs(ids)
	if errInRetrieveByIds != nil {
		return errInRetrieveByIds
	}
	for _, node := range nodes {
		if err := impl.retrieveChildrenNodes(node, childrenNodes); err != nil {
			return err
		}
		*childrenNodes = append(*childrenNodes, node)
	}
	return nil
}

// @todo: add tree ref circle checker.
func (impl *TreeStateStorage) MoveTreeStateNode(currentNode *model.TreeState) error {
	// get currentTreeState by name
	currentTreeState, errInRetrieveCurrentTreeState := impl.RetrieveEditVersionByAppAndName(currentNode.TeamID, currentNode.AppRefID, currentNode.StateType, currentNode.Name)
	if errInRetrieveCurrentTreeState != nil {
		return errInRetrieveCurrentTreeState
	}

	// get oldParentTreeState by id
	oldParentTreeState, errInRetrieveOldParentTreeState := impl.RetrieveByID(currentNode.TeamID, currentTreeState.ParentNodeRefID)
	if errInRetrieveOldParentTreeState != nil {
		return errInRetrieveOldParentTreeState
	}

	// get newParentTreeState by name
	var newParentTreeState *model.TreeState
	var errInRetrieveNewParentTreeState error
	switch currentNode.StateType {
	case model.TREE_STATE_TYPE_COMPONENTS:
		newParentTreeState, errInRetrieveNewParentTreeState = impl.RetrieveEditVersionByAppAndName(currentNode.TeamID, currentNode.AppRefID, currentNode.StateType, currentNode.ParentNode)
		if errInRetrieveNewParentTreeState != nil {
			return errInRetrieveNewParentTreeState
		}
	default:
		return nil
	}

	// fill into database
	// update currentTreeState
	currentTreeState.ParentNodeRefID = newParentTreeState.ID
	if err := impl.Update(currentTreeState); err != nil {
		return err
	}

	// add now TreeState id into new parent TreeState.ChildrenNodeRefIDs
	newParentTreeState.AppendChildrenNodeRefIDs(currentTreeState.ID)

	// update newParentTreeState
	if err := impl.Update(newParentTreeState); err != nil {
		return err
	}

	// remove now TreeState id from old parent TreeState.ChildrenNodeRefIDs
	oldParentTreeState.RemoveChildrenNodeRefIDs(currentTreeState.ID)

	// update oldParentTreeState
	if err := impl.Update(oldParentTreeState); err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateStorage) Delete(teamID int, treeStateID int) error {
	if err := impl.db.Where("id = ? AND team_id = ?", treeStateID, teamID).Delete(&model.TreeState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateStorage) DeleteByIDs(treeStateIDs []int) error {
	if err := impl.db.Where("(id) IN ?", treeStateIDs).Delete(&model.TreeState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateStorage) Update(treeState *model.TreeState) error {
	if err := impl.db.Model(treeState).Where("id = ?", treeState.ID).UpdateColumns(treeState).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateStorage) RetrieveByID(teamID int, treeStateID int) (*model.TreeState, error) {
	treeState := &model.TreeState{}
	if err := impl.db.Where("team_id = ? AND id = ?", teamID, treeStateID).First(&treeState).Error; err != nil {
		return &model.TreeState{}, err
	}
	return treeState, nil
}

func (impl *TreeStateStorage) RetrieveByIDs(ids []int) ([]*model.TreeState, error) {
	treeStates := []*model.TreeState{}
	if err := impl.db.Where("(id) IN ?", ids).Find(&treeStates).Error; err != nil {
		return nil, err
	}
	return treeStates, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesByVersion(teamID int, version int) ([]*model.TreeState, error) {
	var treeStates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND version = ?", teamID, version).Find(&treeStates).Error; err != nil {
		return nil, err
	}
	return treeStates, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesLatestVersion(teamID int, appID int) (int, error) {
	var treeStates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, appID).Order("version desc").Limit(1).Find(&treeStates).Error; err != nil {
		return 0, err
	}
	if len(treeStates) == 0 {
		return 0, nil
	}
	return treeStates[0].Version, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesByName(teamID int, name string) ([]*model.TreeState, error) {
	var treeStates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND name = ?", teamID, name).Find(&treeStates).Error; err != nil {
		return nil, err
	}
	return treeStates, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesByApp(teamID int, apprefid int, statetype int, version int) ([]*model.TreeState, error) {
	var treeStates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ?", teamID, apprefid, statetype, version).Find(&treeStates).Error; err != nil {
		return nil, err
	}
	return treeStates, nil
}

func (impl *TreeStateStorage) RetrieveEditVersionByAppAndName(teamID int, apprefid int, statetype int, name string) (*model.TreeState, error) {
	var treeState *model.TreeState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND state_type = ? AND version = ? AND name = ?", teamID, apprefid, statetype, model.APP_EDIT_VERSION, name).First(&treeState).Error; err != nil {
		return nil, err
	}
	return treeState, nil
}

func (impl *TreeStateStorage) RetrieveTreeStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, version int) ([]*model.TreeState, error) {
	var treeStates []*model.TreeState
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, version).Find(&treeStates).Error; err != nil {
		return nil, err
	}
	return treeStates, nil
}

func (impl *TreeStateStorage) DeleteAllTypeTreeStatesByApp(teamID int, apprefid int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ?", teamID, apprefid).Delete(&model.TreeState{}).Error; err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateStorage) DeleteAllTypeTreeStatesByTeamIDAppIDAndVersion(teamID int, apprefid int, targetVersion int) error {
	if err := impl.db.Where("team_id = ? AND app_ref_id = ? AND version = ?", teamID, apprefid, targetVersion).Delete(&model.TreeState{}).Error; err != nil {
		return err
	}
	return nil
}
