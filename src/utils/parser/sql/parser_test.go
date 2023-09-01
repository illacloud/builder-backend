// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser_sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSample(t *testing.T) {
	assert.Nil(t, nil)
}

func TestIsSelectSQL1(t *testing.T) {
	sql_1 := `
	/* select syntax for client query */
	SELECT * FROM tab1 where id=14 and type=1;
	`
	lexer := NewLexer(sql_1)

	doesItIs, err := IsSelectSQL(lexer)
	assert.Nil(t, err)

	assert.Equal(t, true, doesItIs, "it should be a select query")
}

func TestIsSelectSQL2(t *testing.T) {
	sql_2 := `
	/* select syntax for client query */
	DELETE FROM tab1 where id=12;
	`
	lexer := NewLexer(sql_2)
	doesItIs, err := IsSelectSQL(lexer)
	assert.Nil(t, err)

	assert.Equal(t, false, doesItIs, "it should be a delete query")
}

func TestIsSelectSQL3(t *testing.T) {
	sql_3 := `
	/* update syntax for client query */
	update staff set Name = 'james@illasoft.com' where Email = 'wei'
	`
	lexer := NewLexer(sql_3)
	doesItIs, err := IsSelectSQL(lexer)
	assert.Nil(t, err)

	assert.Equal(t, false, doesItIs, "it should be a update query")
}

func TestIsSelectSQL4(t *testing.T) {
	sql_4 := `
	/* insert syntax for client query */
	Insert into data ( Name, id ) VALUES ('james@illasoft.com', 1219)
	`
	lexer := NewLexer(sql_4)
	doesItIs, err := IsSelectSQL(lexer)
	assert.Nil(t, err)

	assert.Equal(t, false, doesItIs, "it should be an insert query")
}

func TestIsSelectSQL5(t *testing.T) {
	sql_5 := `
	create table if not exists users
(
    id                       bigserial                         not null
        primary key,
    nickname                 varchar(15)                       not null, /* 3-15 character */
    password_digest          varchar(60)                       not null,
    email                    varchar(255)                      not null,
    language                 smallint                          not null,
    is_subscribed            boolean default false             not null,
    created_at               timestamp                         not null,
    updated_at               timestamp                         not null
);
	`
	lexer := NewLexer(sql_5)
	doesItIs, err := IsSelectSQL(lexer)
	assert.Nil(t, err)

	assert.Equal(t, false, doesItIs, "it should be a DDL query")
}

func TestIsSelectSQL6(t *testing.T) {
	sql_6 := `
ALTER TABLE set_states
  DROP CONSTRAINT IF EXISTS set_states_displayname_constrainte
, ADD CONSTRAINT set_states_displayname_constrainte UNIQUE (version, app_ref_id, value);
	`
	lexer := NewLexer(sql_6)
	doesItIs, err := IsSelectSQL(lexer)
	assert.Nil(t, err)

	assert.Equal(t, false, doesItIs, "it should be a DDL query")
}

func TestIsSelectSQL7(t *testing.T) {
	sql_7 := `
	INSERT INTO tb_courses_new
    (course_id,course_name,course_grade,course_info)
    SELECT course_id,course_name,course_grade,course_info
    FROM tb_courses;
	`
	lexer := NewLexer(sql_7)
	doesItIs, err := IsSelectSQL(lexer)
	assert.Nil(t, err)

	assert.Equal(t, false, doesItIs, "it should be an insert query")
}
