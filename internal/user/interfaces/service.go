package interfaces

import (
	"context"

	"github.com/PeterM45/perfolio-api/internal/common/model"
)

// UserService defines methods for user business logic
type UserService interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	UpdateUser(ctx context.Context, id string, req *model.UpdateUserRequest) (*model.User, error)
	ChangePassword(ctx context.Context, id string, req *model.ChangePasswordRequest) error
	VerifyPassword(ctx context.Context, id string, password string) (bool, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]*model.User, error)

	ToggleFollow(ctx context.Context, req *model.FollowRequest, followerID string) error
	IsFollowing(ctx context.Context, followerID, followingID string) (bool, error)
	GetProfileStats(ctx context.Context, userID string) (*model.ProfileStatsResponse, error)
	GetFollowers(ctx context.Context, userID string, limit, offset int) ([]*model.User, error)
	GetFollowing(ctx context.Context, userID string, limit, offset int) ([]*model.User, error)
}

// PostService defines methods for post business logic
type PostService interface {
	CreatePost(ctx context.Context, userID string, req *model.CreatePostRequest) (*model.Post, error)
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	UpdatePost(ctx context.Context, id string, userID string, req *model.UpdatePostRequest) (*model.Post, error)
	DeletePost(ctx context.Context, id string, userID string) error
	GetUserPosts(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error)
	GetFeed(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error)
}

// WidgetService defines methods for widget business logic
type WidgetService interface {
	GetWidgetByID(ctx context.Context, id string) (*model.Widget, error)
	GetUserWidgets(ctx context.Context, userID string) ([]*model.Widget, error)
	CreateWidget(ctx context.Context, userID string, req *model.CreateWidgetRequest) (*model.Widget, error)
	UpdateWidget(ctx context.Context, id string, userID string, req *model.UpdateWidgetRequest) (*model.Widget, error)
	DeleteWidget(ctx context.Context, id string, userID string) error
	BatchUpdateWidgets(ctx context.Context, userID string, req *model.BatchUpdateWidgetsRequest) error
}
