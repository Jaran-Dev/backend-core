package module

import (
	"context"
	"fmt"

	domainCourse "github.com/Jaran-Dev/backend-core/internal/domain/course"
	domainModule "github.com/Jaran-Dev/backend-core/internal/domain/course/module"

	"github.com/google/uuid"
)

type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type AdminModuleService struct {
	moduleRepo domainModule.Repository
	courseRepo domainCourse.Repository
	txManager  TransactionManager
}

func NewAdminModuleService(mr domainModule.Repository, cr domainCourse.Repository, txm TransactionManager) *AdminModuleService {
	return &AdminModuleService{
		moduleRepo: mr,
		courseRepo: cr,
		txManager:  txm,
	}
}

func (s *AdminModuleService) Create(ctx context.Context, courseID domainCourse.CourseID, title string) (*domainModule.Module, error) {
	courseExists, err := s.courseRepo.ExistsByID(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify course existence: %w", err)
	}
	if !courseExists {
		return nil, domainCourse.ErrCourseNotFound
	}

	currentCount, err := s.moduleRepo.CountByCourseID(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate module position: %w", err)
	}
	nextPosition := int(currentCount) // 0-indexed

	module, err := domainModule.NewModule(courseID, title, nextPosition)
	if err != nil {
		return nil, err
	}

	if err := s.moduleRepo.Create(ctx, module); err != nil {
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	return module, nil
}

func (s *AdminModuleService) UpdateMetadata(ctx context.Context, moduleID domainModule.ModuleID, title *string) (*domainModule.Module, error) {
	module, err := s.moduleRepo.GetByID(ctx, moduleID)
	if err != nil {
		return nil, err
	}

	if title != nil {
		if err := module.UpdateMetadata(*title); err != nil {
			return nil, err
		}
	}

	if err := s.moduleRepo.Update(ctx, module); err != nil {
		return nil, err
	}

	return module, nil
}

func (s *AdminModuleService) Reorder(ctx context.Context, courseID domainCourse.CourseID, requestedIDs []uuid.UUID) ([]*domainModule.Module, error) {
	var updatedModules []*domainModule.Module

	err := s.txManager.Do(ctx, func(txCtx context.Context) error {
		dbModules, err := s.moduleRepo.ListByCourseID(txCtx, courseID)
		if err != nil {
			return err
		}

		if len(dbModules) != len(requestedIDs) {
			return domainModule.ErrInvalidReorderPayload
		}

		requestedPositions := make(map[domainModule.ModuleID]domainModule.Position, len(requestedIDs))
		for index, idStr := range requestedIDs {
			modID := domainModule.ModuleID(idStr)

			if _, exists := requestedPositions[modID]; exists {
				return domainModule.ErrInvalidReorderPayload
			}

			pos, err := domainModule.NewPosition(index)
			if err != nil {
				return err // Should never happen as index is >= 0
			}
			requestedPositions[modID] = pos
		}

		for _, dbModule := range dbModules {
			newPos, exists := requestedPositions[dbModule.ID()]
			if !exists {
				return domainModule.ErrInvalidReorderPayload
			}

			if err := dbModule.UpdatePosition(int(newPos)); err != nil {
				return err
			}
			updatedModules = append(updatedModules, dbModule)
		}

		if err := s.moduleRepo.BulkReorder(txCtx, courseID, requestedPositions); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedModules, nil
}

func (s *AdminModuleService) Delete(ctx context.Context, moduleID domainModule.ModuleID) error {
	// The repository will return ErrModuleNotFound if the ID doesn't exist.
	// The DB schema's ON DELETE CASCADE will automatically remove associated lessons. (must be done differently; the lessons must have the ability to have no module)
	if err := s.moduleRepo.Delete(ctx, moduleID); err != nil {
		return err
	}
	return nil
}
