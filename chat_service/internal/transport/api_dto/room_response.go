package api_dto

type GetRoomResponse struct {
	Id   int64  `json:"Id"`
	Name string `json:"Name"`
}

type GetRoomList struct {
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
	RoomList []*GetRoomResponse `json:"roomList"`
}
