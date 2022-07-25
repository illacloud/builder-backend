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

package util

// remove target element from a slice
// NOTE: this method remove first target element in slice only
func DeleteElement(s []int, e int) []int {
	if len(s) == 0 {
		return s
	}
	pos := 0
	for k, v := range s {
		if v == e {
			pos = k
			break
		}
	}
	return append(s[:pos], s[pos+1:]...)
}
