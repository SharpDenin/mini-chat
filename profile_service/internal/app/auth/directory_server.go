package auth

import (
	"context"
	"os"
	"profile_service/internal/app/user/service"
	"proto/generated/profile"
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DirectoryServer struct {
	log *logrus.Logger
	profile.UnimplementedUserDirectoryServer
	uService service.UserServiceInterface
}

func NewDirectoryServer(log *logrus.Logger, uService service.UserServiceInterface) *DirectoryServer {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &DirectoryServer{
		log:      log,
		uService: uService,
	}
}

func (s *DirectoryServer) UserExists(ctx context.Context, req *profile.UserExistsRequest) (*profile.UserExistsResponse, error) {
	s.log.WithField("user_id", req.UserId).Debug("UserExists request")
	userId, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		s.log.WithError(err).WithField("user_id", req.UserId).Error("Invalid user ID format")
		return &profile.UserExistsResponse{
			Exists: false,
		}, status.Error(codes.InvalidArgument, "Invalid user ID format")
	}

	user, err := s.uService.GetUserById(ctx, userId)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			s.log.WithField("user_id", req.UserId).Debug("User not found")
			return &profile.UserExistsResponse{
				Exists: false,
			}, nil
		}

		s.log.WithError(err).Error("Failed to check user existence")
		return &profile.UserExistsResponse{
			Exists: false,
		}, status.Error(codes.Internal, "Failed to check user existence")
	}

	s.log.WithField("user_id", req.UserId).Debug("User exists")
	return &profile.UserExistsResponse{
		Exists: user != nil,
	}, nil
}

func (s *DirectoryServer) UsersExist(ctx context.Context, req *profile.UsersExistRequest) (*profile.UsersExistResponse, error) {
	s.log.WithField("user_ids", req.UserIds).Debug("UsersExist request")

	result := make(map[string]bool)

	for _, userId := range req.UserIds {
		existsResp, err := s.UserExists(ctx, &profile.UserExistsRequest{UserId: userId})
		if err != nil {
			s.log.WithError(err).WithField("user_id", userId).Warn("Failed to check user existence")
			result[userId] = false
			continue
		}
		result[userId] = existsResp.Exists
	}

	return &profile.UsersExistResponse{
		Exists: result,
	}, nil
}
