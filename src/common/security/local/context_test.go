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

package local

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/fske/harbor/src/common"
	"github.com/fske/harbor/src/common/dao"
	"github.com/fske/harbor/src/common/dao/project"
	"github.com/fske/harbor/src/common/models"
	"github.com/fske/harbor/src/common/utils/log"
	"github.com/fske/harbor/src/ui/promgr"
	"github.com/fske/harbor/src/ui/promgr/pmsdriver/local"
)

var (
	private = &models.Project{
		Name:    "private_project",
		OwnerID: 1,
	}

	projectAdminUser = &models.User{
		Username: "projectAdminUser",
		Email:    "projectAdminUser@vmware.com",
	}
	developerUser = &models.User{
		Username: "developerUser",
		Email:    "developerUser@vmware.com",
	}
	guestUser = &models.User{
		Username: "guestUser",
		Email:    "guestUser@vmware.com",
	}

	pm = promgr.NewDefaultProjectManager(local.NewDriver(), true)
)

func TestMain(m *testing.M) {
	dbHost := os.Getenv("POSTGRESQL_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable POSTGRES_HOST is not set")
	}
	dbUser := os.Getenv("POSTGRESQL_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable POSTGRES_USR is not set")
	}
	dbPortStr := os.Getenv("POSTGRESQL_PORT")
	if len(dbPortStr) == 0 {
		log.Fatalf("environment variable POSTGRES_PORT is not set")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("invalid POSTGRESQL_PORT: %v", err)
	}

	dbPassword := os.Getenv("POSTGRESQL_PWD")
	dbDatabase := os.Getenv("POSTGRESQL_DATABASE")
	if len(dbDatabase) == 0 {
		log.Fatalf("environment variable POSTGRESQL_DATABASE is not set")
	}

	database := &models.Database{
		Type: "postgresql",
		PostGreSQL: &models.PostGreSQL{
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPassword,
			Database: dbDatabase,
		},
	}

	log.Infof("POSTGRES_HOST: %s, POSTGRES_USR: %s, POSTGRES_PORT: %d, POSTGRES_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// regiser users
	id, err := dao.Register(*projectAdminUser)
	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}
	projectAdminUser.UserID = int(id)
	defer dao.DeleteUser(int(id))

	id, err = dao.Register(*developerUser)
	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}
	developerUser.UserID = int(id)
	defer dao.DeleteUser(int(id))

	id, err = dao.Register(*guestUser)
	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}
	guestUser.UserID = int(id)
	defer dao.DeleteUser(int(id))

	// add project
	id, err = dao.AddProject(*private)
	if err != nil {
		log.Fatalf("failed to add project: %v", err)
	}
	private.ProjectID = id
	defer dao.DeleteProject(id)

	var projectAdminPMID, developerUserPMID, guestUserPMID int
	// add project members
	projectAdminPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  private.ProjectID,
		EntityID:   projectAdminUser.UserID,
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	})
	if err != nil {
		log.Fatalf("failed to add member: %v", err)
	}
	defer project.DeleteProjectMemberByID(projectAdminPMID)

	developerUserPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  private.ProjectID,
		EntityID:   developerUser.UserID,
		EntityType: common.UserMember,
		Role:       common.RoleDeveloper,
	})
	if err != nil {
		log.Fatalf("failed to add member: %v", err)
	}
	defer project.DeleteProjectMemberByID(developerUserPMID)
	guestUserPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  private.ProjectID,
		EntityID:   guestUser.UserID,
		EntityType: common.UserMember,
		Role:       common.RoleGuest,
	})
	if err != nil {
		log.Fatalf("failed to add member: %v", err)
	}
	defer project.DeleteProjectMemberByID(guestUserPMID)
	os.Exit(m.Run())
}

func TestIsAuthenticated(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsAuthenticated())

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.True(t, ctx.IsAuthenticated())
}

func TestGetUsername(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.Equal(t, "test", ctx.GetUsername())
}

func TestIsSysAdmin(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, non admin
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, admin
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: true,
	}, nil)
	assert.True(t, ctx.IsSysAdmin())
}

func TestIsSolutionUser(t *testing.T) {
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsSolutionUser())
}

func TestHasReadPerm(t *testing.T) {
	// public project
	ctx := NewSecurityContext(nil, pm)
	assert.True(t, ctx.HasReadPerm("library"))

	// private project, unauthenticated
	ctx = NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasReadPerm(private.Name))

	// private project, authenticated, has no perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasReadPerm(private.Name))

	// private project, authenticated, has read perm
	ctx = NewSecurityContext(guestUser, pm)
	assert.True(t, ctx.HasReadPerm(private.Name))

	// private project, authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		HasAdminRole: true,
	}, pm)
	assert.True(t, ctx.HasReadPerm(private.Name))
}

func TestHasWritePerm(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasWritePerm(private.Name))

	// authenticated, has read perm
	ctx = NewSecurityContext(guestUser, pm)
	assert.False(t, ctx.HasWritePerm(private.Name))

	// authenticated, has write perm
	ctx = NewSecurityContext(developerUser, pm)
	assert.True(t, ctx.HasWritePerm(private.Name))

	// authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		HasAdminRole: true,
	}, pm)
	assert.True(t, ctx.HasReadPerm(private.Name))
}

func TestHasAllPerm(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasAllPerm(private.Name))

	// authenticated, has all perms
	ctx = NewSecurityContext(projectAdminUser, pm)
	assert.True(t, ctx.HasAllPerm(private.Name))

	// authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		HasAdminRole: true,
	}, pm)
	assert.True(t, ctx.HasAllPerm(private.Name))
}

func TestHasAllPermWithGroup(t *testing.T) {
	PrepareGroupTest()
	project, err := dao.GetProjectByName("group_project")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	developer, err := dao.GetUser(models.User{Username: "sample01"})
	if err != nil {
		t.Errorf("Error occurred when GetUser: %v", err)
	}
	developer.GroupList = []*models.UserGroup{
		&models.UserGroup{GroupName: "test_group", GroupType: 1, LdapGroupDN: "cn=harbor_user,dc=example,dc=com"},
	}
	ctx := NewSecurityContext(developer, pm)
	assert.False(t, ctx.HasAllPerm(project.Name))
	assert.True(t, ctx.HasWritePerm(project.Name))
	assert.True(t, ctx.HasReadPerm(project.Name))
}

func TestGetMyProjects(t *testing.T) {
	ctx := NewSecurityContext(guestUser, pm)
	projects, err := ctx.GetMyProjects()
	require.Nil(t, err)
	assert.Equal(t, 1, len(projects))
	assert.Equal(t, private.ProjectID, projects[0].ProjectID)
}

func TestGetProjectRoles(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	roles := ctx.GetProjectRoles(private.Name)
	assert.Equal(t, 0, len(roles))

	// authenticated, project name of ID is nil
	ctx = NewSecurityContext(guestUser, pm)
	roles = ctx.GetProjectRoles(nil)
	assert.Equal(t, 0, len(roles))

	// authenticated, has read perm
	ctx = NewSecurityContext(guestUser, pm)
	roles = ctx.GetProjectRoles(private.Name)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, common.RoleGuest, roles[0])

	// authenticated, has write perm
	ctx = NewSecurityContext(developerUser, pm)
	roles = ctx.GetProjectRoles(private.Name)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, common.RoleDeveloper, roles[0])

	// authenticated, has all perms
	ctx = NewSecurityContext(projectAdminUser, pm)
	roles = ctx.GetProjectRoles(private.Name)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, common.RoleProjectAdmin, roles[0])
}
func PrepareGroupTest() {
	initSqls := []string{
		`insert into user_group (group_name, group_type, ldap_group_dn) values ('harbor_group_01', 1, 'cn=harbor_user,dc=example,dc=com')`,
		`insert into harbor_user (username, email, password, realname) values ('sample01', 'sample01@example.com', 'harbor12345', 'sample01')`,
		`insert into project (name, owner_id) values ('group_project', 1)`,
		`insert into project (name, owner_id) values ('group_project_private', 1)`,
		`insert into project_metadata (project_id, name, value) values ((select project_id from project where name = 'group_project'), 'public', 'false')`,
		`insert into project_metadata (project_id, name, value) values ((select project_id from project where name = 'group_project_private'), 'public', 'false')`,
		`insert into project_member (project_id, entity_id, entity_type, role) values ((select project_id from project where name = 'group_project'), (select id from user_group where group_name = 'harbor_group_01'),'g', 2)`,
	}

	clearSqls := []string{
		`delete from project_metadata where project_id in (select project_id from project where name in ('group_project', 'group_project_private'))`,
		`delete from project where name in ('group_project', 'group_project_private')`,
		`delete from project_member where project_id in (select project_id from project where name in ('group_project', 'group_project_private'))`,
		`delete from user_group where group_name = 'harbor_group_01'`,
		`delete from harbor_user where username = 'sample01'`,
	}
	dao.PrepareTestData(clearSqls, initSqls)
}

func TestSecurityContext_GetRolesByGroup(t *testing.T) {
	PrepareGroupTest()
	project, err := dao.GetProjectByName("group_project")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	developer, err := dao.GetUser(models.User{Username: "sample01"})
	if err != nil {
		t.Errorf("Error occurred when GetUser: %v", err)
	}
	developer.GroupList = []*models.UserGroup{
		&models.UserGroup{GroupName: "test_group", GroupType: 1, LdapGroupDN: "cn=harbor_user,dc=example,dc=com"},
	}
	type fields struct {
		user *models.User
		pm   promgr.ProjectManager
	}
	type args struct {
		projectIDOrName interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []int
	}{
		{"Developer", fields{user: developer, pm: pm}, args{project.ProjectID}, []int{2}},
		{"Guest", fields{user: guestUser, pm: pm}, args{project.ProjectID}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SecurityContext{
				user: tt.fields.user,
				pm:   tt.fields.pm,
			}
			if got := s.GetRolesByGroup(tt.args.projectIDOrName); !dao.ArrayEqual(got, tt.want) {
				t.Errorf("SecurityContext.GetRolesByGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecurityContext_GetMyProjects(t *testing.T) {
	type fields struct {
		user *models.User
		pm   promgr.ProjectManager
	}
	tests := []struct {
		name     string
		fields   fields
		wantSize int
		wantErr  bool
	}{
		{"Admin", fields{user: projectAdminUser, pm: pm}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SecurityContext{
				user: tt.fields.user,
				pm:   tt.fields.pm,
			}
			got, err := s.GetMyProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("SecurityContext.GetMyProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantSize {
				t.Errorf("SecurityContext.GetMyProjects() = %v, want %v", len(got), tt.wantSize)
			}
		})
	}
}
