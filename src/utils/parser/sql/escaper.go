package parser_sql

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func escapeSQLString(source string) string {
	var j int = 0
	if len(source) == 0 {
		return ""
	}
	tempStr := source[:]
	desc := make([]byte, len(tempStr)*2)
	for i := 0; i < len(tempStr); i++ {
		flag := false
		var escape byte
		switch tempStr[i] {
		case '\r':
			flag = true
			escape = '\r'
			break
		case '\n':
			flag = true
			escape = '\n'
			break
		case '\\':
			flag = true
			escape = '\\'
			break
		case '\'':
			flag = true
			escape = '\''
			break
		case '"':
			flag = true
			escape = '"'
			break
		case '\032':
			flag = true
			escape = 'Z'
			break
		default:
		}
		if flag {
			desc[j] = '\\'
			desc[j+1] = escape
			j = j + 2
		} else {
			desc[j] = tempStr[i]
			j = j + 1
		}
	}
	return string(desc[0:j])
}

func reserveBuffer(buf []byte, appendSize int) []byte {
	newSize := len(buf) + appendSize
	if cap(buf) < newSize {
		newBuf := make([]byte, len(buf)*2+appendSize)
		copy(newBuf, buf)
		buf = newBuf
	}
	return buf[:newSize]
}

func escapeBytesBackslash(buf []byte, v []byte) []byte {
	pos := len(buf)
	buf = reserveBuffer(buf, len(v)*2)

	for _, c := range v {
		switch c {
		case '\x00':
			buf[pos] = '\\'
			buf[pos+1] = '0'
			pos += 2
		case '\n':
			buf[pos] = '\\'
			buf[pos+1] = 'n'
			pos += 2
		case '\r':
			buf[pos] = '\\'
			buf[pos+1] = 'r'
			pos += 2
		case '\x1a':
			buf[pos] = '\\'
			buf[pos+1] = 'Z'
			pos += 2
		case '\'':
			buf[pos] = '\\'
			buf[pos+1] = '\''
			pos += 2
		case '"':
			buf[pos] = '\\'
			buf[pos+1] = '"'
			pos += 2
		case '\\':
			buf[pos] = '\\'
			buf[pos+1] = '\\'
			pos += 2
		default:
			buf[pos] = c
			pos++
		}
	}

	return buf[:pos]
}

// @todo: this method need hack for all SQL types
func appendSQLArgBool(buf []byte, v bool) []byte {
	if v {
		return append(buf, '1')
	}
	return append(buf, '0')
}

// escapeStringBackslash will escape string into the buffer, with backslash.
func escapeStringBackslash(buf []byte, v string) []byte {
	return escapeBytesBackslash(buf, Slice(v))
}

func appendSQLArgString(buf []byte, s string) []byte {
	buf = escapeStringBackslash(buf, s)
	return buf
}

func reflactAllTypesToString(any interface{}) (string, error) {
	buf := make([]byte, 0)
	switch v := any.(type) {
	case int:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int8:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int16:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int32:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int64:
		buf = strconv.AppendInt(buf, v, 10)
	case uint:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint8:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint16:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint32:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint64:
		buf = strconv.AppendUint(buf, v, 10)
	case float32:
		buf = strconv.AppendFloat(buf, float64(v), 'g', -1, 32)
	case float64:
		buf = strconv.AppendFloat(buf, v, 'g', -1, 64)
	case bool:
		buf = appendSQLArgBool(buf, v)
	case time.Time:
		if v.IsZero() {
			buf = append(buf, "'0000-00-00'"...)
		} else {
			buf = v.AppendFormat(buf, "2006-01-02 15:04:05.999999")
		}
	case json.RawMessage:
		buf = escapeBytesBackslash(buf, v)
	case []byte:
		if v == nil {
			buf = append(buf, "NULL"...)
		} else {
			buf = append(buf, "_binary'"...)
			buf = escapeBytesBackslash(buf, v)
		}
	case string:
		buf = appendSQLArgString(buf, escapeSQLString(v))
	case []string:
		for i, k := range v {
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = escapeStringBackslash(buf, escapeSQLString(k))
		}
	case []float32:
		for i, k := range v {
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = strconv.AppendFloat(buf, float64(k), 'g', -1, 32)
		}
	case []float64:
		for i, k := range v {
			if i > 0 {
				buf = append(buf, ',')
			}
			buf = strconv.AppendFloat(buf, k, 'g', -1, 64)
		}
	default:
		// slow path based on reflection
		reflectTp := reflect.TypeOf(any)
		kind := reflectTp.Kind()
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			buf = strconv.AppendInt(buf, reflect.ValueOf(any).Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			buf = strconv.AppendUint(buf, reflect.ValueOf(any).Uint(), 10)
		case reflect.Float32:
			buf = strconv.AppendFloat(buf, reflect.ValueOf(any).Float(), 'g', -1, 32)
		case reflect.Float64:
			buf = strconv.AppendFloat(buf, reflect.ValueOf(any).Float(), 'g', -1, 64)
		case reflect.Bool:
			buf = appendSQLArgBool(buf, reflect.ValueOf(any).Bool())
		case reflect.String:
			buf = appendSQLArgString(buf, escapeSQLString(reflect.ValueOf(any).String()))
		default:
			return "", errors.New(fmt.Sprintf("unsupported argument: %v", any))
		}
	}
	return string(buf), nil
}

func buildEscapedArgsLookupTable(args map[string]interface{}) (map[string]string, error) {
	escapedArgs := make(map[string]string, 0)
	for key, value := range args {
		valueInString, errInReflact := reflactAllTypesToString(value)
		if errInReflact != nil {
			return nil, errInReflact
		}
		escapedArgs[strings.TrimSpace(key)] = valueInString
	}
	return escapedArgs, nil
}

func EscapeIllaSQLActionTemplate(sql string, args map[string]interface{}) (string, error) {
	escapedArgs, errInBuildArgsLT := buildEscapedArgsLookupTable(args)
	if errInBuildArgsLT != nil {
		return "", errInBuildArgsLT
	}
	ret := ""
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
				ret += variableMappedValue
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
	return ret, nil
}
