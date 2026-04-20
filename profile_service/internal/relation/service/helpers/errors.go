package helpers

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrFriendRequestNotFound = errors.New("friend request not found")
	ErrUsersNotFriends       = errors.New("users are not friends")
	ErrBlockNotFound         = errors.New("block not found")
)
