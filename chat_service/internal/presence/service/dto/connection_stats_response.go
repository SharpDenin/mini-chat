package dto

const (
	DeviceWeb     = "web"
	DeviceIOS     = "ios"
	DeviceAndroid = "android"
	DeviceDesktop = "desktop"
)

type ConnectionStatsResponse struct {
	TotalConnections int
	Devices          map[string]int64
}
