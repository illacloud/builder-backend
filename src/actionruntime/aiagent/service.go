// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aiagent

import (
	"errors"
	"fmt"

	"github.com/illacloud/builder-backend/src/actionruntime/common"
	resourcemanager "github.com/illacloud/builder-backend/src/utils/illaresourcemanagersdk"
	parser_template "github.com/illacloud/builder-backend/src/utils/parser/template"
)

type AIAgentConnector struct {
	Action AIAgentTemplate
}

// AI Agent have no validate resource options method
func (r *AIAgentConnector) ValidateResourceOptions(resourceOptions map[string]interface{}) (common.ValidateResult, error) {
	return common.ValidateResult{Valid: true}, nil
}

func (r *AIAgentConnector) ValidateActionTemplate(actionOptions map[string]interface{}) (common.ValidateResult, error) {
	fmt.Printf("[DUMP] actionOptions: %+v \n", actionOptions)
	_, errorInNewRequest := resourcemanager.NewRunAIAgentRequest(actionOptions)
	if errorInNewRequest != nil {
		return common.ValidateResult{Valid: false}, errorInNewRequest
	}

	return common.ValidateResult{Valid: true}, nil
}

// AI Agent have no test connection method
func (r *AIAgentConnector) TestConnection(resourceOptions map[string]interface{}) (common.ConnectionResult, error) {
	return common.ConnectionResult{Success: false}, errors.New("unsupported type: AI Agent")
}

// AI Agent have no meta info
func (r *AIAgentConnector) GetMetaInfo(resourceOptions map[string]interface{}) (common.MetaInfoResult, error) {
	return common.MetaInfoResult{Success: false}, errors.New("unsupported type: AI Agent")
}

func (r *AIAgentConnector) Run(resourceOptions map[string]interface{}, actionOptions map[string]interface{}, rawActionOptions map[string]interface{}) (common.RuntimeResult, error) {
	res := common.RuntimeResult{
		Success: false,
		Rows:    []map[string]interface{}{},
		Extra:   map[string]interface{}{},
	}
	getContext := func() map[string]interface{} {
		contextRaw := rawActionOptions["context"]
		context, _ := contextRaw.(map[string]interface{})
		return context
	}
	getInput := func() string {
		return actionOptions["input"].(string)
	}
	getVariables := func() []interface{} {
		variablesRaw := actionOptions["variables"]
		variables, _ := variablesRaw.([]interface{})
		return variables
	}
	preprocessKVPairSliceWithContext := func(KVPairSlice []interface{}, context map[string]interface{}) ([]map[string]interface{}, error) {
		var errInPreprocessTemplate error
		if len(KVPairSlice) > 0 {
			newKVPairSlice := make([]map[string]interface{}, 0)
			for _, kvpair := range KVPairSlice {
				kvpairAsserted, _ := kvpair.(map[string]interface{})
				newKVPair := map[string]interface{}{
					"key":          "",
					"value":        "",
					"defaultValue": "",
				}
				// process key
				newKVPair["key"] = kvpairAsserted["key"]

				// process defaultValue
				if kvpairAsserted["defaultValue"] != nil {
					newKVPair["defaultValue"] = kvpairAsserted["defaultValue"]
				}

				// process value
				value, hitValue := kvpairAsserted["value"]
				if !hitValue {
					return nil, errors.New("preprocessKVPairSlice() can not find value field \"value\"")
				}
				valueInString, assertStringPass := value.(string)

				// not a string type value
				if !assertStringPass {
					newKVPair["value"] = value
				} else {
					if newKVPair["value"], errInPreprocessTemplate = parser_template.AssembleTemplateWithVariable(valueInString, context); errInPreprocessTemplate != nil {
						return nil, errInPreprocessTemplate
					}
				}
				newKVPairSlice = append(newKVPairSlice, newKVPair)
			}
			return newKVPairSlice, nil
		}
		return nil, nil
	}

	// replace template with context
	context := getContext()
	actionOptions["input"], _ = parser_template.AssembleTemplateWithVariable(getInput(), context)
	actionOptions["variables"], _ = preprocessKVPairSliceWithContext(getVariables(), context)

	// call api
	api, errInNewAPI := resourcemanager.NewIllaResourceManagerRestAPI()
	if errInNewAPI != nil {
		return res, errInNewAPI
	}
	api.OpenDebug()
	fmt.Printf("[DUMP] r: %+v\n", r)
	fmt.Printf("[DUMP] rawActionOptions: %+v\n", rawActionOptions)
	fmt.Printf("[DUMP] actionOptions: %+v\n", actionOptions)
	runAIAgentResult, errInRunAIAgent := api.RunAIAgent(actionOptions)
	fmt.Printf("[DUMP] runAIAgentResult: %+v\n", runAIAgentResult)
	fmt.Printf("[DUMP] errInRunAIAgent: %+v\n", errInRunAIAgent)

	if errInRunAIAgent != nil {
		return res, errInRunAIAgent
	}

	// feedback
	res.SetSuccess()
	res.Rows = append(res.Rows, runAIAgentResult.ExportAsContent())
	fmt.Printf("[DUMP] res: %+v\n", res)
	return res, nil
}
