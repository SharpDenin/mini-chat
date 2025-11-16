package room_models

import "time"

type Message struct {
	Id     string    `bson:"_id,omitempty"`
	UserId int64     `bson:"user_id"`
	RoomId int64     `bson:"room_id"`
	Text   string    `bson:"text"`
	SentAt time.Time `bson:"sent_at"`
}
