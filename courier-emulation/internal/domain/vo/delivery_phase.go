package vo

// DeliveryPhase represents the current phase in a delivery workflow.
type DeliveryPhase string

const (
	// PhaseIdle indicates the courier is idle and waiting for orders.
	PhaseIdle DeliveryPhase = "idle"

	// PhaseHeadingToPickup indicates the courier is navigating to the pickup location.
	PhaseHeadingToPickup DeliveryPhase = "heading_to_pickup"

	// PhasePickingUp indicates the courier has arrived at pickup and is collecting the package.
	PhasePickingUp DeliveryPhase = "picking_up"

	// PhaseHeadingToCustomer indicates the courier is navigating to the customer's location.
	PhaseHeadingToCustomer DeliveryPhase = "heading_to_customer"

	// PhaseDelivering indicates the courier has arrived at customer and is completing delivery.
	PhaseDelivering DeliveryPhase = "delivering"
)

// String returns the string representation of the phase.
func (p DeliveryPhase) String() string {
	return string(p)
}

// IsActive returns true if the courier is actively working on a delivery.
func (p DeliveryPhase) IsActive() bool {
	return p != PhaseIdle
}

// IsMoving returns true if the courier is in a moving phase.
func (p DeliveryPhase) IsMoving() bool {
	return p == PhaseHeadingToPickup || p == PhaseHeadingToCustomer
}

// IsWaiting returns true if the courier is waiting at a location.
func (p DeliveryPhase) IsWaiting() bool {
	return p == PhasePickingUp || p == PhaseDelivering
}

// ToCourierStatus converts the delivery phase to a courier status for location events.
func (p DeliveryPhase) ToCourierStatus() string {
	switch p {
	case PhaseHeadingToPickup, PhaseHeadingToCustomer:
		return CourierStatusMoving
	case PhasePickingUp:
		return CourierStatusPickingUp
	case PhaseDelivering:
		return CourierStatusDelivering
	default:
		return CourierStatusIdle
	}
}
