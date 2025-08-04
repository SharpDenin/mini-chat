package validation

import (
	"errors"
	"fmt"
	"user_service/internal/app/user/entities/model"
)

func ValidateUserId(userId int64) error {
	if userId <= 0 {
		return fmt.Errorf("user ID must be positive")
	}
	return nil
}

func ValidateUserForCreate(u *model.User) error {
	if u == nil {
		return errors.New("userModel is nil")
	}
	if u.Username == "" || len(u.Username) > 50 {
		return errors.New("invalid username")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

func ValidateUserForUpdate(u *model.User) error {
	if u == nil {
		return errors.New("userModel is nil")
	}
	if u.Id <= 0 {
		return errors.New("user ID must be positive")
	}
	return nil
}
