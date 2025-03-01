package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PeterM45/perfolio-api/internal/common/interfaces"
	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/platform/cache"
	"github.com/PeterM45/perfolio-api/internal/user/repository"
	"github.com/PeterM45/perfolio-api/pkg/apperrors"
	"github.com/PeterM45/perfolio-api/pkg/logger"
	"github.com/PeterM45/perfolio-api/pkg/validator"
	"github.com/google/uuid"
)

// PostService defines methods for post business logic
type PostService interface {
	CreatePost(ctx context.Context, userID string, req *model.CreatePostRequest) (*model.Post, error)
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	UpdatePost(ctx context.Context, id string, userID string, req *model.UpdatePostRequest) (*model.Post, error)
	DeletePost(ctx context.Context, id string, userID string) error
	GetUserPosts(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error)
	GetFeed(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error)
}

type postService struct {
	repo        repository.PostRepository
	userService interfaces.UserService
	cache       cache.Cache
	validator   validator.Validator
	logger      logger.Logger
}

// NewPostService creates a new PostService
func NewPostService(
	repo repository.PostRepository,
	userService interfaces.UserService,
	cache cache.Cache,
	logger logger.Logger,
) PostService {
	return &postService{
		repo:        repo,
		userService: userService,
		cache:       cache,
		validator:   validator.NewValidator(),
		logger:      logger,
	}
}

// CreatePost creates a new post
func (s *postService) CreatePost(ctx context.Context, userID string, req *model.CreatePostRequest) (*model.Post, error) {
	if err := s.validator.Validate(req); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	// Verify user exists
	if _, err := s.userService.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	// Create post
	post := &model.Post{
		ID:         uuid.New().String(),
		UserID:     userID,
		Content:    req.Content,
		EmbedURLs:  req.EmbedURLs,
		Hashtags:   req.Hashtags,
		Visibility: model.VisibilityPublic, // Default visibility
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}

	// Invalidate feed caches
	s.cache.Delete(fmt.Sprintf("user_posts:%s", userID))

	return post, nil
}

// GetPostByID retrieves a post by ID
func (s *postService) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	if id == "" {
		return nil, apperrors.BadRequest("post ID cannot be empty")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("post:%s", id)
	if cachedPost, found := s.cache.Get(cacheKey); found {
		s.logger.Debug().Str("post_id", id).Msg("Post found in cache")
		return cachedPost.(*model.Post), nil
	}

	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(cacheKey, post, 5*time.Minute)

	return post, nil
}

// UpdatePost updates an existing post
func (s *postService) UpdatePost(ctx context.Context, id string, userID string, req *model.UpdatePostRequest) (*model.Post, error) {
	if id == "" {
		return nil, apperrors.BadRequest("post ID cannot be empty")
	}

	if err := s.validator.Validate(req); err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}

	// Get existing post
	post, err := s.GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user owns the post
	if post.UserID != userID {
		return nil, apperrors.Forbidden("you don't have permission to update this post")
	}

	// Update post fields
	post.Content = req.Content
	post.EmbedURLs = req.EmbedURLs
	post.Hashtags = req.Hashtags

	// Save updates
	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("post:%s", id))
	s.cache.Delete(fmt.Sprintf("user_posts:%s", userID))

	return post, nil
}

// DeletePost deletes a post
func (s *postService) DeletePost(ctx context.Context, id string, userID string) error {
	if id == "" {
		return apperrors.BadRequest("post ID cannot be empty")
	}

	// Get existing post
	post, err := s.GetPostByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if user owns the post
	if post.UserID != userID {
		return apperrors.Forbidden("you don't have permission to delete this post")
	}

	// Delete post
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("post:%s", id))
	s.cache.Delete(fmt.Sprintf("user_posts:%s", userID))

	return nil
}

// GetUserPosts gets posts by a user
func (s *postService) GetUserPosts(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error) {
	if userID == "" {
		return nil, apperrors.BadRequest("user ID cannot be empty")
	}

	// Verify user exists
	if _, err := s.userService.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	// Check cache for list of posts
	cacheKey := fmt.Sprintf("user_posts:%s:%d:%d", userID, limit, offset)
	if cachedPosts, found := s.cache.Get(cacheKey); found {
		s.logger.Debug().Str("user_id", userID).Msg("User posts found in cache")
		return cachedPosts.([]*model.Post), nil
	}

	posts, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.Set(cacheKey, posts, 2*time.Minute) // Shorter TTL for lists

	return posts, nil
}

// GetFeed gets posts for user's feed
func (s *postService) GetFeed(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error) {
	if userID == "" {
		return nil, apperrors.BadRequest("user ID cannot be empty")
	}

	// Verify user exists
	if _, err := s.userService.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}

	// Get following users
	following, err := s.userService.GetFollowing(ctx, userID, 500, 0) // Get a reasonable number of followed users
	if err != nil {
		return nil, err
	}

	// Build array of user IDs to fetch posts from (include self)
	userIDs := []string{userID}
	for _, user := range following {
		userIDs = append(userIDs, user.ID)
	}

	// Check cache
	cacheKey := fmt.Sprintf("feed:%s:%d:%d", userID, limit, offset)
	if cachedFeed, found := s.cache.Get(cacheKey); found {
		s.logger.Debug().Str("user_id", userID).Msg("Feed found in cache")
		return cachedFeed.([]*model.Post), nil
	}

	// Get posts
	posts, err := s.repo.GetFeed(ctx, userIDs, limit, offset)
	if err != nil {
		return nil, err
	}

	// Store in cache (short TTL for feeds)
	s.cache.Set(cacheKey, posts, 1*time.Minute)

	return posts, nil
}
