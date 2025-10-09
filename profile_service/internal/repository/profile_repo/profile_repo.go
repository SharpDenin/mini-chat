package profile_repo

import (
	"context"
	"errors"
	"fmt"
	"profile_service/internal/models"
	"profile_service/internal/service/service_dto"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProfileRepo struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewProfileRepo(db *gorm.DB, log *logrus.Logger) ProfileRepoInterface {
	return &ProfileRepo{
		db:  db,
		log: log,
	}

}

func (u *ProfileRepo) Create(ctx context.Context, person *models.User) (*models.User, error) {
	if person == nil {
		u.log.Error("Create user error: user is nil")
		return nil, fmt.Errorf("create user error: user is nil")
	}

	tx := u.db.Begin().WithContext(ctx)
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if err := tx.Create(person).Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err}).Error("Failed to create user")
		return nil, fmt.Errorf("create user error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err}).Error("Failed to commit transaction")
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	committed = true

	return person, nil
}

func (u *ProfileRepo) GetById(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		u.log.WithFields(logrus.Fields{"id": id}).Error("Invalid user id")
		return nil, fmt.Errorf("invalid user ID: %d", id)
	}

	var person models.User

	err := u.db.WithContext(ctx).First(&person, id).Error
	if err != nil {
		u.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to get user by Id")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id error: %w", err)
	}

	return &person, nil
}

func (u *ProfileRepo) GetAll(ctx context.Context, filter service_dto.SearchUserFilter) (int, []*models.User, error) {
	if filter.Limit < 0 || filter.Offset < 0 {
		u.log.WithFields(logrus.Fields{"limit": filter.Limit, "offset": filter.Offset}).Error("Invalid pagination params")
		return 0, nil, fmt.Errorf("invalid pagination params: limit and offset must be positive")
	}

	if filter.Limit > 100 {
		u.log.WithFields(logrus.Fields{"limit": filter.Limit}).Warn("Limit exceeds maximum allowed value, setting to 1000")
		filter.Limit = 100
	}

	query := u.db.WithContext(ctx).Model(&models.User{})

	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.Email+"%")
	}
	if filter.SortBy != "" {
		validSortFields := map[string]bool{
			"username":   true,
			"email":      true,
			"created_at": true,
		}
		if !validSortFields[filter.SortBy] {
			u.log.WithFields(logrus.Fields{"sort_by": filter.SortBy}).Error("Invalid sort field")
			return 0, nil, fmt.Errorf("invalid sort field: %s", filter.SortBy)
		}
		query = query.Order(filter.SortBy)
	}

	var total int64
	var users []*models.User

	if err := query.Count(&total).Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err}).Error("Failed to count users")
		return 0, nil, fmt.Errorf("count users error: %w", err)
	}

	if filter.Limit == 0 {
		return int(total), []*models.User{}, nil
	}

	if err := query.
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&users).Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err}).Error("Failed to get all users")
		return 0, nil, fmt.Errorf("get all users error: %w", err)
	}

	return int(total), users, nil
}

func (u *ProfileRepo) Update(ctx context.Context, id int64, person *models.User) (*models.User, error) {
	if id <= 0 {
		u.log.WithFields(logrus.Fields{"id": id}).Error("Invalid user ID")
		return nil, fmt.Errorf("invalid user ID: %d", id)
	}
	if person == nil {
		u.log.Error("Update user error: user is nil")
		return nil, fmt.Errorf("update user error: user is nil")
	}

	tx := u.db.Begin().WithContext(ctx)
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	var existingUser models.User
	if err := tx.First(&existingUser, id).Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to fetch user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("fetch user error: %w", err)
	}

	updates := map[string]interface{}{}
	if person.Username != "" {
		updates["username"] = person.Username
	}
	if person.Email != "" {
		updates["email"] = person.Email
	}

	if len(updates) == 0 {
		u.log.WithFields(logrus.Fields{"id": id}).Info("No fields to update")
		committed = true
		return &existingUser, nil
	}

	if err := tx.Model(&existingUser).Updates(updates).Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to update user")
		return nil, fmt.Errorf("update user error: %w", err)
	}

	if err := tx.First(&existingUser, id).Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to fetch updated user")
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err}).Error("Failed to commit transaction")
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	committed = true

	return &existingUser, nil
}

func (u *ProfileRepo) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		u.log.WithFields(logrus.Fields{"id": id}).Error("Invalid user ID")
		return fmt.Errorf("invalid user ID: %d", id)
	}

	tx := u.db.Begin().WithContext(ctx)
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&models.User{}, id).Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err, "id": id}).Error("Failed to delete user")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("delete user error: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		u.log.WithFields(logrus.Fields{"error": err}).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit: %w", err)
	}
	committed = true

	return nil
}
