package http

import (
	"github.com/sirupsen/logrus"
	"profile_service/internal/relation/service"
	"profile_service/internal/relation/service/interfaces"
	userService "profile_service/internal/user/service"
)

type FriendshipHandler struct {
	userService       userService.UserServiceInterface
	friendshipService interfaces.FriendshipServiceInterface
	relationChecker   interfaces.UserRelationCheckerInterface
	log               *logrus.Entry
}

func NewFriendshipHandler(userService userService.UserServiceInterface, friendshipService service.FriendshipService,
	relationChecker interfaces.UserRelationCheckerInterface, log *logrus.Entry) *FriendshipHandler {
	return &FriendshipHandler{
		userService:       userService,
		friendshipService: friendshipService,
		relationChecker:   relationChecker,
		log:               log,
	}
}

// Friend Request

func (h *FriendshipHandler) PostFriendRequest() {

}

func (h *FriendshipHandler) CancelFriendRequest() {

}

func (h *FriendshipHandler) GetPendingRequest() {

}

func (h *FriendshipHandler) GetActiveRequest() {

}

func (h *FriendshipHandler) UpdateRequestStatus() {

}

// Friend

func (h *FriendshipHandler) AddFriend() {

}

func (h *FriendshipHandler) GetFriendList() {

}

func (h *FriendshipHandler) DeleteFriend() {

}

// Block

func (h *FriendshipHandler) BlockUser() {

}

func (h *FriendshipHandler) GetBlockInfo() {

}

func (h *FriendshipHandler) UnblockUser() {

}

//

func (h *FriendshipHandler) GetRelationHistory() {

}
