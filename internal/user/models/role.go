package models

import (
	. "github.com/paccy-io/paccy-backend/pkg/errors"
)

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

func ValidateRoles(roles []string) error {
	for _, role := range roles {
		if !IsValidRole(role) {
			return ErrInvalidRole
		}
	}
	return nil
}

func ValidateAndConvertRoles(roles []string) ([]UserRole, error) {
	validRoles := []UserRole{}
	for _, role := range roles {
		if IsValidRole(role) {
			validRoles = append(validRoles, UserRole(role))
		} else {
			return nil, ErrInvalidRole
		}
	}
	return validRoles, nil
}
