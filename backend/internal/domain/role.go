package domain

// Role — роль пользователя. В БД хранится как smallint (int16). Не передавать как int нигде, только тип Role.
// Admin и Support выставляются через админку; Owner — при регистрации (создании компании); Manager — создаётся Owner-ом на сайте.
type Role int16

func (r Role) IsPrivileged() bool {
	return r == RoleAdmin || r == RoleSupport
}

const (
	RoleAdmin   Role = 0
	RoleSupport Role = 1
	RoleOwner   Role = 2
	RoleManager Role = 3
)

var roleNames = map[Role]string{
	RoleAdmin:   "Admin",
	RoleSupport: "Support",
	RoleOwner:   "Owner",
	RoleManager: "Manager",
}

// String возвращает название роли для отображения (в т.ч. в дефолтном full_name).
func (r Role) String() string {
	if name, ok := roleNames[r]; ok {
		return name
	}
	return roleNames[RoleManager]
}
