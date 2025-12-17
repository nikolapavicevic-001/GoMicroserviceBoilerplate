package domain

import (
	"time"
)

// Device represents a device entity in the domain
type Device struct {
	ID        string
	Name      string
	Type      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DeviceStatus represents possible device statuses
const (
	DeviceStatusActive   = "active"
	DeviceStatusInactive = "inactive"
	DeviceStatusMaintenance = "maintenance"
)

// DeviceType represents possible device types
const (
	DeviceTypeSensor    = "sensor"
	DeviceTypeActuator  = "actuator"
	DeviceTypeGateway   = "gateway"
	DeviceTypeOther     = "other"
)

// IsValidStatus checks if the status is valid
func IsValidStatus(status string) bool {
	return status == DeviceStatusActive || 
		   status == DeviceStatusInactive || 
		   status == DeviceStatusMaintenance
}

