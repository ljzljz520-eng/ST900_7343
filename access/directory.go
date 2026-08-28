package access

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

type Directory struct {
	mu         sync.RWMutex
	principals map[string]Principal
}

func NewDirectory(seed []Principal) (*Directory, error) {
	directory := &Directory{principals: make(map[string]Principal)}
	for _, principal := range seed {
		if err := directory.Put(principal); err != nil {
			return nil, err
		}
	}
	return directory, nil
}

func DefaultDirectory() *Directory {
	owner, _ := NewPrincipal("admin", AdminOwner)
	directory, _ := NewDirectory([]Principal{owner})
	return directory
}

func (d *Directory) Put(principal Principal) error {
	if err := principal.Validate(); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.principals[principal.ID] = principal
	return nil
}

func (d *Directory) Get(id string) (Principal, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Principal{}, errors.New("administrator id is required")
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	principal, exists := d.principals[id]
	if !exists {
		return Principal{}, errors.New("administrator account does not exist")
	}
	return principal, nil
}

func (d *Directory) Remove(id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.principals[id]; !exists {
		return errors.New("administrator account does not exist")
	}
	delete(d.principals, id)
	return nil
}

func (d *Directory) List() []Principal {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]Principal, 0, len(d.principals))
	for _, principal := range d.principals {
		result = append(result, principal)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (d *Directory) Authorize(id string, permission Permission) (Principal, error) {
	principal, err := d.Get(id)
	if err != nil {
		return Principal{}, err
	}
	if err := Require(principal, permission); err != nil {
		return Principal{}, err
	}
	return principal, nil
}

func (d *Directory) ReplaceRole(id string, role AdminRole) (Principal, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	current, exists := d.principals[id]
	if !exists {
		return Principal{}, errors.New("administrator account does not exist")
	}
	permissions := PermissionsForRole(role)
	if len(permissions) == 0 {
		return Principal{}, errors.New("administrator role is invalid")
	}
	current.Role = role
	current.Permissions = permissions
	d.principals[id] = current
	return current, nil
}
