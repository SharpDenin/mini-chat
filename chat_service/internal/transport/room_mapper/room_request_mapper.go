package room_mapper

import (
	"chat_service/internal/room/room_service/dto"
	"chat_service/internal/transport/api_dto"
)

func SearchQueryToServiceFilter(r *api_dto.SearchRoomRequest) *dto.SearchFilter {
	return &dto.SearchFilter{
		Search: r.SearchQuery,
		Limit:  r.Limit,
		Offset: r.Offset,
	}
}
