package parser_sql

import (
	"fmt"
	"strings"

	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

var SerlizedParameterizedSQLList = map[int]bool{
	resourcelist.TYPE_POSTGRESQL_ID: true,
}

type SQLEscaper struct {
	ResourceType int `json:"resourceType"`
}

func NewSQLEscaper(resourceType int) *SQLEscaper {
	return &SQLEscaper{
		ResourceType: resourceType,
	}
}

func (sqlEscaper *SQLEscaper) IsSerlizedParameterizedSQL() bool {
	itIs, hit := SerlizedParameterizedSQLList[sqlEscaper.ResourceType]
	return itIs && hit
}

func (sqlEscaper *SQLEscaper) buildEscapedArgsLookupTable(args map[string]interface{}) (map[string]interface{}, error) {
	escapedArgs := make(map[string]interface{}, 0)
	for key, value := range args {
		escapedArgs[strings.TrimSpace(key)] = value
	}
	return escapedArgs, nil
}

func (sqlEscaper *SQLEscaper) EscapeSQLActionTemplate(sql string, args map[string]interface{}) (string, []interface{}, error) {
	escapedArgs, errInBuildArgsLT := sqlEscaper.buildEscapedArgsLookupTable(args)
	if errInBuildArgsLT != nil {
		return "", []interface{}{}, errInBuildArgsLT
	}
	ret := ""
	variable := ""
	escapedBracketWithVariable := ""
	leftBraketCounter := 0
	rightBraketCounter := 0
	usedArgsSerial := 1 // serial start from 1
	userArgs := make([]interface{}, 0)
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
		ret += escapedBracketWithVariable + "{"
		escapedBracketWithVariable = ""
	}
	escapeIllegalRightBracket := func() {
		rightBraketCounter = 0
		ret += escapedBracketWithVariable + "}"
		escapedBracketWithVariable = ""
	}
	isIgnoredCharacter := func(c rune) bool {
		switch c {
		case '\t', '\n', '\v', '\f', '\r', ' ':
			return true
		}
		return false
	}
	for _, c := range sql {

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

			variableMappedValue, hitVariable := escapedArgs[variable]
			if !hitVariable {
				ret += escapedBracketWithVariable
			} else {
				// replace sql param
				if sqlEscaper.IsSerlizedParameterizedSQL() {
					ret += fmt.Sprintf("$%d", usedArgsSerial)
				} else {
					ret += "?"
				}
				// record param serial
				userArgs = append(userArgs, variableMappedValue)
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
		ret += escapedBracketWithVariable + string(c)
		escapedBracketWithVariable = ""
		variable = ""
		continue
	}
	return ret, userArgs, nil
}
