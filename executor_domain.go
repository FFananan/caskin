package caskin

// CreateDomain
// if there does not exist the domain
// then create a new one without permission checking
// 1. create a new domain into metadata database
func (e *Executor) CreateDomain(domain Domain) error {
	return e.createOrRecoverDomain(domain, e.db.Create)
}

// RecoverDomain
// if there exist the domain but soft deleted
// then recover it without permission checking
// 1. recover the soft delete one domain at metadata database
func (e *Executor) RecoverDomain(domain Domain) error {
	return e.createOrRecoverDomain(domain, e.db.Recover)
}

// DeleteDomain
// if there exist the domain
// soft delete the domain without permission checking
// 1. delete all user's g in the domain
// 2. don't delete any role's g or object's g2 in the domain
// 3. soft delete one domain in metadata database
func (e *Executor) DeleteDomain(domain Domain) error {
	fn := func(interface{}) error {
		if err := e.e.RemoveUsersInDomain(domain); err != nil {
			return err
		}
		return e.db.DeleteByID(domain, domain.GetID())
	}

	return e.writeDomain(domain, fn)
}

// UpdateDomain
// if there exist the domain
// update domain without permission checking
// 1. just update domain's properties
func (e *Executor) UpdateDomain(domain Domain) error {
	return e.writeDomain(domain, e.db.Update)
}

// ReInitializeDomain
// if there exist the domain
// re initialize the domain without permission checking
// 1. just re initialize the domain
func (e *Executor) ReInitializeDomain(domain Domain) error {
	fn := func(interface{}) error {
		return e.initializeDomain(domain)
	}

	return e.writeDomain(domain, fn)
}

// GetAllDomain
// get all domain without permission checking
func (e *Executor) GetAllDomain() ([]Domain, error) {
	return e.db.GetAllDomain()
}

func (e *Executor) createOrRecoverDomain(domain Domain, fn func(interface{}) error) error {
	if err := e.db.Take(domain); err == nil {
		return ErrAlreadyExists
	}

	return fn(domain)
}

func (e *Executor) writeDomain(domain Domain, fn func(interface{}) error) error {
	if err := isValid(domain); err != nil {
		return err
	}

	tmp := e.factory.NewDomain()
	tmp.SetID(domain.GetID())

	if err := e.db.Take(tmp); err != nil {
		return ErrNotExists
	}

	return fn(domain)
}

// initializeDomain
// it is reentrant to initialize a new domain
// 1. get roles, objects, policies form DomainCreator
// 2. upsert roles, objects into metadata database
// 3. add policies as p into casbin
func (e *Executor) initializeDomain(domain Domain) error {
	creator := e.options.DomainCreator(domain)
	roles, objects := creator.BuildCreator()
	for _, v := range objects {
		if err := e.db.Upsert(v); err != nil {
			return err
		}
	}
	for _, v := range roles {
		if err := e.db.Upsert(v); err != nil {
			return err
		}
	}

	creator.SetRelation()
	for _, v := range roles {
		if err := e.db.Upsert(v); err != nil {
			return err
		}
	}
	for _, v := range objects {
		if err := e.db.Upsert(v); err != nil {
			return err
		}
	}

	policies := creator.GetPolicy()
	for _, v := range policies {
		if err := e.e.AddPolicyInDomain(v.Role, v.Object, v.Domain, v.Action); err != nil {
			return err
		}
	}

	return nil
}
