package caskin

import "github.com/ahmetb/go-linq/v3"

// GetAllPoliciesForRole
// 1. get all policies which current user has role and object's read permission in current domain
// 2. get role to objects 's p as ObjectsForRole in current domain
// 3. build role's tree
func (e *executor) GetAllPoliciesForRole() ([]*PoliciesForRole, error) {
	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return nil, err
	}

	rs := e.e.GetRolesInDomain(currentDomain)
	tree := getTree(rs)
	roles, err := e.mdb.GetRoleInDomain(currentDomain)
	if err != nil {
		return nil, err
	}
	r := e.filterWithNoError(currentUser, currentDomain, Read, roles)
	roles = []Role{}
	for _, v := range r {
		roles = append(roles, v.(Role))
	}

	objects, err := e.mdb.GetObjectInDomain(currentDomain)
	if err != nil {
		return nil, err
	}

	os := e.filterWithNoError(currentUser, currentDomain, Read, objects)
	objects = []Object{}
	for _, v := range os {
		objects = append(objects, v.(Object))
	}
	om := getIDMap(objects)

	e.e.GetPoliciesInDomain(currentDomain)
	var prs []*PoliciesForRole
	for _, v := range roles {
		if p, ok := tree[v.GetID()]; ok {
			v.SetParentID(p)
		}

		pr := &PoliciesForRole{Role: v}
		policy := e.e.GetPoliciesForRoleInDomain(v, currentDomain)
		for _, p := range policy {
			p.Object.GetID()
			if object, ok := om[p.Object.GetID()]; ok {
				pr.Policies = append(pr.Policies, &Policy{
					Role:   v,
					Object: object.(Object),
					Domain: currentDomain,
					Action: p.Action,
				})
			}
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

// ModifyPoliciesForRole
// if current user has user and role and object's write permission
// 1. modify role to objects 's p in current domain
func (e *executor) ModifyPoliciesForRole(pr *PoliciesForRole) error {
	if err := isValid(pr.Role); err != nil {
		return err
	}

	if err := e.mdb.Take(pr.Role); err != nil {
		return ErrNotExists
	}

	if err := e.check(Write, pr.Role); err != nil {
		return err
	}

	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return err
	}

	role := pr.Role
	policy := e.e.GetPoliciesForRoleInDomain(role, currentDomain)
	var oid, oid1, oid2 []uint64
	for _, v := range policy {
		oid1 = append(oid1, v.Object.GetID())
	}
	for _, v := range pr.Policies {
		oid2 = append(oid2, v.Object.GetID())
	}
	oid = append(oid, oid1...)
	oid = append(oid, oid2...)
	linq.From(oid).Distinct().ToSlice(&oid)
	objects, err := e.mdb.GetObjectByID(oid)
	if err != nil {
		return err
	}

	os := e.filterWithNoError(currentUser, currentDomain, Write, objects)
	objects = []Object{}
	for _, v := range os {
		objects = append(objects, v.(Object))
	}
	om := getIDMap(objects)

	// make source and target role id list
	var source, target []*Policy
	for _, v := range policy {
		if _, ok := om[v.Object.GetID()]; ok {
			source = append(source, v)
		}
	}
	for _, v := range pr.Policies {
		if _, ok := om[v.Object.GetID()]; ok {
			target = append(target, v)
		}
	}

	// get diff to add and remove
	add, remove := DiffPolicy(source, target)
	for _, v := range add {
		if err := e.e.AddPolicyInDomain(v.Role, v.Object, v.Domain, v.Action); err != nil {
			return err
		}
	}
	for _, v := range remove {
		if err := e.e.RemovePolicyInDomain(v.Role, v.Object, v.Domain, v.Action); err != nil {
			return err
		}
	}

	return nil
}

// GetPolicyList
// 1. get all policies which current user has role and object's read permission in current domain
// 2. build role's tree
func (e *executor) GetPolicyList() ([]*Policy, error) {
	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return nil, err
	}

	rs := e.e.GetRolesInDomain(currentDomain)
	tree := getTree(rs)
	roles, err := e.mdb.GetRoleInDomain(currentDomain)
	if err != nil {
		return nil, err
	}
	r := e.filterWithNoError(currentUser, currentDomain, Read, roles)
	roles = []Role{}
	for _, v := range r {
		roles = append(roles, v.(Role))
	}

	objects, err := e.mdb.GetObjectInDomain(currentDomain)
	if err != nil {
		return nil, err
	}
	os := e.filterWithNoError(currentUser, currentDomain, Read, objects)
	objects = []Object{}
	for _, v := range os {
		objects = append(objects, v.(Object))
	}
	om := getIDMap(objects)

	e.e.GetPoliciesInDomain(currentDomain)
	var list []*Policy
	for _, v := range roles {
		if p, ok := tree[v.GetID()]; ok {
			v.SetParentID(p)
		}

		policy := e.e.GetPoliciesForRoleInDomain(v, currentDomain)
		for _, p := range policy {
			if object, ok := om[p.Object.GetID()]; ok {
				list = append(list, &Policy{
					Role:   v,
					Object: object.(Object),
					Domain: currentDomain,
					Action: p.Action,
				})
			}
		}
	}

	return list, nil
}

// GetPolicyListByRole
// 1. get policy which current user has role and object's read permission in current domain
// 2. get user to role 's g as Policy in current domain
func (e *executor) GetPolicyListByRole(role Role) ([]*Policy, error) {
	if err := isValid(role); err != nil {
		return nil, err
	}

	if err := e.mdb.Take(role); err != nil {
		return nil, err
	}

	if err := e.check(Read, role); err != nil {
		return nil, err
	}

	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return nil, err
	}

	objects, err := e.mdb.GetObjectInDomain(currentDomain)
	if err != nil {
		return nil, err
	}
	os := e.filterWithNoError(currentUser, currentDomain, Read, objects)
	objects = []Object{}
	for _, v := range os {
		objects = append(objects, v.(Object))
	}
	om := getIDMap(objects)

	e.e.GetPoliciesInDomain(currentDomain)
	var list []*Policy

	policy := e.e.GetPoliciesForRoleInDomain(role, currentDomain)
	for _, p := range policy {
		if object, ok := om[p.Object.GetID()]; ok {
			list = append(list, &Policy{
				Role:   role,
				Object: object.(Object),
				Domain: currentDomain,
				Action: p.Action,
			})
		}
	}

	return list, nil
}

// GetPolicyListByObject
// 1. get policy which current user has role and object's read permission in current domain
// 2. get user to role 's g as Policy in current domain
func (e *executor) GetPolicyListByObject(object Object) ([]*Policy, error) {
	if err := isValid(object); err != nil {
		return nil, err
	}

	if err := e.mdb.Take(object); err != nil {
		return nil, err
	}

	if err := e.check(Read, object); err != nil {
		return nil, err
	}

	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return nil, err
	}

	roles, err := e.mdb.GetRoleInDomain(currentDomain)
	if err != nil {
		return nil, err
	}
	r := e.filterWithNoError(currentUser, currentDomain, Read, roles)
	roles = []Role{}
	for _, v := range r {
		roles = append(roles, v.(Role))
	}
	rm := getIDMap(roles)

	var list []*Policy
	policy := e.e.GetPoliciesForObjectInDomain(object, currentDomain)
	for _, p := range policy {
		if role, ok := rm[p.Role.GetID()]; ok {
			list = append(list, &Policy{
				Role:   role.(Role),
				Object: object,
				Domain: currentDomain,
				Action: p.Action,
			})
		}
	}

	return list, nil
}

// ModifyPolicyListPerRole
// if current user has role and object's write permission
// 1. modify role to objects 's p in current domain
func (e *executor) ModifyPolicyListPerRole(role Role, input []*Policy) error {
	if err := isValid(role); err != nil {
		return err
	}

	if err := e.mdb.Take(role); err != nil {
		return ErrNotExists
	}

	if err := e.check(Write, role); err != nil {
		return err
	}

	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return err
	}

	policy := e.e.GetPoliciesForRoleInDomain(role, currentDomain)
	var oid, oid1, oid2 []uint64
	for _, v := range policy {
		oid1 = append(oid1, v.Object.GetID())
	}
	for _, v := range input {
		oid2 = append(oid2, v.Object.GetID())
	}
	oid = append(oid, oid1...)
	oid = append(oid, oid2...)
	linq.From(oid).Distinct().ToSlice(&oid)
	objects, err := e.mdb.GetObjectByID(oid)
	if err != nil {
		return err
	}

	os := e.filterWithNoError(currentUser, currentDomain, Write, objects)
	objects = []Object{}
	for _, v := range os {
		objects = append(objects, v.(Object))
	}
	om := getIDMap(objects)

	// make source and target role id list
	var source, target []*Policy
	for _, v := range policy {
		if _, ok := om[v.Object.GetID()]; ok {
			source = append(source, v)
		}
	}
	for _, v := range input {
		if _, ok := om[v.Object.GetID()]; ok {
			target = append(target, v)
		}
	}

	// get diff to add and remove
	add, remove := DiffPolicy(source, target)
	for _, v := range add {
		if err := e.e.AddPolicyInDomain(v.Role, v.Object, v.Domain, v.Action); err != nil {
			return err
		}
	}
	for _, v := range remove {
		if err := e.e.RemovePolicyInDomain(v.Role, v.Object, v.Domain, v.Action); err != nil {
			return err
		}
	}

	return nil
}

// ModifyPolicyListPerObject
// if current user has role and object's write permission
// 1. modify role to objects 's p in current domain
func (e *executor) ModifyPolicyListPerObject(object Object, input []*Policy) error {
	if err := isValid(object); err != nil {
		return err
	}

	if err := e.mdb.Take(object); err != nil {
		return ErrNotExists
	}

	if err := e.check(Write, object); err != nil {
		return err
	}

	currentUser, currentDomain, err := e.provider.Get()
	if err != nil {
		return err
	}

	policy := e.e.GetPoliciesForObjectInDomain(object, currentDomain)
	var rid, rid1, rid2 []uint64
	for _, v := range policy {
		rid1 = append(rid1, v.Role.GetID())
	}
	for _, v := range input {
		rid2 = append(rid2, v.Role.GetID())
	}
	rid = append(rid, rid1...)
	rid = append(rid, rid2...)
	linq.From(rid).Distinct().ToSlice(&rid)
	roles, err := e.mdb.GetRoleByID(rid)
	if err != nil {
		return err
	}

	rs := e.filterWithNoError(currentUser, currentDomain, Write, roles)
	roles = []Role{}
	for _, v := range rs {
		roles = append(roles, v.(Role))
	}
	rm := getIDMap(roles)

	// make source and target role id list
	var source, target []*Policy
	for _, v := range policy {
		if _, ok := rm[v.Role.GetID()]; ok {
			source = append(source, v)
		}
	}
	for _, v := range input {
		if _, ok := rm[v.Role.GetID()]; ok {
			target = append(target, v)
		}
	}

	// get diff to add and remove
	add, remove := DiffPolicy(source, target)
	for _, v := range add {
		if err := e.e.AddPolicyInDomain(v.Role, v.Object, v.Domain, v.Action); err != nil {
			return err
		}
	}
	for _, v := range remove {
		if err := e.e.RemovePolicyInDomain(v.Role, v.Object, v.Domain, v.Action); err != nil {
			return err
		}
	}

	return nil
}

