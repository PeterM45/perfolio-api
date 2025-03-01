package model

import (
	"time"
)

// Follow represents a follow relationship between users
type Follow struct {
	FollowerID  string    `json:"followerId"`
	FollowingID string    `json:"followingId"`
	CreatedAt   time.Time `json:"createdAt"`

	// Optional joined fields
	Follower  *User `json:"follower,omitempty"`
	Following *User `json:"following,omitempty"`
}

// FollowRequest is used to follow/unfollow a user
type FollowRequest struct {
	FollowingID string `json:"followingId" validate:"required"`
	Action      string `json:"action" validate:"required,oneof=follow unfollow"`
}

// ProfileStatsResponse is the response for profile stats
type ProfileStatsResponse struct {
	FollowerCount  int `json:"followerCount"`
	FollowingCount int `json:"followingCount"`
}

// FollowersResponse is the response for getting followers
type FollowersResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}
