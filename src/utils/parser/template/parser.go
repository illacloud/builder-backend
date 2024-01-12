package parser_template

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type JSONNumberConvertor struct {
	Payload json.Number `json:"payload"`
}

func (c *JSONNumberConvertor) ExportNumberInString() string {
	return string(c.Payload)
}

func ExportFloat64ToNumberInString(payload float64) string {
	dummyJSON := map[string]interface{}{
		"payload": payload,
	}
	dummyJSONInByte, _ := json.Marshal(dummyJSON)
	jsonNumberConvertor := &JSONNumberConvertor{}
	json.Unmarshal(dummyJSONInByte, &jsonNumberConvertor)
	return jsonNumberConvertor.ExportNumberInString()
}

func ExtractVariableNameConst(template string) []string {
	variableNames := make([]string, 0)

	variableLT := make(map[string]string, 0)
	processesPrompt := ""
	variable := ""
	escapedBracketWithVariable := ""
	leftBraketCounter := 0
	rightBraketCounter := 0
	leftBracketPlus := func() {
		leftBraketCounter++
		escapedBracketWithVariable += "{"
	}
	rightBracketPlus := func() {
		rightBraketCounter++
		escapedBracketWithVariable += "}"
	}
	escapeIllegalLeftBracket := func() {
		leftBraketCounter = 0
		processesPrompt += escapedBracketWithVariable + "{"
		escapedBracketWithVariable = ""
	}
	escapeIllegalRightBracket := func() {
		rightBraketCounter = 0
		processesPrompt += escapedBracketWithVariable + "}"
		escapedBracketWithVariable = ""
	}
	isIgnoredCharacter := func(c rune) bool {
		switch c {
		case '\t', '\n', '\v', '\f', '\r', ' ':
			return true
		}
		return false
	}
	for _, c := range template {

		// process bracket
		// '' + '{' or '{' + '{'
		if c == '{' && leftBraketCounter <= 1 {
			leftBracketPlus()
			continue
		}
		// '{{...' + '{'
		if c == '{' && leftBraketCounter > 1 {
			escapeIllegalLeftBracket()
			continue
		}
		// '}...' + '{'
		if c == '{' && rightBraketCounter > 0 {
			escapeIllegalRightBracket()
			continue
		}
		// '' + '}' or '{' + '}'
		if c == '}' && leftBraketCounter != 2 && rightBraketCounter == 0 {
			escapeIllegalRightBracket()
			continue
		}
		// '{{' + '}'
		if c == '}' && leftBraketCounter == 2 && rightBraketCounter == 0 {
			rightBracketPlus()
			continue
		}
		// '{{' + '}}', hit!
		if c == '}' && leftBraketCounter == 2 && rightBraketCounter == 1 {
			rightBraketCounter++
			escapedBracketWithVariable += "}"
			// collect variable name
			variableNames = append(variableNames, variable)
			// process varibale signal

			variableMappedValue, hitVariable := variableLT[variable]
			if !hitVariable {
				processesPrompt += escapedBracketWithVariable
			} else {
				processesPrompt += variableMappedValue
			}
			escapedBracketWithVariable = ""
			variable = ""
			continue
		}
		// process bracker inner (record variable name)
		if leftBraketCounter == 2 && rightBraketCounter == 0 {
			// filter escape character
			if isIgnoredCharacter(c) {
				continue
			}
			// collect variable name
			variable += string(c)
			escapedBracketWithVariable += string(c)
			continue
		}
		// process other utf-8 character
		leftBraketCounter = 0
		rightBraketCounter = 0
		processesPrompt += escapedBracketWithVariable + string(c)
		escapedBracketWithVariable = ""
		variable = ""
		continue
	}

	return variableNames
}

func AssembleTemplateWithVariable(template string, variableLT map[string]interface{}) (string, error) {
	// check if do not have variable to replace
	if len(variableLT) == 0 {
		return template, nil
	}
	// check if template is json
	templateIsJSON := false
	var templateInJSONObject interface{}
	errInUnmarshalTemplate := json.Unmarshal([]byte(template), &templateInJSONObject)
	if errInUnmarshalTemplate == nil {
		templateIsJSON = true
	}

	// process start
	processesPrompt := ""
	variable := ""
	escapedBracketWithVariable := ""
	leftBraketCounter := 0
	rightBraketCounter := 0
	leftBracketPlus := func() {
		leftBraketCounter++
		escapedBracketWithVariable += "{"
	}
	rightBracketPlus := func() {
		rightBraketCounter++
		escapedBracketWithVariable += "}"
	}
	escapeIllegalLeftBracket := func() {
		leftBraketCounter = 0
		processesPrompt += escapedBracketWithVariable + "{"
		escapedBracketWithVariable = ""
	}
	escapeIllegalRightBracket := func() {
		rightBraketCounter = 0
		processesPrompt += escapedBracketWithVariable + "}"
		escapedBracketWithVariable = ""
	}
	isIgnoredCharacter := func(c rune) bool {
		switch c {
		case '\t', '\n', '\v', '\f', '\r', ' ':
			return true
		}
		return false
	}
	assertDataAndConvertToString := func(data interface{}) (string, error) {
		switch data.(type) {
		case int:
			dataInInt := data.(int)
			return strconv.Itoa(dataInInt), nil
		case int64:
			dataInInt64 := data.(int64)
			return strconv.FormatInt(dataInInt64, 10), nil
		case float32:
		case float64:
			dataInFloat64 := data.(float64)
			return ExportFloat64ToNumberInString(dataInFloat64), nil
		case string:
			finalStr := data.(string)
			if templateIsJSON {
				finalStr = strings.Replace(finalStr, "\"", "\\\"", -1)
				finalStr = strings.Replace(finalStr, "\n", "\\n", -1)
			}
			return finalStr, nil
		case bool:
			dataInBool := data.(bool)
			if dataInBool {
				return "true", nil
			}
			return "false", nil
		default:
			// treat other types as json
			dataInJsonByte, errInMarshal := json.Marshal(data)
			if errInMarshal != nil {
				return "", errInMarshal
			}
			return string(dataInJsonByte), nil
		}
		return "", errors.New("can not convert target data into string")
	}
	for _, c := range template {

		// process bracket
		// '' + '{' or '{' + '{'
		if c == '{' && leftBraketCounter <= 1 {
			leftBracketPlus()
			continue
		}
		// '{{...' + '{'
		if c == '{' && leftBraketCounter > 1 {
			escapeIllegalLeftBracket()
			continue
		}
		// '}...' + '{'
		if c == '{' && rightBraketCounter > 0 {
			escapeIllegalRightBracket()
			continue
		}
		// '' + '}' or '{' + '}'
		if c == '}' && leftBraketCounter != 2 && rightBraketCounter == 0 {
			escapeIllegalRightBracket()
			continue
		}
		// '{{' + '}'
		if c == '}' && leftBraketCounter == 2 && rightBraketCounter == 0 {
			rightBracketPlus()
			continue
		}
		// '{{' + '}}', hit!
		if c == '}' && leftBraketCounter == 2 && rightBraketCounter == 1 {
			rightBraketCounter++
			escapedBracketWithVariable += "}"
			// process varibale signal

			variableMappedValue, hitVariable := variableLT[variable]
			if !hitVariable {
				processesPrompt += escapedBracketWithVariable
			} else {
				valueInString, errInConvertData := assertDataAndConvertToString(variableMappedValue)
				if errInConvertData != nil {
					return "", errInConvertData
				}
				processesPrompt += valueInString
			}
			escapedBracketWithVariable = ""
			variable = ""
			continue
		}
		// process bracker inner (record variable name)
		if leftBraketCounter == 2 && rightBraketCounter == 0 {
			// filter escape character
			if isIgnoredCharacter(c) {
				continue
			}
			// collect variable name
			variable += string(c)
			escapedBracketWithVariable += string(c)
			continue
		}
		// process other utf-8 character
		leftBraketCounter = 0
		rightBraketCounter = 0
		processesPrompt += escapedBracketWithVariable + string(c)
		escapedBracketWithVariable = ""
		variable = ""
		continue
	}
	return processesPrompt, nil
}
