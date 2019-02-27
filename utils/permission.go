package utils

const (
	ADMIN   = "admin"
	MOD     = "mod"
	JANITOR = "janitor"
)

var permissions = map[string]int{ADMIN: 1, MOD: 2, JANITOR: 3}

func PermissionExists(p string) bool {
	_, ok := permissions[p]
	return ok
}

func CheckPermission(minAllowedPermission string, currentPermission string) bool {
	a, ok := permissions[minAllowedPermission]
	if !ok {
		return false
	}
	c, ok := permissions[currentPermission]
	if !ok {
		return false
	}
	return c <= a
}
