package v1

import "time"

// DeliveryPeriod represents the desired delivery time window for an order.
type DeliveryPeriod struct {
	// startTime is the start of desired delivery period
	startTime time.Time
	// endTime is the end of desired delivery period
	endTime time.Time
}

// NewDeliveryPeriod creates a new DeliveryPeriod value object.
func NewDeliveryPeriod(startTime, endTime time.Time) DeliveryPeriod {
	return DeliveryPeriod{
		startTime: startTime,
		endTime:   endTime,
	}
}

// GetStartTime returns the start time of the delivery period.
func (d DeliveryPeriod) GetStartTime() time.Time {
	return d.startTime
}

// GetEndTime returns the end time of the delivery period.
func (d DeliveryPeriod) GetEndTime() time.Time {
	return d.endTime
}

// IsValid checks if the delivery period is valid (start < end and both are in the future).
func (d DeliveryPeriod) IsValid() bool {
	now := time.Now()
	return d.startTime.Before(d.endTime) && d.startTime.After(now)
}

// Duration returns the duration of the delivery period.
func (d DeliveryPeriod) Duration() time.Duration {
	return d.endTime.Sub(d.startTime)
}
