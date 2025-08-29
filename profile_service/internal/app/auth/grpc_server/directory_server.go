package grpc_server

import (
	"context"
	"os"
	pb "profile_service/internal/app/auth/gRPC"
	"profile_service/internal/app/user/service"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DirectoryServer struct {
	log *logrus.Logger
	pb.UnimplementedUserDirectoryServer
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

func (s *DirectoryServer) UserExists(ctx context.Context, req *pb.UserExistsRequest) (*pb.UserExistsResponse, error) {
	s.log.WithField("user_id", req.UserId).Debug("UserExists request")

	user, err := s.uService.GetUserById(ctx, req.UserId)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			s.log.WithField("user_id", req.UserId).Debug("User not found")
			return &pb.UserExistsResponse{
				Exists: false,
			}, nil
		}

		s.log.WithError(err).Error("Failed to check user existence")
		return &pb.UserExistsResponse{
			Exists: false,
		}, status.Error(codes.Internal, "Failed to check user existence")
	}

	s.log.WithField("user_id", req.UserId).Debug("User exists")
	return &pb.UserExistsResponse{
		Exists: user != nil,
	}, nil
}

func (s *DirectoryServer) UsersExist(ctx context.Context, req *pb.UsersExistRequest) (*pb.UsersExistResponse, error) {
	s.log.WithField("user_ids", req.UserIds).Debug("UsersExist request")

	result := make(map[int64]bool)

	for _, userID := range req.UserIds {
		existsResp, err := s.UserExists(ctx, &pb.UserExistsRequest{UserId: userID})
		if err != nil {
			s.log.WithError(err).Warnf("Failed to check user %d", userID)
			result[userID] = false
			continue
		}
		result[userID] = existsResp.Exists
	}

	return &pb.UsersExistResponse{
		Exists: result,
	}, nil
}
