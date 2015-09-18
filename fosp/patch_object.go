// Copyright (C) 2015 Felix Maurer
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package fosp

import (
	"errors"
	"fmt"
)

type PatchObject map[string]interface{}

type Patchable interface {
	Patch(PatchObject) error
}

func (p PatchObject) PatchString(target *string, field string) error {
	if tmp, ok := p[field]; ok {
		if tmp == nil {
			*target = ""
		} else if value, ok := tmp.(string); ok {
			*target = value
		} else {
			return errors.New("Field " + field + " does not contain a string or nil")
		}
	}
	return nil
}

func (p PatchObject) PatchStruct(target interface{}, field string) error {
	patchable, ok := target.(Patchable)
	if !ok {
		return errors.New(fmt.Sprintf("Target %T is not patchable", target))
	}
	if tmp, ok := p[field]; ok {
		if tmp == nil {
			target = nil
		} else if patch, ok := tmp.(map[string]interface{}); ok {
			return patchable.Patch(PatchObject(patch))
		} else {
			return errors.New(fmt.Sprintf("Field %s does not contain an object (%#v instead)", field, tmp))
		}
	}
	return nil
}

func (p PatchObject) GetStringSlice(field string) ([]string, bool, error) {
	value, ok := p[field]
	if !ok {
		return nil, false, nil
	}
	if value == nil {
		return nil, true, nil
	}
	slice, ok := value.([]interface{})
	if !ok {
		return nil, true, errors.New("Field " + field + " does not contain an array")
	}
	stringSlice := make([]string, len(slice))
	for i, element := range slice {
		if string, ok := element.(string); ok {
			stringSlice[i] = string
		} else {
			return nil, true, errors.New(fmt.Sprintf("Element %d in field %s is not a string (%#v instead)", i, field, element))
		}
	}
	return stringSlice, true, nil
}
