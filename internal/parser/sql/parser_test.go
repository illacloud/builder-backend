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

func TestIsSelectSQL(t *testing.T) {
	sql_1 := `
	/* select syntax for client query */
	SELECT * FROM tab1 where id=12;
	`
	lexer := NewLexer(sql_1)
	doesItIs := IsSelectSQL(lexer)

	assert.Equal(t, true, doesItIs, "it should be select query")

}

func TestIsSelectSQL2(t *testing.T) {
	sql_2 := `
	/* delete syntax for client query */
	DELETE FROM tab1 where id=12;
	`
	lexer := NewLexer(sql_2)
	doesItIs := IsSelectSQL(lexer)

	assert.Equal(t, false, doesItIs, "it should be delete query")

}
