package dto

type GetRoomResponse struct {
	Id   int64
	Name string
}

type GetRoomMemberResponse struct {
	UserId  int64
	RoomId  int64
	IsAdmin bool
}
