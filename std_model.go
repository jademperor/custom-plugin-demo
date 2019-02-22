package main

import (
	// "log"
	"sync"
)

// stdUser ...
type stdUser struct {
	ID     string              `json:"id"`
	UserID string              `json:"user_id"`
	Roles  map[string]*stdRole `json:"roles"`
	mu     sync.RWMutex
}

func (u *stdUser) assign(r *stdRole) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if _, ok := u.Roles[r.ID]; ok {
		return
	}

	u.Roles[r.ID] = r
}

func (u *stdUser) revoke(r *stdRole) {
	u.mu.Lock()
	defer u.mu.Unlock()

	delete(u.Roles, r.ID)
}

// stdPermission ...
type stdPermission struct {
	ID       string `json:"id"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

func (p *stdPermission) equal(other *stdPermission) bool {
	if p.ID != "" && p.ID == other.ID {
		return true
	}

	if p.Action == other.Action && p.Resource == other.Resource {
		return true
	}

	return false
}

// stdRole ...
type stdRole struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Permissions map[string]*stdPermission `json:"ps"`
	mu          sync.RWMutex
}

func (r *stdRole) assign(p *stdPermission) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.Permissions[p.ID]; ok {
		return
	}

	r.Permissions[p.ID] = p
}

func (r *stdRole) revoke(p *stdPermission) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Permissions, p.ID)
}

func (r *stdRole) permit(p *stdPermission) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, perm := range r.Permissions {
		// log.Printf("%v to %v", perm, p)
		if perm.equal(p) {
			return true
		}
	}
	return false
}

// stdPermitURL ...
type stdPermitURL struct {
	Permission *stdPermission `json:"p"`
	URI        string         `json:"uri"`
}
