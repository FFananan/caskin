package caskin_test

import (
	"testing"

	"github.com/awatercolorpen/caskin"
	"github.com/awatercolorpen/caskin/example"
	"github.com/stretchr/testify/assert"
)

func TestExecutorRole_GetRoles(t *testing.T) {
	stage, _ := newStage(t)
	provider := caskin.NewCachedProvider(nil, nil)
	executor := stage.Caskin.GetExecutor(provider)

	provider.Domain = stage.Domain
	provider.User = stage.AdminUser
	roles1, err := executor.GetRoles()
	assert.NoError(t, err)
	assert.Len(t, roles1, 2)

	provider.Domain = stage.Domain
	provider.User = stage.MemberUser
	roles4, err := executor.GetRoles()
	assert.NoError(t, err)
	assert.Len(t, roles4, 0)
}

func TestExecutorRole_GeneralCreate(t *testing.T) {
	stage, _ := newStage(t)
	provider := caskin.NewCachedProvider(nil, nil)
	executor := stage.Caskin.GetExecutor(provider)

	role1 := &example.Role{
		Name: "role_01",
	}
	assert.Equal(t, caskin.ErrProviderGet, executor.CreateRole(role1))

	provider.Domain = stage.Domain
	provider.User = stage.MemberUser
	assert.Equal(t, caskin.ErrNoWritePermission, executor.CreateRole(role1))
	provider.User = stage.AdminUser
	assert.Equal(t, caskin.ErrNoWritePermission, executor.CreateRole(role1))
	role1.ObjectID = 2
	provider.User = stage.MemberUser
	assert.Equal(t, caskin.ErrNoWritePermission, executor.CreateRole(role1))
	provider.User = stage.AdminUser
	assert.Equal(t, caskin.ErrEmptyID, executor.CreateRole(role1))
	role1.ParentID = 2
	assert.NoError(t, executor.CreateRole(role1))

	role2 := &example.Role{
		Name: "role_01",
	}
	assert.Equal(t, caskin.ErrAlreadyExists, executor.CreateRole(role2))
}

func TestExecutorRole_CreateSubNode(t *testing.T) {
	stage, _ := newStage(t)
	assert.NoError(t, stageAddSubAdmin(stage))
	provider := caskin.NewCachedProvider(nil, nil)
	executor := stage.Caskin.GetExecutor(provider)

	role1 := &example.Role{
		Name:     "role_sub_02",
		ObjectID: 5,
		ParentID: 3,
	}
	provider.Domain = stage.Domain
	provider.User = stage.MemberUser
	// member can not read or write object5
	assert.Equal(t, caskin.ErrNoWritePermission, executor.CreateRole(role1))

	// subadmin can read or write object5
	provider.User = stage.SubAdminUser
	assert.NoError(t, executor.CreateRole(role1))

	// 将当前的role移动到member下面，subadminuser没有权限
	role1.ParentID = 2
	assert.Equal(t, caskin.ErrNoWritePermission, executor.UpdateRole(role1))

	role2 := &example.Role{ID: 2}
	assert.Equal(t, caskin.ErrNoWritePermission, executor.DeleteRole(role2))
	role3 := &example.Role{ID: 4}
	assert.NoError(t, executor.DeleteRole(role3))

	//provider.User = stage.SuperadminUser
	//list3, err := executor.GetRoles()
	//assert.NoError(t, err)
	//bytes,_ := json.Marshal(list3)
	//fmt.Println(string(bytes))

	provider.User = stage.AdminUser
	// todo 这里有一个问题
	// parent为0的role能否被删除
	assert.NoError(t, executor.DeleteRole(role2))
	list1, err := executor.GetRoles()
	assert.NoError(t, err)
	assert.Len(t, list1, 3)
}

func TestExecutorRole_GeneralUpdate(t *testing.T) {
	stage, _ := newStage(t)
	provider := caskin.NewCachedProvider(nil, nil)
	assert.NoError(t, stageAddSubAdmin(stage))

	provider.User = stage.AdminUser
	provider.Domain = stage.Domain
	executor := stage.Caskin.GetExecutor(provider)

	role := &example.Role{
		ID:       3,
		Name:     "admin_sub_01_new_name",
		ParentID: 1,
		ObjectID: 5,
	}
	assert.NoError(t, executor.UpdateRole(role))

	role1 := &example.Role{}
	assert.Equal(t, caskin.ErrEmptyID, executor.UpdateRole(role1))

	// 这里的问题和上面的一样
	// 有没有权限操作paretn为0的role
	role2 := &example.Role{ID: 1, Name: "role_01_new_name", ObjectID: 1, ParentID: 0}
	assert.Equal(t, caskin.ErrCanNotOperateRootObjectWithoutSuperadmin, executor.UpdateRole(role2))

	provider.User = stage.MemberUser
	role4 := &example.Role{
		ID:       3,
		Name:     "admin_sub_01_new_name2",
		ParentID: 1,
		ObjectID: 5,
	}
	assert.Equal(t, caskin.ErrNoWritePermission, executor.UpdateRole(role4))
}

func TestExecutorRole_GeneralRecover(t *testing.T) {
	stage, _ := newStage(t)
	assert.NoError(t, stageAddSubAdmin(stage))
	provider := caskin.NewCachedProvider(nil, nil)
	provider.User = stage.AdminUser
	provider.Domain = stage.Domain
	executor := stage.Caskin.GetExecutor(provider)

	role := &example.Role{
		ID:       3,
		ParentID: 1,
	}
	assert.NoError(t, executor.DeleteRole(role))

	role1 := &example.Role{}
	assert.Equal(t, caskin.ErrAlreadyExists, executor.RecoverRole(role1))

	//provider.User = stage.SuperadminUser
	assert.NoError(t, executor.RecoverRole(role))
	assert.Equal(t, caskin.ErrAlreadyExists, executor.RecoverRole(role))

	role2 := &example.Role{ID: 3}
	assert.NoError(t, executor.DeleteRole(role2))
	provider.User = stage.MemberUser
	role2.ID = 3
	assert.Equal(t, caskin.ErrNoWritePermission, executor.RecoverRole(role2))
}

func TestExecutorRole_GeneralDelete(t *testing.T) {
	stage, _ := newStage(t)
	provider := caskin.NewCachedProvider(nil, nil)
	assert.NoError(t, stageAddSubAdmin(stage))

	provider.User = stage.SubAdminUser
	provider.Domain = stage.Domain
	executor := stage.Caskin.GetExecutor(provider)

	role := &example.Role{}
	assert.Equal(t, caskin.ErrEmptyID, executor.DeleteRole(role))

	role1 := &example.Role{ID: 2}
	assert.Equal(t, caskin.ErrNoWritePermission, executor.DeleteRole(role1))

	provider.User = stage.AdminUser
	assert.NoError(t, executor.DeleteRole(role1))

	role4 := &example.Role{ID: 1}
	assert.Equal(t, caskin.ErrCanNotOperateRootObjectWithoutSuperadmin, executor.DeleteRole(role4))

	role3 := &example.Role{ID: 4, ParentID: 1}
	assert.Equal(t, caskin.ErrNotExists, executor.RecoverRole(role3))
}
