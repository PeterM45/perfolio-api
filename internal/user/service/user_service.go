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
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo      repository.UserRepository
	cache     cache.Cache
	validator validator.Validator
	logger    logger.Logger
}

// NewUserService creates a new UserService
func NewUserService(repo repository.UserRepository, cache cache.Cache, logger logger.Logger) interfaces.UserService {
	return &userService{
		repo:      repo,
		cache:     cache,
		validator: validator.NewValidator(),
		logger:    logger,
	}
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, apperrors.BadRequest("user ID cannot be empty")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("user:%s", id)
	if cachedUser, found := s.cache.Get(cacheKey); found {
		s.logger.Debug().Str("user_id", id).Msg("User found in cache")
		return cachedUser.(*model.User), nil
	}

	// Get from repository
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(cacheKey, user, 5*time.Minute)

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	if username == "" {
		return nil, apperrors.BadRequest("username cannot be empty")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("user:username:%s", username)
	if cachedUser, found := s.cache.Get(cacheKey); found {
		s.logger.Debug().Str("username", username).Msg("User found in cache by username")
		return cachedUser.(*model.User), nil
	}

	// Get from repository
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(cacheKey, user, 5*time.Minute)
	s.cache.Set(fmt.Sprintf("user:%s", user.ID), user, 5*time.Minute)

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if email == "" {
		return nil, apperrors.BadRequest("email cannot be empty")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("user:email:%s", email)
	if cachedUser, found := s.cache.Get(cacheKey); found {
		s.logger.Debug().Str("email", email).Msg("User found in cache by email")
		return cachedUser.(*model.User), nil
	}

	// Get from repository
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(cacheKey, user, 5*time.Minute)
	s.cache.Set(fmt.Sprintf("user:%s", user.ID), user, 5*time.Minute)

	return user, nil
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	// Check if username is available
	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, apperrors.BadRequest("username already taken")
	}

	// Check if email is available for custom auth
	if req.AuthProvider == model.AuthProviderCustom && req.Email != "" {
		existingUser, err := s.repo.GetByEmail(ctx, req.Email)
		if err == nil && existingUser != nil {
			return nil, apperrors.BadRequest("email already registered")
		}
	}

	// Generate UUID for custom auth if not provided
	if req.ID == "" && req.AuthProvider == model.AuthProviderCustom {
		// Use your UUID generation logic here
		// For example, using github.com/google/uuid:
		// req.ID = uuid.New().String()

		// Or you can let your repository handle ID generation
	}

	// Convert to user model
	user := &model.User{
		ID:           req.ID,
		Email:        req.Email,
		Username:     req.Username,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Bio:          req.Bio,
		AuthProvider: req.AuthProvider,
		ImageURL:     req.ImageURL,
		PasswordHash: req.PasswordHash, // Set from hashed password
		IsActive:     true,
	}

	// Save to repository
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, id string, req *model.UpdateUserRequest) (*model.User, error) {
	if id == "" {
		return nil, apperrors.BadRequest("user ID cannot be empty")
	}

	if err := s.validator.Validate(req); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	// Check if user exists
	existingUser, err := s.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.Username != nil {
		// Check if new username is available
		if *req.Username != existingUser.Username {
			user, err := s.repo.GetByUsername(ctx, *req.Username)
			if err == nil && user != nil {
				return nil, apperrors.BadRequest("username already taken")
			}
			updates["username"] = *req.Username
		}
	}

	// Add other fields if provided
	if req.FirstName != nil {
		updates["firstName"] = *req.FirstName
	}

	if req.LastName != nil {
		updates["lastName"] = *req.LastName
	}

	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}

	if req.ImageURL != nil {
		updates["imageUrl"] = *req.ImageURL
	}

	if req.IsActive != nil {
		updates["isActive"] = *req.IsActive
	}

	// Handle password update if provided
	if req.Password != nil && *req.Password != "" {
		// Verify current password if provided
		if req.CurrentPassword != nil && *req.CurrentPassword != "" {
			isValid, err := s.VerifyPassword(ctx, id, *req.CurrentPassword)
			if err != nil {
				return nil, err
			}
			if !isValid {
				return nil, apperrors.BadRequest("current password is incorrect")
			}
		} else {
			// Only allow password change without verification for OAuth users converting to custom auth
			if existingUser.AuthProvider != model.AuthProviderOAuth &&
				existingUser.AuthProvider != model.AuthProviderGoogle &&
				existingUser.AuthProvider != model.AuthProviderGithub {
				return nil, apperrors.BadRequest("current password verification required")
			}
			// Update auth provider if converting from OAuth
			updates["authProvider"] = model.AuthProviderCustom
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to hash password")
			return nil, apperrors.InternalError("password update failed")
		}
		updates["passwordHash"] = string(hashedPassword)
	}

	// Only update if there are changes
	if len(updates) > 0 {
		err = s.repo.Update(ctx, id, updates)
		if err != nil {
			return nil, err
		}

		// Invalidate cache
		s.cache.Delete(fmt.Sprintf("user:%s", id))
		if req.Username != nil {
			s.cache.Delete(fmt.Sprintf("user:username:%s", existingUser.Username))
		}
		s.cache.Delete(fmt.Sprintf("user:email:%s", existingUser.Email))

		// Get updated user
		return s.repo.GetByID(ctx, id)
	}

	return existingUser, nil
}

// ChangePassword changes a user's password with verification
func (s *userService) ChangePassword(ctx context.Context, id string, req *model.ChangePasswordRequest) error {
	if err := s.validator.Validate(req); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	// Verify current password
	isValid, err := s.VerifyPassword(ctx, id, req.CurrentPassword)
	if err != nil {
		return err
	}

	if !isValid {
		return apperrors.BadRequest("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to hash password")
		return apperrors.InternalError("password update failed")
	}

	// Update the password hash
	updates := map[string]interface{}{
		"passwordHash": string(hashedPassword),
	}

	err = s.repo.Update(ctx, id, updates)
	if err != nil {
		return err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("user:%s", id))

	return nil
}

// VerifyPassword verifies a user's password
func (s *userService) VerifyPassword(ctx context.Context, id string, password string) (bool, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	if user.PasswordHash == "" {
		return false, apperrors.BadRequest("user has no password set")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil, nil
}

// SearchUsers searches for users
func (s *userService) SearchUsers(ctx context.Context, query string, limit int) ([]*model.User, error) {
	if query == "" {
		return nil, apperrors.BadRequest("search query cannot be empty")
	}

	if limit <= 0 {
		limit = 10
	}

	return s.repo.Search(ctx, query, limit)
}

// ToggleFollow toggles a follow relationship
func (s *userService) ToggleFollow(ctx context.Context, req *model.FollowRequest, followerID string) error {
	if err := s.validator.Validate(req); err != nil {
		return apperrors.BadRequest(err.Error())
	}

	// Can't follow yourself
	if followerID == req.FollowingID {
		return apperrors.BadRequest("cannot follow yourself")
	}

	// Verify both users exist
	if _, err := s.GetUserByID(ctx, followerID); err != nil {
		return err
	}

	if _, err := s.GetUserByID(ctx, req.FollowingID); err != nil {
		return err
	}

	// Perform requested action
	var err error
	if req.Action == "follow" {
		err = s.repo.AddFollow(ctx, followerID, req.FollowingID)
	} else {
		err = s.repo.RemoveFollow(ctx, followerID, req.FollowingID)
	}

	if err != nil {
		return err
	}

	// Invalidate caches
	s.cache.Delete(fmt.Sprintf("follows:%s:%s", followerID, req.FollowingID))
	s.cache.Delete(fmt.Sprintf("follower_count:%s", req.FollowingID))
	s.cache.Delete(fmt.Sprintf("following_count:%s", followerID))

	return nil
}

// IsFollowing checks if a user is following another
func (s *userService) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("follows:%s:%s", followerID, followingID)
	if cachedResult, found := s.cache.Get(cacheKey); found {
		return cachedResult.(bool), nil
	}

	isFollowing, err := s.repo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return false, err
	}

	// Cache the result
	s.cache.Set(cacheKey, isFollowing, 5*time.Minute)

	return isFollowing, nil
}

// GetProfileStats gets follower and following counts
func (s *userService) GetProfileStats(ctx context.Context, userID string) (*model.ProfileStatsResponse, error) {
	// Check if user exists
	if _, err := s.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	// Get follower count (with cache)
	var followerCount int
	followerCacheKey := fmt.Sprintf("follower_count:%s", userID)

	if cachedCount, found := s.cache.Get(followerCacheKey); found {
		followerCount = cachedCount.(int)
	} else {
		var err error
		followerCount, err = s.repo.GetFollowerCount(ctx, userID)
		if err != nil {
			return nil, err
		}
		s.cache.Set(followerCacheKey, followerCount, 5*time.Minute)
	}

	// Get following count (with cache)
	var followingCount int
	followingCacheKey := fmt.Sprintf("following_count:%s", userID)

	if cachedCount, found := s.cache.Get(followingCacheKey); found {
		followingCount = cachedCount.(int)
	} else {
		var err error
		followingCount, err = s.repo.GetFollowingCount(ctx, userID)
		if err != nil {
			return nil, err
		}
		s.cache.Set(followingCacheKey, followingCount, 5*time.Minute)
	}

	return &model.ProfileStatsResponse{
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
	}, nil
}

// GetFollowers gets users who follow the given user
func (s *userService) GetFollowers(ctx context.Context, userID string, limit, offset int) ([]*model.User, error) {
	// Check if user exists
	if _, err := s.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	return s.repo.GetFollowers(ctx, userID, limit, offset)
}

// GetFollowing gets users the given user follows
func (s *userService) GetFollowing(ctx context.Context, userID string, limit, offset int) ([]*model.User, error) {
	// Check if user exists
	if _, err := s.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	return s.repo.GetFollowing(ctx, userID, limit, offset)
}
