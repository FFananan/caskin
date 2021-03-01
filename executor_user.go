package caskin

import "github.com/ahmetb/go-linq/v3"

// GetAllRolesForUser
// 1. get all user which current user has read permission in current domain
// 2. get all role which current user has read permission in current domain
// 3. get user to roles 's g as RolesForUser in current domain
func (e *executor) GetAllRolesForUser() ([]*RolesForUser, error) {
	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return nil, err
	}

	us := e.e.GetUsersInDomain(currentDomain)
	uid := getIDList(us)
	linq.From(uid).Distinct().ToSlice(&uid)
	users, err := e.mdb.GetUserByID(uid)
	if err != nil {
		return nil, err
	}
	users = e.filterWithNoError(currentUser, currentDomain, Read, users).([]User)

	roles, err := e.mdb.GetRoleInDomain(currentDomain)
	if err != nil {
		return nil, err
	}
	roles = e.filterWithNoError(currentUser, currentDomain, Read, roles).([]Role)
	rm := getIDMap(roles)

	var rus []*RolesForUser
	for _, v := range users {
		ru := &RolesForUser{User: v}
		rs := e.e.GetRolesForUserInDomain(v, currentDomain)
		for _, r := range rs {
			if role, ok := rm[r.GetID()]; ok {
				ru.Roles = append(ru.Roles, role.(Role))
			}
		}
		rus = append(rus, ru)
	}

	return rus, nil
}

// ModifyRolesForUser
// if current user has user and role's write permission
// 1. modify user to roles 's g in current domain
func (e *executor) ModifyRolesForUser(ru *RolesForUser) error {
	if err := isValid(ru.User); err != nil {
		return err
	}

	if err := e.mdb.TakeUser(ru.User); err != nil {
		return ErrNotExists
	}

	if err := e.check(Write, ru.User); err != nil {
		return err
	}

	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return err
	}

	user := ru.User
	rs := e.e.GetRolesForUserInDomain(user, currentDomain)
	rid1 := getIDList(rs)
	rid2 := getIDList(ru.Roles)

	// get all role data
	var rid []uint64
	rid = append(rid, rid1...)
	rid = append(rid, rid2...)
	linq.From(rid).Distinct().ToSlice(&rid)
	roles, err := e.mdb.GetRoleByID(rid)
	if err != nil {
		return err
	}
	roles = e.filterWithNoError(currentUser, currentDomain, Write, roles).([]Role)
	rm := getIDMap(roles)

	// make source and target role id list
	var source, target []interface{}
	for _, v := range rid1 {
		if _, ok := rm[v]; ok {
			source = append(source, v)
		}
	}
	for _, v := range rid2 {
		if _, ok := rm[v]; ok {
			target = append(target, v)
		}
	}

	// get diff to add and remove
	add, remove := Diff(source, target)
	for _, v := range add {
		r := rm[v.(uint64)]
		if err := e.e.AddRoleForUserInDomain(user, r.(Role), currentDomain); err != nil {
			return err
		}
	}
	for _, v := range remove {
		r := rm[v.(uint64)]
		if err := e.e.RemoveRoleForUserInDomain(user, r.(Role), currentDomain); err != nil {
			return err
		}
	}

	return nil
}

// CreateUser
// if current user has user's write permission and there does not exist the user
// then create a new one
// 1. create a new user into metadata database
func (e *executor) CreateUser(user User) error {
	return e.createOrRecoverUser(user, e.mdb.CreateUser)
}

// RecoverUser
// if current user has user's write permission and there exist the user but soft deleted
// then recover it
// 1. recover the soft delete one user at metadata database
func (e *executor) RecoverUser(user User) error {
	return e.createOrRecoverUser(user, e.mdb.RecoverUser)
}

// DeleteUser
// if current user has user's write permission
// 1. delete all user's g in the domain
// 2. soft delete one user in metadata database
func (e *executor) DeleteUser(user User) error {
	fn := func(u User) error {
		_, domain, _ := e.provider.Get()
		if err := e.e.RemoveUserInDomain(u, domain); err != nil {
			return err
		}
		return e.mdb.DeleteUserByID(u.GetID())
	}

	return e.writeUser(user, fn)
}

// UpdateUser
// if current user has user's write permission and there exist the user
// 1. just update user's properties
func (e *executor) UpdateUser(user User) error {
	return e.writeUser(user, e.mdb.UpdateUser)
}

func (e *executor) createOrRecoverUser(domain User, fn func(User) error) error {
	if err := e.mdb.TakeUser(domain); err == nil {
		return ErrAlreadyExists
	}

	return fn(domain)
}

func (e *executor) writeUser(user User, fn func(User) error) error {
	if err := isValid(user); err != nil {
		return err
	}

	if err := e.mdb.TakeUser(user); err != nil {
		return ErrNotExists
	}

	if err := e.check(Write, user); err != nil {
		return err
	}

	return fn(user)
}
