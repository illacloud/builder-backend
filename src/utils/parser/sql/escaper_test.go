package parser_sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeIllaSQLActionTemplate(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name={{ !input1.value }};`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
	}
	escapedSQL, errInEscape := EscapeIllaSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, escapedSQL, "", "the token should be equal")
}
