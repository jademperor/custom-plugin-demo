package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"hash"
	"log"
	"os"
	"sync"

	"github.com/jademperor/api-proxier/plugin"
)

func main() {}

var (
	_               plugin.Plugin = &RBAC{}
	errNoPermission               = errors.New("No Permission")
)

// New ... only return a RBAC instance, and must load rules manually
func New(cfgData []byte) plugin.Plugin {
	c := new(rbacCfg)
	if err := json.Unmarshal(cfgData, c); err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	r := &RBAC{
		enabled:         true,
		status:          plugin.Working,
		userIDFieldName: c.UserIDFieldName,
		defaultRoleName: c.DefaultRoleName,
		md5er:           md5.New(),
	}

	r.loadUsers(c.Users, c.Roles, c.Permissions)
	r.loadPermitURLs(c.PermitURLs)

	return r
}

// RBAC ...
type RBAC struct {
	enabled         bool
	status          plugin.PlgStatus
	userIDFieldName string
	defaultRoleName string
	// cache using
	mapHashedPermitURL map[string]*stdPermission // map[$hashed_url_key]$permission
	md5er              hash.Hash

	// all users
	users       map[string]*stdUser
	roles       map[string]*stdRole
	permissions map[string]*stdPermission
}

// Handle ....
func (r *RBAC) Handle(ctx *plugin.Context) {
	defer plugin.Recover("plugin.rbac")
	var (
		hasPermission       bool
		needCheckPermission bool
	)
	// permit url
	if hasPermission, needCheckPermission = r.permit(ctx.Path,
		ctx.Form.Get(r.userIDFieldName)); needCheckPermission && !hasPermission {
		ctx.SetError(errNoPermission)
		ctx.Abort()
		return
	}

	log.Printf("with permission request passed: %v, %v", hasPermission, needCheckPermission)
	ctx.Next()
}

// Name ...
func (r *RBAC) Name() string {
	return "plugin.rbac"
}

// Enabled ...
func (r *RBAC) Enabled() bool {
	return r.enabled
}

// Status ...
func (r *RBAC) Status() plugin.PlgStatus {
	return r.status
}

// Enable ...
func (r *RBAC) Enable(enabled bool) {
	r.enabled = enabled
	r.status = plugin.Working
	if !enabled {
		r.status = plugin.Stopped
	}
}

func (r *RBAC) permit(uri, userID string) (hasPermission, needCheckPermission bool) {
	log.Printf("permit path: %s, with UserID: %s", uri, userID)
	needCheckPermission = true
	hasPermission = false
	if userID == "" {
		userID = r.defaultRoleName
	}

	var (
		p    *stdPermission
		user *stdUser
		ok   bool
	)
	if p, needCheckPermission = r.mapHashedPermitURL[r.hashURI(uri)]; !needCheckPermission {
		// no needCheckPermission to permit the request
		return
	}

	if user, ok = r.users[userID]; !ok {
		// missed userID
		log.Printf("could not found userId: %s", userID)
		hasPermission = false
		return
	}

	// brute force
	for _, role := range user.Roles {
		if hasPermission = role.permit(p); hasPermission {
			return
		}
	}

	log.Printf("permit user onw permission: [%v]", hasPermission)
	return
}

func (r *RBAC) hashURI(uri string) string {
	r.md5er.Reset()
	_, err := r.md5er.Write([]byte(uri))
	if err != nil {
		panic(err)
	}

	return string(r.md5er.Sum(nil))
}

// loadUsers from config.json load data into memory
func (r *RBAC) loadUsers(users map[string]*userCfg, roles map[string]*roleCfg, perms map[string]*permissionCfg) {
	r.users = make(map[string]*stdUser)
	r.roles = make(map[string]*stdRole)
	r.permissions = make(map[string]*stdPermission)

	for k, v := range perms {
		r.permissions[k] = &stdPermission{
			ID:       v.ID,
			Action:   v.Action,
			Resource: v.Resource,
		}
	}

	for k, v := range roles {
		r.roles[k] = &stdRole{
			ID:   v.ID,
			Name: v.Name,
			mu:   sync.RWMutex{},
		}
		r.roles[k].Permissions = make(map[string]*stdPermission)
		for _, id := range v.Permissions {
			r.roles[k].Permissions[id] = r.permissions[id]
		}
	}

	for _, v := range users {
		r.users[v.UserID] = &stdUser{
			ID:     v.ID,
			UserID: v.UserID,
			mu:     sync.RWMutex{},
		}
		r.users[v.UserID].Roles = make(map[string]*stdRole)
		for _, id := range v.Roles {
			r.users[v.UserID].Roles[id] = r.roles[id]
		}
	}
}

// loadPermitURLs ...
func (r *RBAC) loadPermitURLs(rules []*permitURLCfg) {
	r.mapHashedPermitURL = make(map[string]*stdPermission)

	for _, rule := range rules {
		hashed := r.hashURI(rule.URI)
		if rule.PermissionID == "" {
			panic("could not be nil Permission")
		}

		r.mapHashedPermitURL[hashed] = r.permissions[rule.PermissionID]
	}
}
