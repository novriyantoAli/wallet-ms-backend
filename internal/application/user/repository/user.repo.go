package repository

import (
	"context"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uint) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetAll(ctx context.Context, filter *dto.UserFilter) ([]entity.User, int64, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uint) error
	EmailExists(ctx context.Context, email string) (bool, error)
}

type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	r.logger.Info("Creating user", zap.String("email", user.Email))
	db := database.GetDB(ctx, r.db)
	return db.Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	db := database.GetDB(ctx, r.db)
	err := db.First(&user, id).Error
	if err != nil {
		r.logger.Error("Failed to get user by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	db := database.GetDB(ctx, r.db)
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		r.logger.Error("Failed to get user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, filter *dto.UserFilter) ([]entity.User, int64, error) {
	var users []entity.User
	var totalCount int64

	db := database.GetDB(ctx, r.db)
	query := db.Model(&entity.User{})

	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.Email+"%")
	}
	if filter.Level != "" {
		query = query.Where("level = ?", filter.Level)
	}

	query.Count(&totalCount)

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err := query.Find(&users).Error
	if err != nil {
		r.logger.Error("Failed to get users", zap.Error(err))
		return nil, 0, err
	}

	return users, totalCount, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	r.logger.Info("Updating user", zap.Uint("id", user.ID))
	db := database.GetDB(ctx, r.db)
	return db.Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Info("Deleting user", zap.Uint("id", id))
	db := database.GetDB(ctx, r.db)
	return db.Delete(&entity.User{}, id).Error
}

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	db := database.GetDB(ctx, r.db)
	err := db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
