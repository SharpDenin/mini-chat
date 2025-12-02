package dto

import "time"

type MarkOptionRequest func(*markOptions)

type markOptions struct {
	Source    string
	DeviceID  string
	Force     bool
	Timestamp time.Time
}

func WithSource(source string) MarkOptionRequest {
	return func(o *markOptions) { o.Source = source }
}

func WithDeviceID(deviceID string) MarkOptionRequest {
	return func(o *markOptions) { o.DeviceID = deviceID }
}

func WithForce() MarkOptionRequest {
	return func(o *markOptions) { o.Force = true }
}
