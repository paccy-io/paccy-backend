package user

type UserRole string

const (
	RoleBanned    UserRole = "banned"
	RoleUser      UserRole = "user"
	RoleAlpha     UserRole = "alpha_user"
	RoleBeta      UserRole = "beta_user"
	RoleReviewer  UserRole = "reviewer"
	RoleModerator UserRole = "moderator"
	RoleDev       UserRole = "developer"
	RoleAdmin     UserRole = "admin"
	RoleOwner     UserRole = "owner"
)

func IsValidRole(role string) bool {
	switch UserRole(role) {
	case RoleUser, RoleAdmin, RoleReviewer, RoleModerator, RoleAlpha, RoleBeta, RoleDev, RoleOwner:
		return true
	default:
		return false
	}
}
