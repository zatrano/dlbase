package services

import (
	"context"
	"errors"

	"davet.link/models"
	"davet.link/pkg/queryparams"
	"davet.link/repositories"

	"davet.link/configs/configslog"
	"go.uber.org/zap"
)

type IUserService interface {
	GetAllUsers(params queryparams.ListParams) (*queryparams.PaginatedResult, error)
	GetUserByID(id uint) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	BulkCreateUsers(ctx context.Context, users []models.User) error
	UpdateUser(ctx context.Context, id uint, userData *models.User) error
	BulkUpdateUsers(ctx context.Context, condition map[string]interface{}, data map[string]interface{}) error
	DeleteUser(ctx context.Context, id uint) error
	BulkDeleteUsers(ctx context.Context, condition map[string]interface{}) error
	GetUserCount() (int64, error)
	CreateUserWithPassword(ctx context.Context, user *models.User, password string) error
	UpdateUserWithPassword(ctx context.Context, id uint, userData *models.User, newPassword string) error
}

type UserService struct {
	repo repositories.IUserRepository
}

func NewUserService() IUserService {
	return &UserService{
		repo: repositories.NewUserRepository(),
	}
}

func (s *UserService) GetAllUsers(params queryparams.ListParams) (*queryparams.PaginatedResult, error) {
	users, totalCount, err := s.repo.GetAllUsers(params)
	if err != nil {
		configslog.Log.Error("Kullanıcılar alınamadı", zap.Error(err))
		return nil, errors.New("kullanıcılar getirilirken bir hata oluştu")
	}

	return &queryparams.PaginatedResult{
		Data: users,
		Meta: queryparams.PaginationMeta{
			CurrentPage: params.Page,
			PerPage:     params.PerPage,
			TotalItems:  totalCount,
			TotalPages:  queryparams.CalculateTotalPages(totalCount, params.PerPage),
		},
	}, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		configslog.Log.Warn("Kullanıcı bulunamadı", zap.Uint("user_id", id), zap.Error(err))
		return nil, errors.New("kullanıcı bulunamadı")
	}
	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	if user.Password == "" {
		return errors.New("şifre alanı boş olamaz")
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) BulkCreateUsers(ctx context.Context, users []models.User) error {
	return s.repo.BulkCreateUsers(ctx, users)
}

func (s *UserService) UpdateUser(ctx context.Context, id uint, userData *models.User) error {
	currentUserID, ok := ctx.Value("user_id").(uint)
	if !ok || currentUserID == 0 {
		return errors.New("güncelleyen kullanıcı kimliği geçersiz")
	}

	updateData := map[string]interface{}{
		"name":    userData.Name,
		"account": userData.Account,
		"status":  userData.Status,
		"type":    userData.Type,
	}

	return s.repo.UpdateUser(ctx, id, updateData, currentUserID)
}

func (s *UserService) BulkUpdateUsers(ctx context.Context, condition map[string]interface{}, data map[string]interface{}) error {
	currentUserID, ok := ctx.Value("user_id").(uint)
	if !ok || currentUserID == 0 {
		return errors.New("güncelleyen kullanıcı kimliği geçersiz")
	}

	return s.repo.BulkUpdateUsers(ctx, condition, data, currentUserID)
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) BulkDeleteUsers(ctx context.Context, condition map[string]interface{}) error {
	return s.repo.BulkDeleteUsers(ctx, condition)
}

func (s *UserService) GetUserCount() (int64, error) {
	return s.repo.GetUserCount()
}

func (s *UserService) CreateUserWithPassword(ctx context.Context, user *models.User, password string) error {
	if password == "" {
		return errors.New("şifre alanı boş olamaz")
	}
	if err := user.SetPassword(password); err != nil {
		configslog.Log.Error("Şifre oluşturulamadı", zap.Error(err))
		return errors.New("şifre oluşturulurken hata oluştu")
	}
	return s.CreateUser(ctx, user)
}

func (s *UserService) UpdateUserWithPassword(ctx context.Context, id uint, userData *models.User, newPassword string) error {
	currentUserID, ok := ctx.Value("user_id").(uint)
	if !ok || currentUserID == 0 {
		return errors.New("güncelleyen kullanıcı kimliği geçersiz")
	}

	updateData := map[string]interface{}{
		"name":    userData.Name,
		"account": userData.Account,
		"status":  userData.Status,
		"type":    userData.Type,
	}

	if newPassword != "" {
		hashed := models.User{}
		if err := hashed.SetPassword(newPassword); err != nil {
			return errors.New("şifre oluşturulurken hata oluştu")
		}
		updateData["password"] = hashed.Password
	}

	return s.repo.UpdateUser(ctx, id, updateData, currentUserID)
}

var _ IUserService = (*UserService)(nil)
