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
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
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
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
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
	assert.Equal(t, []interface{}{"122 pan", "222 pan", "333 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%', $1::text, '.', $2::text, ' sir%') or name like CONCAT('%', $3::text, '%');", escapedSQL, "the token should be equal")
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
	assert.Equal(t, []interface{}{"122 pan", "222 pan", "333 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%', ?, '.', ?, ' sir%') or name like CONCAT('%', ?, '%');", escapedSQL, "the token should be equal")
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
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%''', $1::text, '''%');", escapedSQL, "the token should be equal")
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
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%\\'', $1::text, '\\'%');", escapedSQL, "the token should be equal")
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
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT(\"%'\", $1::text, \"'%\");", escapedSQL, "the token should be equal")
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
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "SELECT * FROM actions where name like CONCAT('%\"', $1::text, '\"%');", escapedSQL, "the token should be equal")
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
	assert.Equal(t, []interface{}{"222 pan", "122 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select *  from users join orders on users.id = orders.id where $1 or lower(users.name) like CONCAT('%', $2::text, '%')", escapedSQL, "the token should be equal")
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
	assert.Equal(t, []interface{}{"222 pan"}, usedArgs, "the usedArgs should be equal")
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
	assert.Equal(t, []interface{}{}, usedArgs, "the usedArgs should be equal")
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
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select *  from users where created_at >= :1", escapedSQL, "the token should be equal")
}

func TestEscapeSQLActionTemplateSingleValueWithoutQuote(t *testing.T) {
	sql_1 := `select * from users where created_at >= {{input1.value}}`
	args := map[string]interface{}{
		"input1.value": "122 pan",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_ORACLE_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
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
	assert.Equal(t, []interface{}{"122 pan"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select id as '序列号' from users where created_at >= :1", escapedSQL, "the token should be equal")
}

func TestEscapePostgresSQLWithDate(t *testing.T) {
	sql_1 := `select count(distinct email) from users where DATE_TRUNC('day', created_at) >= '{{date1.value}}'::timestamp and DATE_TRUNC('day', created_at) <= '{{date2.value}}'::timestamp and email like '%.edu.%'`
	args := map[string]interface{}{
		"date1.value": "2000-01-01",
		"date2.value": "2020-01-01",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{"2000-01-01", "2020-01-01"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select count(distinct email) from users where DATE_TRUNC('day', created_at) >= $1::timestamp and DATE_TRUNC('day', created_at) <= $2::timestamp and email like '%.edu.%'", escapedSQL, "the token should be equal")
}

func TestEscapePostgresSQLWithDateUnsafe(t *testing.T) {
	sql_1 := `select count(distinct email) from users where DATE_TRUNC('day', created_at) >= '{{date1.value}}'::timestamp and DATE_TRUNC('day', created_at) <= '{{date2.value}}'::timestamp and email like '%.edu.%'`
	args := map[string]interface{}{
		"date1.value": "2000-01-01",
		"date2.value": "2020-01-01",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, _, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, false)
	assert.Nil(t, errInEscape)
	assert.Equal(t, "select count(distinct email) from users where DATE_TRUNC('day', created_at) >= '2000-01-01'::timestamp and DATE_TRUNC('day', created_at) <= '2020-01-01'::timestamp and email like '%.edu.%'", escapedSQL, "the token should be equal")
}

func TestEscapePostgresSQLOnboarding(t *testing.T) {
	sql_1 := `select * from users join orders on users.id = orders.id where {{!input1.value}} or lower(users.name) like '%{{input1.value.toLowerCase()}}%'`
	args := map[string]interface{}{
		"!input1.value":              "true",
		"input1.value.toLowerCase()": "james",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, _, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, false)
	assert.Nil(t, errInEscape)
	assert.Equal(t, "select * from users join orders on users.id = orders.id where true or lower(users.name) like CONCAT('%', 'james'::text, '%')", escapedSQL, "the token should be equal")
}

func TestEscapePostgresSQLIssue3463(t *testing.T) {
	sql_1 := `select count(1) as CNT from TMP_OPTION_CLOSE where TASK_ID like 'EO_MID_{{date2.value}}%'`
	args := map[string]interface{}{
		"date2.value": "11",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_ORACLE_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{"11"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select count(1) as CNT from TMP_OPTION_CLOSE where TASK_ID like CONCAT('EO_MID_', :1, '%')", escapedSQL, "the token should be equal")
}

func TestEscapeMySQLSQLInStatementQueryInIntType(t *testing.T) {
	sql_1 := `select * from users where id in ({{multiselect1.value.map(b => Number(b))}})`
	args := map[string]interface{}{
		`multiselect1.value.map(b => Number(b))`: []int{1, 2, 3},
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_MYSQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{"'1', '2', '3'"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select * from users where id in (?)", escapedSQL, "the token should be equal")
}

func TestEscapeMySQLSQLInStatementQueryInIntString(t *testing.T) {
	sql_1 := `select * from users where id in ({{multiselect1.value.map(b => Number(b))}})`
	args := map[string]interface{}{
		`multiselect1.value.map(b => Number(b))`: []interface{}{"a", "b", "c"},
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_MYSQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{"a", "b", "c"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select * from users where id in (?, ?, ?)", escapedSQL, "the token should be equal")
}

func TestEscapePostgresSQLInStatementQueryInIntString(t *testing.T) {
	sql_1 := `select * from users where id in ({{multiselect1.value.map(b => Number(b))}})`
	args := map[string]interface{}{
		`multiselect1.value.map(b => Number(b))`: []interface{}{"a", "b", "c"},
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{"a", "b", "c"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select * from users where id in ($1, $2, $3)", escapedSQL, "the token should be equal")
}

func TestEscapeMySQLSQLInStatementQueryInIntStringInUnsafeMode(t *testing.T) {
	sql_1 := `select * from users where id in ({{multiselect1.value.map(b => Number(b))}})`
	args := map[string]interface{}{
		`multiselect1.value.map(b => Number(b))`: []interface{}{"a", "b", "c"},
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_MYSQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, false)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select * from users where id in ('a', 'b', 'c')", escapedSQL, "the token should be equal")
}

// The Postgres ANY case are in:
// ```sql
// select * from apps where id = ANY('{1,2,3}');
// select * from apps where id = ANY(ARRAY[1,2]);
// SELECT * FROM apps WHERE uid = ANY (VALUES ('ca5e3145-f9b4-4610-bd25-0ffbf258cce7'::uuid), ('feb398fa-e5eb-43f6-8488-82a9f4806570'::uuid));
// ```
func TestEscapePostgresSQLAnyStatementQuery(t *testing.T) {
	sql_1 := `select * from users where name = ANY(ARRAY[{{multiselect1.value.map(b => Number(b))}}])`
	args := map[string]interface{}{
		`multiselect1.value.map(b => Number(b))`: []interface{}{"a", "b", "c"},
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{"a", "b", "c"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, "select * from users where name = ANY(ARRAY[$1, $2, $3])", escapedSQL, "the token should be equal")
}

func TestEscapePostgresSQLInvaliedLengthUTF8Case(t *testing.T) {
	sql_1 := `SELECT * FROM "库存数据视图" WHERE "产品名称" LIKE '%{{input1.value}}%';   `
	args := map[string]interface{}{
		`input1.value`: "value_1",
	}
	sqlEscaper := NewSQLEscaper(resourcelist.TYPE_POSTGRESQL_ID)
	escapedSQL, usedArgs, errInEscape := sqlEscaper.EscapeSQLActionTemplate(sql_1, args, true)
	assert.Nil(t, errInEscape)
	assert.Equal(t, []interface{}{"value_1"}, usedArgs, "the usedArgs should be equal")
	assert.Equal(t, `SELECT * FROM "库存数据视图" WHERE "产品名称" LIKE CONCAT('%', $1::text, '%');   `, escapedSQL, "the token should be equal")
}
