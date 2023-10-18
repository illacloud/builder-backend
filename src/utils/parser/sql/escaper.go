package parser_sql

import (
	"errors"
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

func (sqlEscaper *SQLEscaper) IsPostgres() bool {
	return sqlEscaper.ResourceType == resourcelist.TYPE_POSTGRESQL_ID
}

func (sqlEscaper *SQLEscaper) buildEscapedArgsLookupTable(args map[string]interface{}) (map[string]interface{}, error) {
	escapedArgs := make(map[string]interface{}, 0)
	for key, value := range args {
		escapedArgs[strings.TrimSpace(key)] = value
	}
	return escapedArgs, nil
}

type stringConcatTarget struct {
	Target     strings.Builder
	IsVariable bool
}

func newStringConcatTarget(targetString string, isVariable bool) *stringConcatTarget {
	ret := &stringConcatTarget{
		Target:     strings.Builder{},
		IsVariable: isVariable,
	}
	ret.Target.WriteString(targetString)
	return ret
}

func (i *stringConcatTarget) concat(str string) {
	i.Target.WriteString(str)
}

func (i *stringConcatTarget) Export(singleQuoteStart bool, doubleQuoteSart bool) string {
	if i.IsVariable {
		return i.Target.String()
	} else {
		if singleQuoteStart {
			return "'" + i.Target.String() + "'"

		} else if doubleQuoteSart {
			return "\"" + i.Target.String() + "\""

		}
	}
	return i.Target.String()
}

// select * from users where name like '%{{first_name}}.{{last_name}}%'
func (sqlEscaper *SQLEscaper) EscapeSQLActionTemplate(sql string, args map[string]interface{}) (string, []interface{}, error) {
	escapedArgs, errInBuildArgsLT := sqlEscaper.buildEscapedArgsLookupTable(args)
	if errInBuildArgsLT != nil {
		return "", []interface{}{}, errInBuildArgsLT
	}
	var ret strings.Builder
	variable := ""
	escapedBracketWithVariable := ""
	leftBraketCounter := 0
	rightBraketCounter := 0
	singleQuoteSegmentCounter := 0
	doubleQuoteSegmentCounter := 0
	singleQuoteStart := false
	doubleQuoteStart := false
	usedArgsSerial := 1 // serial start from 1
	userArgs := make([]interface{}, 0)
	concatStringTargets := make([]*stringConcatTarget, 0)
	initConcatStringTargetsIndex := func(index int) {
		for {
			if len(concatStringTargets)-1 < index {
				concatStringTargets = append(concatStringTargets, newStringConcatTarget("", false))
			} else {
				break
			}
		}
	}
	singleQuoteSegPlus := func() {
		singleQuoteSegmentCounter++
	}
	doubleQuoteSegPlus := func() {
		doubleQuoteSegmentCounter++
	}
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
		ret.WriteString(escapedBracketWithVariable)
		ret.WriteString("{")
		escapedBracketWithVariable = ""
	}
	escapeIllegalRightBracket := func() {
		rightBraketCounter = 0
		ret.WriteString(escapedBracketWithVariable)
		ret.WriteString("}")
		escapedBracketWithVariable = ""
	}
	isIgnoredCharacter := func(c rune) bool {
		switch c {
		case '\t', '\n', '\v', '\f', '\r', ' ':
			return true
		}
		return false
	}
	getNextChar := func(serial int) (rune, error) {
		if len(sql)-1 <= serial {
			return rune(0), errors.New("over range")
		}
		return rune(sql[serial+1]), nil
	}
	charSerial := -1
	for {
		charSerial++
		if charSerial > len(sql)-1 {
			break
		}
		c := rune(sql[charSerial])
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
			// process quoute counter
			if singleQuoteStart {
				singleQuoteSegPlus()
			}
			if doubleQuoteStart {
				doubleQuoteSegPlus()
			}

			// process bracket counter
			rightBraketCounter++
			escapedBracketWithVariable += "}"

			// process variable signal
			variableMappedValue, hitVariable := escapedArgs[variable]
			variableContent := ""
			if !hitVariable {
				if singleQuoteStart {
					variableContent = "''"
				} else if doubleQuoteStart {
					variableContent = "\"\""
				} else {
					variableContent = escapedBracketWithVariable
				}
			} else {
				// replace sql param
				if sqlEscaper.IsSerlizedParameterizedSQL() {
					variableContent = fmt.Sprintf("$%d", usedArgsSerial)
					usedArgsSerial++
				} else {
					variableContent = "?"
				}
				// record param serial
				userArgs = append(userArgs, variableMappedValue)
			}
			if singleQuoteStart {
				initConcatStringTargetsIndex(singleQuoteSegmentCounter)
				if sqlEscaper.IsSerlizedParameterizedSQL() {
					variableContent += "::text"
				}
				concatStringTargets[singleQuoteSegmentCounter] = newStringConcatTarget(variableContent, true)
				singleQuoteSegPlus()
			} else if doubleQuoteStart {
				initConcatStringTargetsIndex(doubleQuoteSegmentCounter)
				if sqlEscaper.IsSerlizedParameterizedSQL() {
					variableContent += "::text"
				}
				concatStringTargets[doubleQuoteSegmentCounter] = newStringConcatTarget(variableContent, true)
				doubleQuoteSegPlus()
			} else {
				ret.WriteString(variableContent)
			}
			escapedBracketWithVariable = ""
			variable = ""
			continue
		}
		// process bracket inner (record variable name)
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
		// process single quote start
		if c == '\'' && !singleQuoteStart && !doubleQuoteStart {
			singleQuoteStart = true
			initConcatStringTargetsIndex(singleQuoteSegmentCounter)
			continue
		}
		// check if is escape character in single quote
		if c == '\\' && singleQuoteStart && !doubleQuoteStart {
			nextChar, errInGetNextChar := getNextChar(charSerial)
			if errInGetNextChar == nil {
				// psotgres specified escape method
				if nextChar == '\'' {
					initConcatStringTargetsIndex(singleQuoteSegmentCounter)
					concatStringTargets[singleQuoteSegmentCounter].concat("\\'")
					charSerial++
					continue
				}
			}
		}
		// single quote end, form concat function to sql
		if c == '\'' && singleQuoteStart && !doubleQuoteStart {
			// check if is escape character
			nextChar, errInGetNextChar := getNextChar(charSerial)
			if errInGetNextChar == nil {
				// psotgres specified escape method
				if nextChar == '\'' {
					initConcatStringTargetsIndex(singleQuoteSegmentCounter)
					concatStringTargets[singleQuoteSegmentCounter].concat("''")
					charSerial++
					continue
				}
			}

			// not escape, it is string finish quote
			ret.WriteString("CONCAT(")
			exportedTarget := make([]string, 0)
			for _, target := range concatStringTargets {
				exportedTarget = append(exportedTarget, target.Export(singleQuoteStart, doubleQuoteStart))
			}
			ret.WriteString(strings.Join(exportedTarget, ", "))
			ret.WriteString(")")
			// clean status
			singleQuoteStart = false
			singleQuoteSegmentCounter = 0
			concatStringTargets = make([]*stringConcatTarget, 0)
			continue
		}
		// process double quote start
		if c == '"' && !doubleQuoteStart && !singleQuoteStart {
			doubleQuoteStart = true
			initConcatStringTargetsIndex(doubleQuoteSegmentCounter)
			continue
		}
		// check if is escape character in double quote
		if c == '\\' && doubleQuoteStart && !singleQuoteStart {
			nextChar, errInGetNextChar := getNextChar(charSerial)
			if errInGetNextChar == nil {
				// psotgres specified escape method
				if nextChar == '"' {
					initConcatStringTargetsIndex(doubleQuoteSegmentCounter)
					concatStringTargets[doubleQuoteSegmentCounter].concat("\\\"")
					charSerial++
					continue
				}
			}
		}
		// double quote end, form concat function to sql
		if c == '"' && doubleQuoteStart && !singleQuoteStart {
			// not escape, it is string finish quote
			ret.WriteString("CONCAT(")
			exportedTarget := make([]string, 0)
			for _, target := range concatStringTargets {
				exportedTarget = append(exportedTarget, target.Export(singleQuoteStart, doubleQuoteStart))
			}
			ret.WriteString(strings.Join(exportedTarget, ", "))
			ret.WriteString(")")
			// clean status
			doubleQuoteStart = false
			doubleQuoteSegmentCounter = 0
			concatStringTargets = make([]*stringConcatTarget, 0)
			continue
		}

		// process bracket process end, reset bracket counter
		leftBraketCounter = 0
		rightBraketCounter = 0

		// process quote inner
		if singleQuoteStart {
			initConcatStringTargetsIndex(singleQuoteSegmentCounter)
			concatStringTargets[singleQuoteSegmentCounter].concat(string(c))
			continue
		}
		if doubleQuoteStart {
			initConcatStringTargetsIndex(doubleQuoteSegmentCounter)
			concatStringTargets[doubleQuoteSegmentCounter].concat(string(c))
			continue
		}
		// process other utf-8 character
		ret.WriteString(escapedBracketWithVariable)
		ret.WriteRune(c)
		escapedBracketWithVariable = ""
		variable = ""
		continue
	}
	return ret.String(), userArgs, nil
}
