package helpers

// TODO Пробросить вот эту шнягу в сервис
//func (u *service.UserService) handleNotFound(err error, id int64, operation string) error {
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//		u.log.Infof("User Not Found, id: %d", id)
//		return middleware.NewCustomError(http.StatusNotFound, fmt.Sprintf("User not found, id: %d", id), err)
//	}
//	u.log.Errorf("%s error: %v", operation, err)
//	return middleware.NewCustomError(http.StatusInternalServerError, fmt.Sprintf("%s error", operation), err)
//}
