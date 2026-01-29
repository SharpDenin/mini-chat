package dto

type PresenceCommand string

const (
	CmdSubscribe        PresenceCommand = "subscribe"
	CmdUnsubscribe      PresenceCommand = "unsubscribe"
	CmdGetOnlineFriends PresenceCommand = "get_online_friends"
)

type PresencePayload struct {
	Cmd     PresenceCommand `json:"cmd"`
	UserIds []int64         `json:"user_ids"`
}
