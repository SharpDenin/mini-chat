package room_mapper

import (
	"chat_service/http/api_dto"
	"chat_service/internal/room/service/dto"
)

func SearchQueryToServiceFilter(r *api_dto.SearchRoomRequest) *dto.SearchFilter {
	return &dto.SearchFilter{
		Search: r.SearchQuery,
		Limit:  r.Limit,
		Offset: r.Offset,
	}
}
