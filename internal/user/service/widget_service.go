package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/platform/cache"
	"github.com/PeterM45/perfolio-api/internal/user/interfaces"
	"github.com/PeterM45/perfolio-api/internal/user/repository"
	"github.com/PeterM45/perfolio-api/pkg/apperrors"
	"github.com/PeterM45/perfolio-api/pkg/logger"
	"github.com/PeterM45/perfolio-api/pkg/validator"
	"github.com/google/uuid"
)

type widgetService struct {
	repo        repository.WidgetRepository
	userService interfaces.UserService
	cache       cache.Cache
	validator   validator.Validator
	logger      logger.Logger
}

// NewWidgetService creates a new WidgetService
func NewWidgetService(
	repo repository.WidgetRepository,
	userService interfaces.UserService,
	cache cache.Cache,
	logger logger.Logger,
) interfaces.WidgetService {
	return &widgetService{
		repo:        repo,
		userService: userService,
		cache:       cache,
		validator:   validator.NewValidator(),
		logger:      logger,
	}
}

// GetWidgetByID retrieves a widget by ID
func (s *widgetService) GetWidgetByID(ctx context.Context, id string) (*model.Widget, error) {
	if id == "" {
		return nil, apperrors.BadRequest("widget ID cannot be empty")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("widget:%s", id)
	if cachedWidget, found := s.cache.Get(cacheKey); found {
		s.logger.Debug().Str("widget_id", id).Msg("Widget found in cache")
		return cachedWidget.(*model.Widget), nil
	}

	widget, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(cacheKey, widget, 5*time.Minute)

	return widget, nil
}

// GetUserWidgets gets all widgets for a user
func (s *widgetService) GetUserWidgets(ctx context.Context, userID string) ([]*model.Widget, error) {
	if userID == "" {
		return nil, apperrors.BadRequest("user ID cannot be empty")
	}

	// Verify user exists
	if _, err := s.userService.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := fmt.Sprintf("user_widgets:%s", userID)
	if cachedWidgets, found := s.cache.Get(cacheKey); found {
		s.logger.Debug().Str("user_id", userID).Msg("User widgets found in cache")
		return cachedWidgets.([]*model.Widget), nil
	}

	widgets, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(cacheKey, widgets, 5*time.Minute)

	return widgets, nil
}

// CreateWidget creates a new widget
func (s *widgetService) CreateWidget(ctx context.Context, userID string, req *model.CreateWidgetRequest) (*model.Widget, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	// Verify user exists
	if _, err := s.userService.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	// Create widget
	widget := &model.Widget{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      req.Type,
		Component: req.Component,
		X:         req.X,
		Y:         req.Y,
		W:         req.W,
		H:         req.H,
	}

	if req.Settings != "" {
		widget.Settings = &req.Settings
	}

	if err := s.repo.Create(ctx, widget); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("user_widgets:%s", userID))

	return widget, nil
}

// UpdateWidget updates an existing widget
func (s *widgetService) UpdateWidget(ctx context.Context, id string, userID string, req *model.UpdateWidgetRequest) (*model.Widget, error) {
	if id == "" {
		return nil, apperrors.BadRequest("widget ID cannot be empty")
	}

	if err := s.validator.Validate(req); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	// Get existing widget
	widget, err := s.GetWidgetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user owns the widget
	if widget.UserID != userID {
		return nil, apperrors.Forbidden("you don't have permission to update this widget")
	}

	// Update widget fields
	if req.Type != nil {
		widget.Type = *req.Type
	}

	if req.Component != nil {
		widget.Component = *req.Component
	}

	if req.X != nil {
		widget.X = *req.X
	}

	if req.Y != nil {
		widget.Y = *req.Y
	}

	if req.W != nil {
		widget.W = *req.W
	}

	if req.H != nil {
		widget.H = *req.H
	}

	if req.Settings != nil {
		widget.Settings = req.Settings
	}

	// Save updates
	if err := s.repo.Update(ctx, widget); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("widget:%s", id))
	s.cache.Delete(fmt.Sprintf("user_widgets:%s", userID))

	return widget, nil
}

// DeleteWidget deletes a widget
func (s *widgetService) DeleteWidget(ctx context.Context, id string, userID string) error {
	if id == "" {
		return apperrors.BadRequest("widget ID cannot be empty")
	}

	// Get existing widget
	widget, err := s.GetWidgetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user owns the widget
	if widget.UserID != userID {
		return apperrors.Forbidden("you don't have permission to delete this widget")
	}

	// Delete widget
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("widget:%s", id))
	s.cache.Delete(fmt.Sprintf("user_widgets:%s", userID))

	return nil
}

// BatchUpdateWidgets updates multiple widget positions
func (s *widgetService) BatchUpdateWidgets(ctx context.Context, userID string, req *model.BatchUpdateWidgetsRequest) error {
	if err := s.validator.Validate(req); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	// Verify user exists
	if _, err := s.userService.GetUserByID(ctx, userID); err != nil {
		return err
	}

	// Verify all widgets belong to user
	updates := make([]*model.WidgetPositionUpdate, 0, len(req.Updates))
	for _, update := range req.Updates {
		if update.UserID != userID {
			return apperrors.Forbidden("you can only update your own widgets")
		}
		updates = append(updates, &update)
	}

	// Perform batch update
	if err := s.repo.BatchUpdatePositions(ctx, updates); err != nil {
		return err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("user_widgets:%s", userID))

	return nil
}
