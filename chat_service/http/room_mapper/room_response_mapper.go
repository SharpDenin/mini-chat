package room_mapper

import (
	"chat_service/http/api_dto"
	"chat_service/internal/room/service/dto"
)

func GetRoomToHandlerDto(r *dto.GetRoomResponse) *api_dto.GetRoomResponse {
	return &api_dto.GetRoomResponse{
		Id:   r.Id,
		Name: r.Name,
	}
}

func GetRoomListToHandlerDto(r []*dto.GetRoomResponse, limit, offset int) *api_dto.GetRoomListResponse {
	roomList := make([]*api_dto.GetRoomResponse, len(r))
	for i, room := range r {
		roomList[i] = GetRoomToHandlerDto(room)
	}
	return &api_dto.GetRoomListResponse{
		RoomList: roomList,
		Limit:    limit,
		Offset:   offset,
	}
}

func GetRoomMemberToHandlerDto(r *dto.GetRoomMemberResponse) *api_dto.GetRoomMemberResponse {
	return &api_dto.GetRoomMemberResponse{
		UserId:  r.UserId,
		IsAdmin: r.IsAdmin,
	}
}

func GetRoomMemberListToHandlerDto(roomId int64, r []*dto.GetRoomMemberResponse) *api_dto.GetRoomMemberListResponse {
	roomMemberList := make([]*api_dto.GetRoomMemberResponse, len(r))
	for i, roomMember := range r {
		roomMemberList[i] = GetRoomMemberToHandlerDto(roomMember)
	}
	return &api_dto.GetRoomMemberListResponse{
		RoomId:         roomId,
		RoomMemberList: roomMemberList,
	}
}
