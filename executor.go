package caskin

import "github.com/ahmetb/go-linq/v3"

type Executor struct {
	Enforcer IEnforcer
	DB       MetaDB
	provider CurrentProvider
	factory  EntryFactory
	options  *Options
}

func (e *Executor) GetCurrentProvider() CurrentProvider {
	return e.provider
}

func (e *Executor) newObject() treeNodeEntry {
	return e.factory.NewObject()
}

func (e *Executor) newRole() treeNodeEntry {
	return e.factory.NewRole()
}

func (e *Executor) objectParentUpdater() *parentEntryUpdater {
	return &parentEntryUpdater{
		newEntry:    e.newObject,
		parentGetFn: e.objectParentsFn(),
		parentAddFn: func(p1 treeNodeEntry, p2 treeNodeEntry, domain Domain) error {
			return e.Enforcer.AddParentForObjectInDomain(p1.(Object), p2.(Object), domain)
		},
		parentDelFn: func(p1 treeNodeEntry, p2 treeNodeEntry, domain Domain) error {
			return e.Enforcer.RemoveParentForObjectInDomain(p1.(Object), p2.(Object), domain)
		},
	}
}

func (e *Executor) objectDeleteFn() deleteFn {
	return func(p treeNodeEntry, d Domain) error {
		if err := e.Enforcer.RemoveObjectInDomain(p.(Object), d); err != nil {
			return err
		}
		return e.DB.DeleteByID(p, p.GetID())
	}
}

func (e *Executor) objectChildrenFn() childrenFn {
	return e.childrenOrParentGetFn(func(p treeNodeEntry, domain Domain) interface{} {
		return e.Enforcer.GetChildrenForObjectInDomain(p.(Object), domain)
	})
}

func (e *Executor) objectParentsFn() parentGetFn {
	return e.childrenOrParentGetFn(func(p treeNodeEntry, domain Domain) interface{} {
		return e.Enforcer.GetParentsForObjectInDomain(p.(Object), domain)
	})
}

func (e *Executor) roleParentUpdater() *parentEntryUpdater {
	return &parentEntryUpdater{
		newEntry:    e.newRole,
		parentGetFn: e.roleParentsFn(),
		parentAddFn: func(p1 treeNodeEntry, p2 treeNodeEntry, domain Domain) error {
			return e.Enforcer.AddParentForRoleInDomain(p1.(Role), p2.(Role), domain)
		},
		parentDelFn: func(p1 treeNodeEntry, p2 treeNodeEntry, domain Domain) error {
			return e.Enforcer.RemoveParentForRoleInDomain(p1.(Role), p2.(Role), domain)
		},
	}
}

func (e *Executor) roleDeleteFn() deleteFn {
	return func(p treeNodeEntry, d Domain) error {
		if err := e.Enforcer.RemoveRoleInDomain(p.(Role), d); err != nil {
			return err
		}
		return e.DB.DeleteByID(p, p.GetID())
	}
}

func (e *Executor) roleChildrenFn() childrenFn {
	return e.childrenOrParentGetFn(func(p treeNodeEntry, domain Domain) interface{} {
		return e.Enforcer.GetChildrenForRoleInDomain(p.(Role), domain)
	})
}

func (e *Executor) roleParentsFn() parentGetFn {
	return e.childrenOrParentGetFn(func(p treeNodeEntry, domain Domain) interface{} {
		return e.Enforcer.GetParentsForRoleInDomain(p.(Role), domain)
	})
}

func (e *Executor) childrenOrParentGetFn(fn func(treeNodeEntry, Domain) interface{}) childrenFn {
	return func(p treeNodeEntry, domain Domain) []treeNodeEntry {
		var out []treeNodeEntry
		linq.From(fn(p, domain)).ToSlice(&out)
		return out
	}
}

func (e *Executor) filter(action Action, source interface{}) ([]interface{}, error) {
	u, d, err := e.provider.Get()
	if err != nil {
		return nil, err
	}
	return Filter(e.Enforcer, u, d, action, source), nil
}

func (e *Executor) filterWithNoError(user User, domain Domain, action Action, source interface{}) []interface{} {
	return Filter(e.Enforcer, user, domain, action, source)
}

func (e *Executor) check(one ObjectData, action Action) error {
	u, d, err := e.provider.Get()
	if err != nil {
		return err
	}

	if ok := Check(e.Enforcer, u, d, one, action); !ok {
		switch action {
		case Read:
			return ErrNoReadPermission
		case Write:
			return ErrNoWritePermission
		default:
		}
	}

	return nil
}
