package helpers

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrFriendRequestNotFound = errors.New("friend request not found")
	ErrUsersNotFriends       = errors.New("users are not friends")
	ErrBlockNotFound         = errors.New("block not found")
	ErrUsersAreEquals        = errors.New("user_id matches yours")
	ErrAlreadyFriends        = errors.New("users are already friends")
	ErrFriendRequestExists   = errors.New("friend request already exists")
	ErrUserAlreadyBlocked    = errors.New("user already blocked")
	ErrUserNotBlocked        = errors.New("user not blocked")
	ErrBlockedByUser         = errors.New("cannot send friend request: user is blocked")
	ErrCannotBlockYourself   = errors.New("cannot block yourself")
	ErrCannotUnblockYourself = errors.New("cannot unblock yourself")
	ErrCannotFriendYourself  = errors.New("cannot send friend request to yourself")
	ErrCannotDeleteYourself  = errors.New("cannot remove yourself from friend list")
)
