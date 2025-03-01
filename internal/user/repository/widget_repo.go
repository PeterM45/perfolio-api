package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/platform/database"
	"github.com/PeterM45/perfolio-api/pkg/apperrors"
	"github.com/google/uuid"
)

// WidgetRepository defines methods to interact with widget data
type WidgetRepository interface {
	GetByID(ctx context.Context, id string) (*model.Widget, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Widget, error)
	Create(ctx context.Context, widget *model.Widget) error
	Update(ctx context.Context, widget *model.Widget) error
	Delete(ctx context.Context, id string) error
	BatchUpdatePositions(ctx context.Context, updates []*model.WidgetPositionUpdate) error
}

type widgetRepository struct {
	db *database.DB
}

// NewWidgetRepository creates a new WidgetRepository
func NewWidgetRepository(db *database.DB) WidgetRepository {
	return &widgetRepository{
		db: db,
	}
}

// GetByID fetches a widget by ID
func (r *widgetRepository) GetByID(ctx context.Context, id string) (*model.Widget, error) {
	query := `
		SELECT 
			id, user_id, type, component, x, y, w, h, settings
		FROM 
			widgets
		WHERE 
			id = $1
	`

	var widget model.Widget
	var settings sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&widget.ID,
		&widget.UserID,
		&widget.Type,
		&widget.Component,
		&widget.X,
		&widget.Y,
		&widget.W,
		&widget.H,
		&settings,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("widget", id)
		}
		return nil, fmt.Errorf("query widget: %w", err)
	}

	if settings.Valid {
		widget.Settings = &settings.String
	}

	return &widget, nil
}

// GetByUserID fetches widgets for a user
func (r *widgetRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Widget, error) {
	query := `
		SELECT 
			id, user_id, type, component, x, y, w, h, settings
		FROM 
			widgets
		WHERE 
			user_id = $1
		ORDER BY
			y ASC, x ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query user widgets: %w", err)
	}
	defer rows.Close()

	var widgets []*model.Widget

	for rows.Next() {
		var widget model.Widget
		var settings sql.NullString

		err := rows.Scan(
			&widget.ID,
			&widget.UserID,
			&widget.Type,
			&widget.Component,
			&widget.X,
			&widget.Y,
			&widget.W,
			&widget.H,
			&settings,
		)

		if err != nil {
			return nil, fmt.Errorf("scan widget row: %w", err)
		}

		if settings.Valid {
			widget.Settings = &settings.String
		}

		widgets = append(widgets, &widget)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return widgets, nil
}

// Create adds a new widget
func (r *widgetRepository) Create(ctx context.Context, widget *model.Widget) error {
	if widget.ID == "" {
		widget.ID = uuid.New().String()
	}

	query := `
		INSERT INTO widgets (
			id, user_id, type, component, x, y, w, h, settings
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	var settingsSQL sql.NullString
	if widget.Settings != nil {
		settingsSQL = sql.NullString{String: *widget.Settings, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		widget.ID,
		widget.UserID,
		widget.Type,
		widget.Component,
		widget.X,
		widget.Y,
		widget.W,
		widget.H,
		settingsSQL,
	)

	if err != nil {
		return fmt.Errorf("create widget: %w", err)
	}

	return nil
}

// Update updates a widget
func (r *widgetRepository) Update(ctx context.Context, widget *model.Widget) error {
	query := `
		UPDATE widgets
		SET 
			type = $1, 
			component = $2, 
			x = $3, 
			y = $4, 
			w = $5, 
			h = $6, 
			settings = $7
		WHERE 
			id = $8 AND user_id = $9
		RETURNING id
	`

	var settingsSQL sql.NullString
	if widget.Settings != nil {
		settingsSQL = sql.NullString{String: *widget.Settings, Valid: true}
	}

	var id string
	err := r.db.QueryRowContext(ctx, query,
		widget.Type,
		widget.Component,
		widget.X,
		widget.Y,
		widget.W,
		widget.H,
		settingsSQL,
		widget.ID,
		widget.UserID,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound("widget", widget.ID)
		}
		return fmt.Errorf("update widget: %w", err)
	}

	return nil
}

// Delete removes a widget
func (r *widgetRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM widgets WHERE id = $1 RETURNING id`

	var deletedID string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&deletedID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound("widget", id)
		}
		return fmt.Errorf("delete widget: %w", err)
	}

	return nil
}

// BatchUpdatePositions updates multiple widget positions in a transaction
func (r *widgetRepository) BatchUpdatePositions(ctx context.Context, updates []*model.WidgetPositionUpdate) error {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement for repeated use
	stmt, err := tx.PrepareContext(ctx, `
		UPDATE widgets
		SET x = $1, y = $2, w = $3, h = $4
		WHERE id = $5 AND user_id = $6
		RETURNING id
	`)

	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute updates
	for _, update := range updates {
		var id string
		err := stmt.QueryRowContext(ctx,
			update.X,
			update.Y,
			update.W,
			update.H,
			update.ID,
			update.UserID,
		).Scan(&id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return apperrors.NotFound("widget", update.ID)
			}
			return fmt.Errorf("update widget position: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
