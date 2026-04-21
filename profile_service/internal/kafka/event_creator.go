package kafka

import (
	"profile_service/internal/kafka/models"
	"time"
)

func NewFriendRequestSentEvent(senderId, receiverId, requestId int64, message string) *models.FriendRequestSentEvent {
	return &models.FriendRequestSentEvent{
		BaseEvent: models.BaseEvent{
			EventId:   generateEventID(),
			EventType: models.EventFriendRequestSent,
			UserId:    senderId,
			Timestamp: time.Now().UTC(),
			Service:   "profile-service",
			Version:   "1.0",
		},
		SenderId:   senderId,
		ReceiverId: receiverId,
		RequestId:  requestId,
		Message:    message,
	}
}

func NewFriendRequestActionEvent(userId, requestId, friendId int64, action string) *models.FriendRequestActionEvent {
	var eventType models.EventType
	switch action {
	case "accepted":
		eventType = models.EventFriendRequestAccepted
	case "rejected":
		eventType = models.EventFriendRequestRejected
	case "cancelled":
		eventType = models.EventFriendRequestCancelled
	default:
		eventType = models.EventFriendRequestAccepted
	}
	return &models.FriendRequestActionEvent{
		BaseEvent: models.BaseEvent{
			EventId:   generateEventID(),
			EventType: eventType,
			UserId:    userId,
			Timestamp: time.Now().UTC(),
			Service:   "profile-service",
			Version:   "1.0",
		},
		RequestId: requestId,
		FriendId:  friendId,
		Action:    action,
	}
}

func NewBlockEvent(blockerId, blockedId int64, action, reason string) *models.BlockEvent {
	eventType := models.EventUserBlocked
	if action == "unblock" {
		eventType = models.EventUserUnblocked
	}

	return &models.BlockEvent{
		BaseEvent: models.BaseEvent{
			EventId:   generateEventID(),
			EventType: eventType,
			UserId:    blockerId,
			Timestamp: time.Now().UTC(),
			Service:   "profile-service",
			Version:   "1.0",
		},
		BlockedId: blockedId,
		Action:    action,
		Reason:    reason,
	}
}

func NewFriendEvent(userId, friendId int64, action string) *models.FriendEvent {
	eventType := models.EventFriendAdded
	if action == "remove" {
		eventType = models.EventFriendRemoved
	}

	return &models.FriendEvent{
		BaseEvent: models.BaseEvent{
			EventId:   generateEventID(),
			EventType: eventType,
			UserId:    userId,
			Timestamp: time.Now().UTC(),
			Service:   "profile-service",
			Version:   "1.0",
		},
		FriendId: friendId,
		Action:   action,
	}
}

func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
