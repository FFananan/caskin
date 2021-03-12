package caskin

import "github.com/ahmetb/go-linq/v3"

// FilterObjectData
// filter object_data with action
func (e *executor) FilterObjectData(source interface{}, action Action) ([]ObjectData, error) {
	u, d, err := e.provider.Get()
	if err != nil {
		return nil, err
	}

	var result []ObjectData
	linq.From(source).Where(func(v interface{}) bool {
		return Check(e.e, u, d, v.(ObjectData), action)
	}).ToSlice(&result)
	return result, nil
}

// Enforce
// check permission of object_data with action
func (e *executor) Enforce(item ObjectData, action Action) error {
	return e.check(item, action)
}

// CreateObjectDataCheck
// check permission of creating object_data
func (e *executor) CreateObjectDataCheck(item ObjectData, ty ObjectType) error {
	return e.writeObjectDataCheck(item, ty)
}

// RecoverObjectDataCheck
// check permission of recover object_data
func (e *executor) RecoverObjectDataCheck(item ObjectData, ty ObjectType) error {
	return e.writeObjectDataCheck(item, ty)
}

// UpdateObjectDataCheck
// check permission of updating object_data's
func (e *executor) UpdateObjectDataCheck(item ObjectData, old ObjectData, ty ObjectType) error {
	list := []ObjectData{item}
	if item.GetObject().GetID() != old.GetObject().GetID() {
		list = append(list, old)
	}
	for _, v := range list {
		if err := e.writeObjectDataCheck(v, ty); err != nil {
			return err
		}
	}
	return nil
}

// DeleteObjectDataCheck
// check permission of deleting object_data
func (e *executor) DeleteObjectDataCheck(item ObjectData, ty ObjectType) error {
	return e.writeObjectDataCheck(item, ty)
}

func (e *executor) writeObjectDataCheck(item ObjectData, ty ObjectType) error {
	if err := e.check(item, Write); err != nil {
		return err
	}
	o := item.GetObject()
	if err := e.mdb.Take(o); err != nil {
		return ErrInValidObject
	}
	if o.GetObjectType() != ty {
		return ErrInValidObjectType
	}
	return nil
}
