package api_dto

type SendFriendRequestRequest struct {
	Message string `json:"message" binding:"max=500"`
}

type SendFriendRequestResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RequestId int64  `json:"request_id"`
}

type AnswerFriendRequestRequest struct {
	Accept bool `json:"accept"`
}

type RequestStateResponse struct {
	RequestId int64  `json:"requestId"`
	Status    string `json:"status"` // pending, accepted, rejected, none
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
	Reason string `json:"reason" binding:"max=500"`
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
