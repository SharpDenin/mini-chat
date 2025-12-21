package api_dto

type GetRoomResponse struct {
	Id   int64  `json:"Id"`
	Name string `json:"Name"`
}

type GetRoomListResponse struct {
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	RoomList []*GetRoomResponse `json:"roomList"`
}

type GetRoomMemberResponse struct {
	UserId  int64 `json:"userId"`
	IsAdmin bool  `json:"isAdmin"`
}

type GetRoomMemberListResponse struct {
	RoomId         int64                    `json:"roomId"`
	RoomMemberList []*GetRoomMemberResponse `json:"roomMemberList"`
}
