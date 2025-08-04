package utils

import (
	"user_service/internal/app/user/delivery/dto"
)

func ErrorHandler(message string) *dto.ErrorResponse {
	return &dto.ErrorResponse{
		Error: message,
	}
}

//func (u *UserService) handleNotFound(err error, id int64, operation string) error {
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//		u.log.Infof("User Not Found, id: %d", id)
//		return fmt.Errorf("user not found: %w", err)
//	}
//	u.log.Errorf("%s error: %v", operation, err)
//	return fmt.Errorf("%s error: %w", operation, err)
//}
