/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateRandomString(t *testing.T) {
	_, err := generateRandomString(0)
	assert.Error(t, err)

	_, err = generateRandomString(-1)
	assert.Error(t, err)

	for i := 1; i < 100; i++ {
		s, err := generateRandomString(i)
		assert.NoError(t, err)
		assert.Equal(t, i, len(s))
	}

	for i := 0; i < 1000; i++ {
		one, err := generateRandomString(32)
		assert.NoError(t, err)
		two, err := generateRandomString(32)
		assert.NoError(t, err)
		assert.NotEqual(t, one, two)
	}
}
