package services

import (
	"context"
	"errors"

	"davet.link/pkg/queryparams"

	"davet.link/configs/configslog"
	"go.uber.org/zap"
)

const contextUserIDKey = "user_id"

type IBaseService[T any] interface {
	GetAll(params queryparams.ListParams) (*queryparams.PaginatedResult, error)
	GetByID(id uint) (*T, error)
	Create(ctx context.Context, entity *T) error
	BulkCreate(ctx context.Context, entities []T) error
	Update(ctx context.Context, id uint, data map[string]interface{}) error
	BulkUpdate(ctx context.Context, condition map[string]interface{}, data map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	BulkDelete(ctx context.Context, condition map[string]interface{}) error
	GetCount() (int64, error)
}

type BaseService[T any] struct {
	repo IBaseRepository[T]
}

func NewBaseService[T any](repo IBaseRepository[T]) *BaseService[T] {
	return &BaseService[T]{repo: repo}
}

func (s *BaseService[T]) GetAll(params queryparams.ListParams) (*queryparams.PaginatedResult, error) {
	entities, totalCount, err := s.repo.GetAll(params)
	if err != nil {
		configslog.Log.Error("Kayıtlar alınamadı", zap.Error(err))
		return nil, errors.New("kayıtlar getirilirken bir hata oluştu")
	}

	result := &queryparams.PaginatedResult{
		Data: entities,
		Meta: queryparams.PaginationMeta{
			CurrentPage: params.Page,
			PerPage:     params.PerPage,
			TotalItems:  totalCount,
			TotalPages:  queryparams.CalculateTotalPages(totalCount, params.PerPage),
		},
	}
	return result, nil
}

func (s *BaseService[T]) GetByID(id uint) (*T, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		configslog.Log.Warn("Kayıt bulunamadı", zap.Uint("id", id), zap.Error(err))
		return nil, errors.New("kayıt bulunamadı")
	}
	return entity, nil
}

func (s *BaseService[T]) Create(ctx context.Context, entity *T) error {
	return s.repo.Create(ctx, entity)
}

func (s *BaseService[T]) BulkCreate(ctx context.Context, entities []T) error {
	return s.repo.BulkCreate(ctx, entities)
}

func (s *BaseService[T]) Update(ctx context.Context, id uint, data map[string]interface{}) error {
	currentUserID, ok := ctx.Value(contextUserIDKey).(uint)
	if !ok || currentUserID == 0 {
		return errors.New("güncelleyen kullanıcı kimliği geçersiz")
	}

	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("kayıt bulunamadı")
	}

	return s.repo.Update(ctx, id, data, currentUserID)
}

func (s *BaseService[T]) BulkUpdate(ctx context.Context, condition map[string]interface{}, data map[string]interface{}) error {
	currentUserID, ok := ctx.Value(contextUserIDKey).(uint)
	if !ok || currentUserID == 0 {
		return errors.New("güncelleyen kullanıcı kimliği geçersiz")
	}

	return s.repo.BulkUpdate(ctx, condition, data, currentUserID)
}

func (s *BaseService[T]) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *BaseService[T]) BulkDelete(ctx context.Context, condition map[string]interface{}) error {
	return s.repo.BulkDelete(ctx, condition)
}

func (s *BaseService[T]) GetCount() (int64, error) {
	return s.repo.GetCount()
}

var _ IBaseService[any] = (*BaseService[any])(nil)
