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
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name=$1;", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateTypeMySQL(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name={{ !input1.value }};`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_MYSQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name=?;", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteStringTemplatePostgres(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%{{ !input1.value }}.{{input2.value}} sir%' or name like '%{{input3}}%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan", "222 pan", "333 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%', '$1'::text, '.', '$2'::text, ' sir%') or name like CONCAT('%', '$3'::text, '%');", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteStringTemplateMySQL(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%{{ !input1.value }}.{{input2.value}} sir%' or name like '%{{input3}}%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_MYSQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan", "222 pan", "333 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%', '?', '.', '?', ' sir%') or name like CONCAT('%', '?', '%');", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteEscape(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%''{{ !input1.value }}''%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%''', '$1'::text, '''%');", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteEscapeSlash(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%\'{{ !input1.value }}\'%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%\\'', '$1'::text, '\\'%');", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateDoubleQuoteContainSingleQuote(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like "%'{{ !input1.value }}'%";`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT(\"%'\", \"$1\"::text, \"'%\");", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleQuoteContainDoubleQuote(t *testing.T) {
	sql_1 := `SELECT * FROM actions where name like '%"{{ !input1.value }}"%';`
	args := map[string]interface{}{
		" !input1.value ": "122 pan",
		"input2.value":    "222 pan",
		"input3":          "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%\"', '$1'::text, '\"%');", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateMixVariable(t *testing.T) {
	sql_1 := `select *  from users join orders on users.id = orders.id where {{!input1.value}} or lower(users.name) like '%{{input1.value.toLowerCase()}}%'`
	args := map[string]interface{}{
		"input1.value.toLowerCase()": "122 pan",
		"!input1.value":              "222 pan",
		"input3":                     "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"222 pan", "122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "select *  from users join orders on users.id = orders.id where $1 or lower(users.name) like CONCAT('%', '$2'::text, '%')", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateMixVariableMissingParam(t *testing.T) {
	sql_1 := `select *  from users join orders on users.id = orders.id where {{!input1.value}} or lower(users.name) like '%{{input1.value.toLowerCase()}}%'`
	args := map[string]interface{}{
		"!input1.value": "222 pan",
		"input3":        "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"222 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "select *  from users join orders on users.id = orders.id where $1 or lower(users.name) like CONCAT('%', ''::text, '%')", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateDoubleQuotePostgresSchemaName(t *testing.T) {
	sql_1 := `select *  from "usersInfoTable"`
	args := map[string]interface{}{
		"!input1.value": "222 pan",
		"input3":        "333 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{}, "the usedArgs should be equal")
	assert.Equal(t, "select *  from \"usersInfoTable\"", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleValue(t *testing.T) {
	sql_1 := `select *  from users where created_at >= '{{input1.value}}'`
	args := map[string]interface{}{
		"input1.value": "122 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_ORACLE_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "select *  from users where created_at >= ':1'", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleValueWithoutQuote(t *testing.T) {
	sql_1 := `select * from users where created_at >= {{input1.value}}`
	args := map[string]interface{}{
		"input1.value": "122 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_ORACLE_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "select * from users where created_at >= :1", escapedSQL, "the token should be equal")
}

func TestEscapeSQLWithChinese(t *testing.T) {
	sql_1 := `select id as '序列号' from users where created_at >= {{input1.value}}`
	args := map[string]interface{}{
		"input1.value": "122 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_ORACLE_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, usedArgs, []interface{}{"122 pan"}, "the usedArgs should be equal")
	assert.Equal(t, "select id as '序列号' from users where created_at >= :1", escapedSQL, "the token should be equal")
}
