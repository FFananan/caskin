package caskin

type ObjectType string

type Action string

const (
	Read  Action = "read"
	Write Action = "write"
)

type Policy struct {
	Role   Role
	Object Object
	Domain Domain
	Action Action
}

type RolesForUser struct {
	User  User   `json:"user"`
	Roles []Role `json:"roles"`
}

type UsersForRole struct {
	Role  Role   `json:"role"`
	Users []User `json:"users"`
}

type entry interface {
	// get id method
	GetID() uint64
	// get id method
	SetID(uint64)
	// encode entry to string method
	Encode() string
	// decode string to entry method
	Decode(string) error
	// is object method
	IsObject() bool
	// get object string method
	GetObject() string
}

type parent interface {
	// get parent id method
	GetParentID() uint64
	// set parent id method
	SetParentID(uint64)
}

type inDomain interface {
	// set domain id method
	SetDomainID(uint64)
}

type parentEntry interface {
	entry
	parent
}

type User interface {
	entry
}

type Role interface {
	parentEntry
	inDomain
}

type Object interface {
	parentEntry
	inDomain
	GetObjectType() ObjectType
}

type Domain interface {
	entry
}

type EntryFactory interface {
	NewUser() User
	NewRole() Role
	NewObject() Object
	NewDomain() Domain
}
