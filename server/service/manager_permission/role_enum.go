package permission

type RoleCode string

const (
	RoleCodeSuperAdmin RoleCode = "super_admin"
	RoleCodeSalesman   RoleCode = "salesman"
)

const (
	RoleNameSuperAdmin = "超级管理员"
	RoleNameSalesman   = "业务员"
)

type RoleEnum struct {
	Name string
	Code RoleCode
}

var RoleEnums = []RoleEnum{
	{Name: RoleNameSuperAdmin, Code: RoleCodeSuperAdmin},
	{Name: RoleNameSalesman, Code: RoleCodeSalesman},
}
