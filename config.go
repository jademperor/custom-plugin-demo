package main

type rbacCfg struct {
	UserIDFieldName string                    `json:"field_name"`
	DefaultRoleName string                    `json:"default_role_name"`
	Users           map[string]*userCfg       `json:"users"`
	Roles           map[string]*roleCfg       `json:"roles"`
	Permissions     map[string]*permissionCfg `json:"permissions"`
	PermitURLs      []*permitURLCfg           `json:"permit_urls"`
}

type userCfg struct {
	ID     string   `json:"id"`
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
}
type roleCfg struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}
type permissionCfg struct {
	ID       string `json:"id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}
type permitURLCfg struct {
	URI          string `json:"uri"`
	PermissionID string `json:"p"`
}
