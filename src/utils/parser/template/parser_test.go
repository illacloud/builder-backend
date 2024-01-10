package parser_template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSample(t *testing.T) {
	assert.Nil(t, nil)
}

func TestGetAllVariableNameConstFromActionTemplate1(t *testing.T) {
	actionTemplate := `{"mode": "sql", "query": "select * \nfrom users\njoin orders\non users.id = orders.id\nwhere {{!input1.value}} or lower(users.name) like '%{{input1.value.toLowerCase()}}%'"}`
	variableNames := ExtractVariableNameConst(actionTemplate)

	assert.Equal(t, "!input1.value", variableNames[0], "it should be string \"yes\" ")
	assert.Equal(t, "input1.value.toLowerCase()", variableNames[1], "it should be string \"yes\" ")

}

func TestGetAllVariableNameConstFromActionTemplate2(t *testing.T) {
	actionTemplate := `{"mode": "sql-safe", "query": "select count(distinct email) from users\nwhere DATE_TRUNC('day', created_at AT TIME ZONE 'UTC' AT TIME ZONE 'GMT+8') BETWEEN '{{date1.value}}'::date AND '{{date2.value}}'::date"}`
	variableNames := ExtractVariableNameConst(actionTemplate)

	assert.Equal(t, "date1.value", variableNames[0], "it should be string \"yes\" ")
	assert.Equal(t, "date2.value", variableNames[1], "it should be string \"yes\" ")

}

func TestAssembleTemplateWithVariable_case1_BoolAndString(t *testing.T) {
	actionTemplate := `{"mode": "sql", "query": "select * \nfrom users\njoin orders\non users.id = orders.id\nwhere {{!input1.value}} or lower(users.name) like '%{{input1.value.toLowerCase()}}%'"}`
	dataLT := map[string]interface{}{
		"!input1.value":              false,
		"input1.value.toLowerCase()": "jackmall",
	}
	finalTemplate, errInAssemble := AssembleTemplateWithVariable(actionTemplate, dataLT)
	assert.Nil(t, errInAssemble)
	assert.Equal(t, `{"mode": "sql", "query": "select * \nfrom users\njoin orders\non users.id = orders.id\nwhere false or lower(users.name) like '%jackmall%'"}`, finalTemplate, "it should be equal ")

}

func TestAssembleTemplateWithVariable_case2_integerAndFloat(t *testing.T) {
	actionTemplate := `{"mode": "sql-safe", "query": "select count(distinct email) from users\nwhere DATE_TRUNC('day', created_at AT TIME ZONE 'UTC' AT TIME ZONE 'GMT+8') BETWEEN '{{date1.value}}'::date AND '{{date2.value}}'::date"}`
	dataLT := map[string]interface{}{
		"date1.value": 99811111111231220,
		"date2.value": 14.90000002,
	}
	finalTemplate, errInAssemble := AssembleTemplateWithVariable(actionTemplate, dataLT)
	assert.Nil(t, errInAssemble)
	assert.Equal(t, `{"mode": "sql-safe", "query": "select count(distinct email) from users\nwhere DATE_TRUNC('day', created_at AT TIME ZONE 'UTC' AT TIME ZONE 'GMT+8') BETWEEN '99811111111231220'::date AND '14.90000002'::date"}`, finalTemplate, "it should be equal ")

}

func TestAssembleTemplateWithVariable_case3_WarppedString(t *testing.T) {
	actionTemplate := `{"mode": "sql", "query": "select * \nfrom users\njoin orders\non users.id = orders.id\nwhere {{!input1.value}} or lower(users.name) like '%{{input1.value.toLowerCase()}}%'"}`
	dataLT := map[string]interface{}{
		"!input1.value":              "\"BIG APPLE\"",
		"input1.value.toLowerCase()": "[A\nAA]",
	}
	finalTemplate, errInAssemble := AssembleTemplateWithVariable(actionTemplate, dataLT)
	assert.Nil(t, errInAssemble)
	assert.Equal(t, `{"mode": "sql", "query": "select * \nfrom users\njoin orders\non users.id = orders.id\nwhere \"BIG APPLE\" or lower(users.name) like '%[A\nAA]%'"}`, finalTemplate, "it should be equal ")

}
