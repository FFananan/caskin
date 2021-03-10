package caskin_test

import (
	"github.com/awatercolorpen/caskin"
	"github.com/awatercolorpen/caskin/example"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/gorm-adapter/v3"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
	"testing"
)

func TestNewCaskin(t *testing.T) {
	_, err := newCaskin(t, &caskin.Options{})
	assert.NoError(t, err)
}

func TestCaskin_GetExecutor(t *testing.T) {
	_, err := newStage(t)
	assert.NoError(t, err)
}

func getTestDB(tb testing.TB) (*gorm.DB, error) {
	dsn := filepath.Join(tb.TempDir(), "sqlite")
	//dsn := filepath.Join("./", "sqlite")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
	//return db.Debug(), nil
}

var casbinModelMap = map[bool]model.Model{}

func getCasbinModel(options *caskin.Options) (model.Model, error) {
	k := options.IsEnableSuperAdmin()
	if _, ok := casbinModelMap[k]; !ok {
		m, err := caskin.CasbinModel(options)
		if err != nil {
			return nil, err
		}
		casbinModelMap[k] = m
	}

	return casbinModelMap[k], nil
}

func newCaskin(tb testing.TB, options *caskin.Options) (*caskin.Caskin, error) {
	db, err := getTestDB(tb)
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&example.User{},
		&example.Domain{},
		&example.Role{},
		&example.Object{})
	if err != nil {
		return nil, err
	}

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	m, err := getCasbinModel(options)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	return caskin.New(options,
		caskin.DomainCreatorOption(example.NewDomainCreator),
		caskin.EnforcerOption(enforcer),
		caskin.EntryFactoryOption(&example.EntryFactory{}),
		caskin.MetaDBOption(example.NewGormMDBByDB(db)),
	)
}

func newStage(t *testing.T) (*example.Stage, error) {
	options := &caskin.Options{
		SuperadminOption: &caskin.SuperadminOption{
			Enable: true,
		},
	}
	c, err := newCaskin(t, options)
	if err != nil {
		return nil, err
	}

	provider := &example.Provider{}
	executor := c.GetExecutor(provider)

	domain := &example.Domain{Name: "domain_01"}
	if err := executor.CreateDomain(domain); err != nil {
		return nil, err
	}
	if err := executor.ReInitializeDomain(domain); err != nil {
		return nil, err
	}

	superadmin := &example.User{
		PhoneNumber: "12345678901",
		Email:       "superadmin@qq.com",
	}
	admin := &example.User{
		PhoneNumber: "12345678902",
		Email:       "admin@qq.com",
	}
	member := &example.User{
		PhoneNumber: "12345678903",
		Email:       "member@qq.com",
	}
	for _, v := range []caskin.User{superadmin, admin, member} {
		if err := executor.CreateUser(v); err != nil {
			return nil, err
		}
	}

	if err := executor.AddSuperadminUser(superadmin); err != nil {
		return nil, err
	}

	provider.Domain = domain
	provider.User = superadmin
	roles, err := executor.GetRoles()

	for k, v := range map[caskin.Role][]*caskin.UserRolePair{
		roles[0]: {{User: admin, Role: roles[0]}},
		roles[1]: {{User: member, Role: roles[1]}},
	} {
		if err := executor.ModifyUserRolePairPerRole(k, v); err != nil {
			return nil, err
		}
	}

	stage := &example.Stage{
		Caskin:         c,
		Options:        options,
		Domain:         domain,
		SuperadminUser: superadmin,
		AdminUser:      admin,
		MemberUser:     member,
	}

	return stage, nil
}
