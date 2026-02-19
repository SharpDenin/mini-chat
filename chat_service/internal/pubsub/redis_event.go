package pubsub

import "encoding/json"

type RedisEvent struct {
	Type       string          `json:"type"`
	InstanceId string          `json:"instance_id"`
	Data       json.RawMessage `json:"data"`
}
