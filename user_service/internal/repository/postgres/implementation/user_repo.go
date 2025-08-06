package implementation

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"user_service/internal/app/models"
	"user_service/internal/app/user/service/dto"
	"user_service/internal/repository/postgres"
)

type UserRepo struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewUserRepo(db *gorm.DB, log *logrus.Logger) postgres.UserRepoInterface {
	return &UserRepo{
		db:  db,
		log: log,
	}

}

func (u *UserRepo) Create(ctx context.Context, person *models.User) (*models.User, error) {
	if err := u.db.Create(person).Error; err != nil {
		u.log.WithError(err).Error("create user error:")
		return nil, fmt.Errorf("create user error: %w", err)
	}
	return person, nil
}

func (u UserRepo) GetById(ctx context.Context, id int64) (*models.User, error) {
	var person models.User

	err := u.db.WithContext(ctx).
		First(&person, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.log.WithError(err).Info("user not found")
			return nil, fmt.Errorf("user not found: %w", err)
		}
		u.log.WithError(err).Error("get user by id error")
		return nil, fmt.Errorf("get user by id error: %w", err)
	}

	return &person, nil
}

func (u UserRepo) GetAll(ctx context.Context, filter dto.SearchUserFilter) (int, []*models.User, error) {
	if filter.Limit < 0 || filter.Offset < 0 {
		return 0, nil, fmt.Errorf("invalid pagination params: limit and offset must be positive")
	}

	query := u.db.WithContext(ctx).Model(&models.User{})

	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.Email+"%")
	}
	if filter.SortBy != "" {
		query = query.Order(filter.SortBy)
	}

	var total int64
	var users []*models.User

	if err := query.Count(&total).Error; err != nil {
		u.log.WithError(err).Error("count users error")
		return 0, nil, fmt.Errorf("count users error: %w", err)
	}

	if filter.Limit == 0 {
		return int(total), nil, nil
	}

	if err := query.
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&users).Error; err != nil {
		u.log.WithError(err).Error("get all users error")
		return 0, nil, fmt.Errorf("get all users error: %w", err)
	}

	return int(total), users, nil
}

func (u *UserRepo) Update(ctx context.Context, id int64, person *models.User) (*models.User, error) {
	tx := u.db.Begin().WithContext(ctx)
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	var existingUser models.User

	if err := tx.First(&existingUser, id).Error; err != nil {
		u.log.WithError(err).Error("get user by id error")
		return nil, fmt.Errorf("user not found: %w", err)
	}
	updates := map[string]interface{}{}
	if person.Email != "" {
		updates["Email"] = person.Email
	}
	if person.Username != "" {
		updates["Username"] = person.Username
	}
	if person.Password != "" {
		updates["Password"] = person.Password
	}
	if len(updates) == 0 {
		return &existingUser, nil
	}
	if err := tx.Model(&existingUser).Updates(updates).Error; err != nil {
		u.log.WithError(err).Error("update user error")
		return nil, fmt.Errorf("update user error: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		u.log.WithError(err).Error("failed to commit")
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	committed = true
	return &existingUser, nil
}

func (u *UserRepo) Delete(ctx context.Context, id int64) error {
	if err := u.db.Delete(&models.User{}, id).Error; err != nil {
		u.log.WithError(err).Error("delete user error:")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.log.WithError(err).Info("user not found")
			return fmt.Errorf("user not found: %w", err)
		}
		return fmt.Errorf("delete user error: %w", err)
	}
	return nil
}
