package api_dto

type SearchRoomRequest struct {
	SearchQuery string `json:"searchQuery" binding:"omitempty,max=100"`
	Limit       int    `json:"limit" binding:"required, min=0, max=100"`
	Offset      int    `json:"offset" binding:"required, min=0, max=100"`
}

type CreateRoomRequest struct {
	Name string `json:"name" binding:"required, min=1, max=255"`
}

type UpdateRoomRequest struct {
	Name string `json:"name" binding:"omitempty,max=255"`
}
