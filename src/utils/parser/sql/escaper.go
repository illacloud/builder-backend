package parser_sql

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

var SerlizedParameterizedSQLList = map[int]bool{
	resourcelist.TYPE_POSTGRESQL_ID: true,
	resourcelist.TYPE_ORACLE_9I_ID:  true,
	resourcelist.TYPE_ORACLE_ID:     true,
}

var SerlizedParameterPrefixMap = map[int]string{
	resourcelist.TYPE_POSTGRESQL_ID: "$",
	resourcelist.TYPE_ORACLE_9I_ID:  ":",
	resourcelist.TYPE_ORACLE_ID:     ":",
}

var ParameterTextTypeCastList = map[int]string{
	resourcelist.TYPE_POSTGRESQL_ID: "::text",
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

func (sqlEscaper *SQLEscaper) GetSerlizedParameterPrefixMap() string {
	prefix, hit := SerlizedParameterPrefixMap[sqlEscaper.ResourceType]
	if !hit {
		return ""
	}
	return prefix
}

func (sqlEscaper *SQLEscaper) GetParameterTextTypeCastList() string {
	typeIDF, hit := ParameterTextTypeCastList[sqlEscaper.ResourceType]
	if !hit {
		return ""
	}
	return typeIDF
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
	IsEmpty    bool
}

func newStringConcatTarget(targetString string, isVariable bool) *stringConcatTarget {
	ret := &stringConcatTarget{
		Target:     strings.Builder{},
		IsVariable: isVariable,
		IsEmpty:    false,
	}
	ret.Target.WriteString(targetString)
	return ret
}

func newEmptyStringConcatTarget() *stringConcatTarget {
	ret := &stringConcatTarget{
		Target:     strings.Builder{},
		IsVariable: false,
		IsEmpty:    true,
	}
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

func (i *stringConcatTarget) ExportWithoutQuote() string {
	if i.IsVariable {
		return i.Target.String()
	} else {
		return i.Target.String()
	}
}

func reflectVariableToString(variable interface{}) (string, error) {
	// check type of variable value
	typeOfVariableMappedValue := reflect.TypeOf(variable)
	switch typeOfVariableMappedValue.Kind() {
	case reflect.String:
		variableAsserted, variableMappedValueAssertPass := variable.(string)
		if !variableMappedValueAssertPass {
			return "", errors.New("provided variables assert to string failed")
		}
		return variableAsserted, nil
	case reflect.Int:
		variableAsserted, variableMappedValueAssertPass := variable.(int)
		if !variableMappedValueAssertPass {
			return "", errors.New("provided variables assert to int failed")
		}
		return strconv.Itoa(variableAsserted), nil
	case reflect.Float64:
		variableAsserted, variableMappedValueAssertPass := variable.(float64)
		if !variableMappedValueAssertPass {
			return "", errors.New("provided variables assert to float64 failed")
		}
		return strconv.FormatFloat(variableAsserted, 'f', -1, 64), nil
	case reflect.Bool:
		variableAsserted, variableMappedValueAssertPass := variable.(bool)
		if !variableMappedValueAssertPass {
			return "", errors.New("provided variables assert to float64 failed")
		}
		if variableAsserted {
			return "TRUE", nil
		} else {
			return "FALSE", nil
		}
	default:
		return "", nil
	}
}

// select * from users where name like '%{{first_name}}.{{last_name}}%'
// safeMode for varibale mode, unsafeMode for variable replace mode.
func (sqlEscaper *SQLEscaper) EscapeSQLActionTemplate(sql string, args map[string]interface{}, safeMode bool) (string, []interface{}, error) {
	fmt.Printf("\n\n-- [CALL] EscapeSQLActionTemplate()\n")
	fmt.Printf("    -- [CALL] sql string: %s\n", sql)
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
	fmt.Printf("-- [DUMP] first len(concatStringTargets): %d\n", len(concatStringTargets))

	initConcatStringTargetsIndex := func(index int) {
		for {
			if len(concatStringTargets)-1 < index {
				concatStringTargets = append(concatStringTargets, newEmptyStringConcatTarget())
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
			// process quoute counter, only bump quoute segment counter when having leading character
			fmt.Printf("-- [DUMP] doubleQuoteSegmentCounter: %d\n", doubleQuoteSegmentCounter)
			fmt.Printf("-- [DUMP] len(concatStringTargets): %d\n", len(concatStringTargets))
			if singleQuoteStart && !concatStringTargets[singleQuoteSegmentCounter].IsEmpty {
				singleQuoteSegPlus()
			}
			if doubleQuoteStart && !concatStringTargets[doubleQuoteSegmentCounter].IsEmpty {
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
				fmt.Printf("-- [DUMP] sqlEscaper.ResourceType: %+v\n", sqlEscaper.ResourceType)
				fmt.Printf("-- [DUMP] sqlEscaper.IsSerlizedParameterizedSQL(): %+v\n", sqlEscaper.IsSerlizedParameterizedSQL())
				fmt.Printf("-- [DUMP] sqlEscaper.GetSerlizedParameterPrefixMap(): %+v\n", sqlEscaper.GetSerlizedParameterPrefixMap())
				// replace sql param
				if !safeMode {
					// check type of variable value
					variableMappedValueInString, errInReflect := reflectVariableToString(variableMappedValue)
					if errInReflect != nil {
						return "", nil, errInReflect
					}
					if singleQuoteStart {
						variableContent = fmt.Sprintf("'%s'", variableMappedValueInString)
					} else if doubleQuoteStart {
						variableContent = fmt.Sprintf("\"%s\"", variableMappedValueInString)
					} else {
						variableContent = variableMappedValueInString
					}
				} else if sqlEscaper.IsSerlizedParameterizedSQL() {
					if singleQuoteStart {
						variableContent = fmt.Sprintf("'%s%d'", sqlEscaper.GetSerlizedParameterPrefixMap(), usedArgsSerial)
					} else if doubleQuoteStart {
						variableContent = fmt.Sprintf("\"%s%d\"", sqlEscaper.GetSerlizedParameterPrefixMap(), usedArgsSerial)
					} else {
						variableContent = fmt.Sprintf("%s%d", sqlEscaper.GetSerlizedParameterPrefixMap(), usedArgsSerial)
					}
					usedArgsSerial++
				} else {
					if singleQuoteStart {
						variableContent = "'?'"
					} else if doubleQuoteStart {
						variableContent = "\"?\""
					} else {
						variableContent = "?"
					}
				}
				// record param serial
				userArgs = append(userArgs, variableMappedValue)
			}
			// process type cast
			if singleQuoteStart {
				initConcatStringTargetsIndex(singleQuoteSegmentCounter)
				variableContent += sqlEscaper.GetParameterTextTypeCastList()
				fmt.Printf("-- [DUMP] fill in concatStringTargets[%d]: %s\n", singleQuoteSegmentCounter, newStringConcatTarget(variableContent, true).Target.String())
				concatStringTargets[singleQuoteSegmentCounter] = newStringConcatTarget(variableContent, true)
				singleQuoteSegPlus()
			} else if doubleQuoteStart {
				initConcatStringTargetsIndex(doubleQuoteSegmentCounter)
				variableContent += sqlEscaper.GetParameterTextTypeCastList()
				concatStringTargets[doubleQuoteSegmentCounter] = newStringConcatTarget(variableContent, true)
				doubleQuoteSegPlus()
			} else {
				ret.WriteString(variableContent)
			}
			fmt.Printf("---[DUMP] variableContent: %+v\n", variableContent)
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
			ret.WriteString(formatConcatTarget(concatStringTargets, singleQuoteStart, doubleQuoteStart))

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
			// double quete have no escape, it is string finish quote
			ret.WriteString(formatConcatTarget(concatStringTargets, singleQuoteStart, doubleQuoteStart))

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
			concatStringTargets[singleQuoteSegmentCounter].IsEmpty = false
			continue
		}
		if doubleQuoteStart {
			initConcatStringTargetsIndex(doubleQuoteSegmentCounter)
			concatStringTargets[doubleQuoteSegmentCounter].concat(string(c))
			concatStringTargets[doubleQuoteSegmentCounter].IsEmpty = false
			continue
		}
		// process other utf-8 character
		ret.WriteString(escapedBracketWithVariable)
		ret.WriteRune(c)
		escapedBracketWithVariable = ""
		variable = ""
		continue
	}
	fmt.Printf("[DUMP] escaped SQL: %s\n", ret.String())
	fmt.Printf("[DUMP] escaped SQL params: %+v\n", userArgs)
	return ret.String(), userArgs, nil
}

var formatConcatTargetCalls = 0

func formatConcatTarget(concatStringTargets []*stringConcatTarget, singleQuoteStart bool, doubleQuoteStart bool) string {
	formatConcatTargetCalls++
	fmt.Printf("-- [DUMP] formatConcatTargetCalls: %+v\n", formatConcatTargetCalls)
	var ret strings.Builder
	haveVariable := false
	exportedTarget := make([]string, 0)
	// check if have any variable
	for _, target := range concatStringTargets {
		if target.IsVariable {
			haveVariable = true
		}
	}
	// export with variable
	if haveVariable {
		// only have single valiable
		if len(concatStringTargets) == 1 {
			ret.WriteString(concatStringTargets[0].Export(singleQuoteStart, doubleQuoteStart))
		} else {
			// multi variable
			ret.WriteString("CONCAT(")
			for _, target := range concatStringTargets {
				fmt.Printf("----- [DUMP] target: %+v\n", string(target.Target.String()))
				exportedTarget = append(exportedTarget, target.Export(singleQuoteStart, doubleQuoteStart))
			}
			ret.WriteString(strings.Join(exportedTarget, ", "))
			ret.WriteString(")")
		}
	} else {
		if singleQuoteStart {
			ret.WriteString("'")
		} else if doubleQuoteStart {
			ret.WriteString("\"")
		}
		for _, target := range concatStringTargets {
			ret.WriteString(target.ExportWithoutQuote())
		}
		if singleQuoteStart {
			ret.WriteString("'")
		} else if doubleQuoteStart {
			ret.WriteString("\"")
		}
	}
	return ret.String()

}
