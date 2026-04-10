package api_dto

type SendFriendRequestRequest struct {
	ReceiverId int64  `json:"receiver_id" binding:"required,gt=0"`
	Message    string `json:"message" binding:"max=500"`
}

type AnswerFriendRequestRequest struct {
	Accept bool `json:"accept" binding:"required"`
}

type RequestStateResponse struct {
	Status string `json:"status"` // pending, accepted, rejected, none
}

type FriendListResponse struct {
	Friends []FriendView `json:"friends"`
	Total   int          `json:"total"`
}

type FriendView struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at,omitempty"`
}

type BlockUserRequest struct {
	BlockedId int64  `json:"blocked_id" binding:"required,gt=0"`
	Reason    string `json:"reason" binding:"max=500"`
}

type UnblockUserRequest struct {
	BlockedId int64 `json:"blocked_id" binding:"required,gt=0"`
}

type BlockInfoResponse struct {
	IsBlocked bool  `json:"is_blocked"`
	BlockerId int64 `json:"blocker_id,omitempty"`
	BlockedId int64 `json:"blocked_id,omitempty"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AreFriendsResponse struct {
	AreFriends bool `json:"are_friends"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
