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

func TestEscapeSQLActionTemplateSingleQuoteStringTemplatePostgres(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%{{ !input1.value }}.{{input2.value}} sir%' or name like '%{{input3}}%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan", "222 pan", "333 pan"}, "the usedArgs should be equal")
	assert.Equal(t, escapedSQL, "SELECT * FROM actions where name like CONCAT('%', $1, '.', $2, ' sir%') or name like CONCAT('%', $3, '%');", "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteStringTemplateMySQL(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%{{ !input1.value }}.{{input2.value}} sir%' or name like '%{{input3}}%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_MYSQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan", "222 pan", "333 pan"}, "the usedArgs should be equal")
	assert.Equal(t, escapedSQL, "SELECT * FROM actions where name like CONCAT('%', ?, '.', ?, ' sir%') or name like CONCAT('%', ?, '%');", "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteEscape(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%''{{ !input1.value }}''%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, escapedSQL, "SELECT * FROM actions where name like CONCAT('%''', $1, '''%');", "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteEscapeSlash(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%\'{{ !input1.value }}\'%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, escapedSQL, "SELECT * FROM actions where name like CONCAT('%\\'', $1, '\\'%');", "the token should be equal")
}

func TestEscapeSQLActionTemplateDoubleQuoteContainSingleQuote(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like "%'{{ !input1.value }}'%";`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, escapedSQL, "SELECT * FROM actions where name like CONCAT(\"%'\", $1, \"'%\");", "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteContainDoubleQuote(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%"{{ !input1.value }}"%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, escapedSQL, "SELECT * FROM actions where name like CONCAT('%\"', $1, '\"%');", "the token should be equal")
}