package repo_dto

type AddConnectionRequest struct {
	UserId     int64  `json:"user_id" validate:"required,gt=0"`
	ConnId     int64  `json:"conn_id" validate:"required"`
	DeviceType string `json:"device_type" validate:"required"`
}
