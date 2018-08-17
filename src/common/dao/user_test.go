// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/fske/harbor/src/common/models"
)

func TestDeleteUser(t *testing.T) {
	username := "user_for_test"
	email := "user_for_test@vmware.com"
	password := "P@ssword"
	realname := "user_for_test"

	u := models.User{
		Username: username,
		Email:    email,
		Password: password,
		Realname: realname,
	}
	id, err := Register(u)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	defer func(id int64) {
		if err := CleanUser(id); err != nil {
			t.Fatalf("failed to delete user %d: %v", id, err)
		}
	}(id)

	err = DeleteUser(int(id))
	if err != nil {
		t.Fatalf("Error occurred in DeleteUser: %v", err)
	}

	user := &models.User{}
	sql := "select * from user where user_id = ?"
	if err = GetOrmer().Raw(sql, id).
		QueryRow(user); err != nil {
		t.Fatalf("failed to query user: %v", err)
	}

	if user.Deleted != 1 {
		t.Error("user is not deleted")
	}

	expected := fmt.Sprintf("%s#%d", u.Username, id)
	if user.Username != expected {
		t.Errorf("unexpected username: %s != %s", user.Username,
			expected)
	}

	expected = fmt.Sprintf("%s#%d", u.Email, id)
	if user.Email != expected {
		t.Errorf("unexpected email: %s != %s", user.Email,
			expected)
	}
}

func TestOnBoardUser(t *testing.T) {
	assert := assert.New(t)
	u := &models.User{
		Username: "user1",
		Password: "password1",
		Email:    "dummy@placehodler.com",
		Realname: "daniel",
	}
	err := OnBoardUser(u)
	assert.Nil(err)
	id := u.UserID
	assert.True(id > 0)
	err = OnBoardUser(u)
	assert.Nil(err)
	assert.True(u.UserID == id)
	CleanUser(int64(id))
}
