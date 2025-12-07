package repo_dto

type ConnectionListResponse struct {
	UserId      int64                    `json:"user_id"`
	Connections []ConnectionInfoResponse `json:"connections"`
	TotalCount  int                      `json:"total_count"`
}
