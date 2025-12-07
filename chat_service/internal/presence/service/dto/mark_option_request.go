package dto

import "time"

type MarkOptionRequest func(*MarkOptions)

type MarkOptions struct {
	Source    string
	DeviceId  int64
	Force     bool
	Timestamp time.Time
}

func WithSource(source string) MarkOptionRequest {
	return func(o *MarkOptions) { o.Source = source }
}

func WithDeviceID(deviceId int64) MarkOptionRequest {
	return func(o *MarkOptions) { o.DeviceId = deviceId }
}

func WithForce() MarkOptionRequest {
	return func(o *MarkOptions) { o.Force = true }
}
