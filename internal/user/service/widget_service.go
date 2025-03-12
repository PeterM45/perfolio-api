package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/platform/cache"
	"github.com/PeterM45/perfolio-api/internal/user/interfaces"
	"github.com/PeterM45/perfolio-api/internal/user/repository"
	"github.com/PeterM45/perfolio-api/internal/widgets"
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

// GetWidgetTypes returns all available widget types
func (s *widgetService) GetWidgetTypes(ctx context.Context) (map[string]model.WidgetType, error) {
	registryTypes := widgets.GetWidgetTypes()
	result := make(map[string]model.WidgetType)

	for key, config := range registryTypes {
		result[key] = model.WidgetType{
			Type:             config.Type,
			DisplayName:      config.DisplayName,
			Description:      config.Description,
			DefaultComponent: config.DefaultComponent,
			DefaultSize: model.WidgetSize{
				W: config.DefaultW,
				H: config.DefaultH,
			},
			MinSize: model.WidgetSize{
				W: config.MinW,
				H: config.MinH,
			},
			MaxSize: model.WidgetSize{
				W: config.MaxW,
				H: config.MaxH,
			},
			Schema:          config.Schema,
			DefaultSettings: config.DefaultSettings,
			Customizations:  config.Customizations,
		}
	}

	return result, nil
}

func (s *widgetService) CreateWidget(ctx context.Context, userID string, req *model.CreateWidgetRequest) (*model.Widget, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	// Verify widget type exists
	if !widgets.ValidWidgetType(req.Type) {
		return nil, apperrors.BadRequest(fmt.Sprintf("invalid widget type: %s", req.Type))
	}

	// Validate settings
	if req.Settings != "" {
		if err := widgets.ValidateWidgetSettings(req.Type, req.Settings); err != nil {
			return nil, apperrors.BadRequest(err.Error())
		}
	} else {
		// Use default settings if none provided
		defaultSettings, err := widgets.GetDefaultSettings(req.Type)
		if err != nil {
			return nil, err
		}
		req.Settings = defaultSettings
	}

	// Verify user exists
	if _, err := s.userService.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	// Get widget type config for defaults
	config, _ := widgets.GetWidgetTypeConfig(req.Type)

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
		IsVisible: true,
		Version:   1,
	}

	// Use component from config if not specified
	if widget.Component == "" {
		widget.Component = config.DefaultComponent
	}

	if req.Settings != "" {
		widget.Settings = &req.Settings
	}

	if req.DisplayName != "" {
		widget.DisplayName = &req.DisplayName
	}

	if err := s.repo.Create(ctx, widget); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("user_widgets:%s", userID))

	return widget, nil
}

// Modify UpdateWidget for optimistic locking and validation
func (s *widgetService) UpdateWidget(ctx context.Context, id string, userID string, req *model.UpdateWidgetRequest) (*model.Widget, error) {
	if id == "" {
		return nil, apperrors.BadRequest("widget ID cannot be empty")
	}

	if req.Version == nil {
		return nil, apperrors.BadRequest("version is required for updates")
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

	// Check version for optimistic locking
	if widget.Version != *req.Version {
		return nil, apperrors.Conflict("widget has been modified, please refresh and try again")
	}

	// Handle type change with settings validation
	if req.Type != nil && *req.Type != widget.Type {
		if !widgets.ValidWidgetType(*req.Type) {
			return nil, apperrors.BadRequest(fmt.Sprintf("invalid widget type: %s", *req.Type))
		}

		// If type changes, validate settings or reset to defaults
		if req.Settings != nil {
			if err := widgets.ValidateWidgetSettings(*req.Type, *req.Settings); err != nil {
				return nil, apperrors.BadRequest(err.Error())
			}
		} else {
			// Use default settings for new type
			defaultSettings, err := widgets.GetDefaultSettings(*req.Type)
			if err != nil {
				return nil, err
			}
			req.Settings = &defaultSettings
		}

		widget.Type = *req.Type
	} else if req.Settings != nil {
		// If just settings change, validate against current type
		if err := widgets.ValidateWidgetSettings(widget.Type, *req.Settings); err != nil {
			return nil, apperrors.BadRequest(err.Error())
		}
	}

	// Update widget fields
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

	if req.DisplayName != nil {
		widget.DisplayName = req.DisplayName
	}

	if req.IsVisible != nil {
		widget.IsVisible = *req.IsVisible
	}

	// Increment version
	widget.Version++

	// Save updates
	if err := s.repo.Update(ctx, widget); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("widget:%s", id))
	s.cache.Delete(fmt.Sprintf("user_widgets:%s", userID))

	return widget, nil
}

// Modify BatchUpdateWidgets for optimistic locking
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

	// Perform batch update with version checking
	if err := s.repo.BatchUpdatePositions(ctx, updates); err != nil {
		return err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("user_widgets:%s", userID))

	return nil
}
