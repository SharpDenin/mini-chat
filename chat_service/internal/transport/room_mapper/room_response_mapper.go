package room_mapper

import (
	"chat_service/internal/service/dto"
	"chat_service/internal/transport/api_dto"
)

func GetRoomToHandlerDto(r *dto.GetRoomResponse) *api_dto.GetRoomResponse {
	return &api_dto.GetRoomResponse{
		Id:   r.Id,
		Name: r.Name,
	}
}

func GetRoomListToHandlerDto(r []*dto.GetRoomResponse, limit, offset int) *api_dto.GetRoomList {
	roomList := make([]*api_dto.GetRoomResponse, len(r))
	for i, room := range r {
		roomList[i] = GetRoomToHandlerDto(room)
	}
	return &api_dto.GetRoomList{
		RoomList: roomList,
		Limit:    limit,
		Offset:   offset,
	}
}
