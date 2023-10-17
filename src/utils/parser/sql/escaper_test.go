package parser_sql

import (
	"testing"

	"github.com/illacloud/builder-backend/src/utils/resourcelist"
	"github.com/stretchr/testify/assert"
)

func TestEscapeSQLActionTemplateTypePostgres(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name={{ !input1.value }};`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, escapedSQL, "SELECT * FROM actions where name=$1;", "the token should be equal")
}

func TestEscapeSQLActionTemplateTypeMySQL(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name={{ !input1.value }};`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_MYSQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, escapedSQL, "SELECT * FROM actions where name=?;", "the token should be equal")
}
