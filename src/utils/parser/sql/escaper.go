package parser_sql

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/illacloud/builder-backend/src/utils/resourcelist"
)

var SerializedParameterizedSQLList = map[int]bool{
	resourcelist.TYPE_POSTGRESQL_ID: true,
	resourcelist.TYPE_ORACLE_9I_ID:  true,
	resourcelist.TYPE_ORACLE_ID:     true,
}

var SerializedParameterPrefixMap = map[int]string{
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

func (sqlEscaper *SQLEscaper) IsSerializedParameterizedSQL() bool {
	itIs, hit := SerializedParameterizedSQLList[sqlEscaper.ResourceType]
	return itIs && hit
}

func (sqlEscaper *SQLEscaper) GetSerializedParameterPrefixMap() string {
	prefix, hit := SerializedParameterPrefixMap[sqlEscaper.ResourceType]
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

func reflectVariableToSlice(variable interface{}) ([]interface{}, error) {
	// try interface slice
	if variableAssertedInInterface, pass := variable.([]interface{}); pass {
		return variableAssertedInInterface, nil
	}
	return nil, errors.New("provided variables assert to slice failed")
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
			return "", errors.New("provided variables assert to bool failed")
		}
		if variableAsserted {
			return "TRUE", nil
		} else {
			return "FALSE", nil
		}
	case reflect.Slice:
		fmt.Printf("???: %+v\n", variable)
		// try int slice
		variableAssertedInInt, variableMappedIntValueAssertPass := variable.([]int)
		if variableMappedIntValueAssertPass {
			finalString := ""
			for i, subVar := range variableAssertedInInt {
				if i != 0 {
					finalString += ", "
				}
				subVarInString, errInReflect := reflectVariableToString(subVar)
				if errInReflect != nil {
					return "", errInReflect
				}
				finalString += "'" + subVarInString + "'"
			}
			return finalString, nil
		}
		// try float64 slice
		variableAssertedInFloat64, variableMappedFloat64ValueAssertPass := variable.([]float64)
		if variableMappedFloat64ValueAssertPass {
			finalString := ""
			for i, subVar := range variableAssertedInFloat64 {
				if i != 0 {
					finalString += ", "
				}
				subVarInString, errInReflect := reflectVariableToString(subVar)
				if errInReflect != nil {
					return "", errInReflect
				}
				finalString += "'" + subVarInString + "'"
			}
			return finalString, nil
		}
		// try string slice
		variableAssertedInString, variableMappedStringValueAssertPass := variable.([]string)
		if variableMappedStringValueAssertPass {
			finalString := ""
			for i, subVar := range variableAssertedInString {
				if i != 0 {
					finalString += ", "
				}
				subVarInString, errInReflect := reflectVariableToString(subVar)
				if errInReflect != nil {
					return "", errInReflect
				}
				finalString += "'" + subVarInString + "'"
			}
			return finalString, nil
		}
		// try interface slice
		variableAssertedInInterface, variableMappedInterfaceValueAssertPass := variable.([]interface{})
		if variableMappedInterfaceValueAssertPass {
			finalString := ""
			for i, subVar := range variableAssertedInInterface {
				if i != 0 {
					finalString += ", "
				}
				subVarInString, errInReflect := reflectVariableToString(subVar)
				if errInReflect != nil {
					return "", errInReflect
				}
				finalString += "'" + subVarInString + "'"
			}
			return finalString, nil
		}

		return "", errors.New("invalied array type inputed")
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
	// init sql rune serial list
	sqlRuneSerialList := make([]int, 0)
	sqlRuneList := []rune(sql)
	// convert to rune slice
	for i, _ := range sqlRuneList {
		sqlRuneSerialList = append(sqlRuneSerialList, i)
	}

	// get next char method.
	getNextChar := func(serial int) (rune, error) {
		if len(sqlRuneSerialList)-1 <= serial {
			return rune(0), errors.New("over range")
		}
		return sqlRuneList[sqlRuneSerialList[serial+1]], nil
	}

	charSerial := -1
	for {
		charSerial++
		if charSerial > len(sqlRuneSerialList)-1 {
			break
		}
		c := sqlRuneList[sqlRuneSerialList[charSerial]]

		// fmt.Printf("[%d:%d] char: %s\n", sqlRuneSerialList[charSerial], charSerial, string(c))
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

			// process variable
			variableMappedValue, hitVariable := escapedArgs[strings.TrimSpace(variable)]
			variableContent := ""
			// missing variable
			if !hitVariable {
				if singleQuoteStart {
					variableContent = "''"
				} else if doubleQuoteStart {
					variableContent = "\"\""
				} else {
					variableContent = escapedBracketWithVariable
				}
				// hit variable
				// - check if it is safe mode:
				//     - 'safe-mode', format the var with template
				//     - 'unsafe-mode', just replace the template with var
				// - then, check if param is a array,
				//     - it is, split array into single variable and push them into userArgs, and rewrite template variable into '(?, ?, ...)' stype
				// 	       - then, process single var by mode, also note the MySQL instance need convert all single var into string value.
			} else {
				// hit variable
				fmt.Printf("-- [DUMP] sqlEscaper.ResourceType: %+v\n", sqlEscaper.ResourceType)
				fmt.Printf("-- [DUMP] sqlEscaper.SafeMode(): %+v\n", safeMode)
				fmt.Printf("-- [DUMP] sqlEscaper.IsSerializedParameterizedSQL(): %+v\n", sqlEscaper.IsSerializedParameterizedSQL())
				fmt.Printf("-- [DUMP] sqlEscaper.GetSerializedParameterPrefixMap(): %+v\n", sqlEscaper.GetSerializedParameterPrefixMap())

				// replace sql param
				if !safeMode {
					// unsafe mode
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
				} else {
					// safe mode
					// check the variable is array
					variableMappedValueInSlice, errInReflectVariableToSlice := reflectVariableToSlice(variableMappedValue)
					fmt.Printf("[DUMP] variableMappedValue: %+v\n", variableMappedValue)
					fmt.Printf("[DUMP] variableMappedValueInSlice: %+v\n", variableMappedValueInSlice)
					fmt.Printf("[DUMP] errInReflectVariableToSlice: %+v\n", errInReflectVariableToSlice)
					variableIsArray := errInReflectVariableToSlice == nil
					if !variableIsArray {
						fmt.Printf("---------- variable in safe mode, and it is a basic variable input!\n")
						// process variable content
						if sqlEscaper.IsSerializedParameterizedSQL() {
							// safe mode, with serialized param
							variableContent = fmt.Sprintf("%s%d", sqlEscaper.GetSerializedParameterPrefixMap(), usedArgsSerial)
							usedArgsSerial++
						} else {
							// safe mode, with "?" as param
							variableContent = "?"
						}
						// process user args
						if sqlEscaper.ResourceType == resourcelist.TYPE_MYSQL_ID {
							// hack for mysql, according to this link: https://github.com/sidorares/node-mysql2/issues/1239#issuecomment-718471799
							// the MysQL 8.0.22 above version only accept string type valiable, so convert all varable to string
							variableMappedValueInString, errInReflectVariableToString := reflectVariableToString(variableMappedValue)
							if errInReflectVariableToString != nil {
								return "", nil, errInReflectVariableToString
							}
							userArgs = append(userArgs, variableMappedValueInString)
						} else {
							userArgs = append(userArgs, variableMappedValue)
						}
					} else {
						fmt.Printf("---------- variable in safe mode, and it is a slice variable input!\n")
						// process variable content
						variableMappedValueInSliceLen := len(variableMappedValueInSlice)
						for i := 0; i < variableMappedValueInSliceLen; i++ {
							if i > 0 {
								variableContent += ", "
							}
							if sqlEscaper.IsSerializedParameterizedSQL() {
								// safe mode, with serialized param
								variableContent += fmt.Sprintf("%s%d", sqlEscaper.GetSerializedParameterPrefixMap(), usedArgsSerial)
								usedArgsSerial++
							} else {
								// safe mode, with "?" as param
								variableContent += "?"
							}
						}
						// process user args
						for _, subVariableMappedValue := range variableMappedValueInSlice {
							if sqlEscaper.ResourceType == resourcelist.TYPE_MYSQL_ID {
								// hack for mysql, according to this link: https://github.com/sidorares/node-mysql2/issues/1239#issuecomment-718471799
								// the MysQL 8.0.22 above version only accept string type valiable, so convert all varable to string
								subVariableMappedValueInString, errInReflectVariableToString := reflectVariableToString(subVariableMappedValue)
								if errInReflectVariableToString != nil {
									return "", nil, errInReflectVariableToString
								}
								userArgs = append(userArgs, subVariableMappedValueInString)
							} else {
								userArgs = append(userArgs, subVariableMappedValue)
							}
						}
					}

				}
			}

			// process bracket
			if singleQuoteStart {
				initConcatStringTargetsIndex(singleQuoteSegmentCounter)
				fmt.Printf("-- [DUMP] fill in concatStringTargets[%d]: %s\n", singleQuoteSegmentCounter, newStringConcatTarget(variableContent, true).Target.String())
				concatStringTargets[singleQuoteSegmentCounter] = newStringConcatTarget(variableContent, true)
				singleQuoteSegPlus()
			} else if doubleQuoteStart {
				initConcatStringTargetsIndex(doubleQuoteSegmentCounter)
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
			fmt.Printf("[DUMP] single quote end at (%d)\n", charSerial)

			// check if is escape character
			nextChar, errInGetNextChar := getNextChar(charSerial)
			fmt.Printf("-- [DUMP] single quote nextChar: %s\n", string(nextChar))

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
			concatResult := formatConcatTarget(sqlEscaper, concatStringTargets, singleQuoteStart, doubleQuoteStart)
			fmt.Printf("-- [DUMP] single quote concatStringTargets: %v\n", concatStringTargets)
			fmt.Printf("-- [DUMP] single quote concatResult: %s\n", concatResult)
			ret.WriteString(concatResult)

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
			ret.WriteString(formatConcatTarget(sqlEscaper, concatStringTargets, singleQuoteStart, doubleQuoteStart))

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

func formatConcatTarget(sqlEscaper *SQLEscaper, concatStringTargets []*stringConcatTarget, singleQuoteStart bool, doubleQuoteStart bool) string {
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
				// process variable type cast
				if target.IsVariable {
					target.concat(sqlEscaper.GetParameterTextTypeCastList())
				}
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
