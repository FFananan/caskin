package caskin

import "fmt"

var (
	// errors about entry
	ErrNil           = fmt.Errorf("nil data")
	ErrEmptyID       = fmt.Errorf("empty id")
	ErrAlreadyExists = fmt.Errorf("already exists")
	ErrNotExists     = fmt.Errorf("not exists")
	// errors about permission
	ErrNoReadPermission  = fmt.Errorf("no read permission")
	ErrNoWritePermission = fmt.Errorf("no write permission")
	// errors about superadmin
	ErrIsNotSuperAdmin       = fmt.Errorf("is not superadmin")
	ErrSuperAdminIsNoEnabled = fmt.Errorf("superadmin is not enabled")
	// errors about caskin initialization
	ErrInitializationNilDomainCreator = fmt.Errorf("domain creator is nil")
	ErrInitializationNilEnforcer      = fmt.Errorf("enforcer is nil")
	ErrInitializationNilEntryFactory  = fmt.Errorf("entry factory is nil")
	ErrInitializationNilMetaDB        = fmt.Errorf("metadata database is nil")
	// errors about current provider
	ErrProviderGet = fmt.Errorf("provider can't get current status")
	// errors about user role pair
	ErrInputArrayNotBelongSameUser = fmt.Errorf("input user role pair array are not belong to same user")
	ErrInputArrayNotBelongSameRole  = fmt.Errorf("input user role pair array are not belong to same role")
)

