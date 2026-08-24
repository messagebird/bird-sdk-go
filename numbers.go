package bird

// NumbersService is the numbers a workspace holds, plus the Available search
// and the Orders that turn one into the other. Reach it via Client.Numbers.
//
// Buying is an order rather than a direct create: most complete inside the
// request, but one that has to wait on a carrier comes back pending and is
// polled through Client.Numbers.Orders.Get.
type NumbersService struct {
	resource

	// Available searches the numbers on sale in a country.
	Available *NumbersAvailableService
	// Orders buys a number and reports where a purchase stands.
	Orders *NumbersOrdersService
}

// NumbersAvailableService searches numbers on sale. Reach it via
// Client.Numbers.Available.
type NumbersAvailableService struct{ resource }

// NumbersOrdersService buys numbers and reads purchases. Reach it via
// Client.Numbers.Orders.
type NumbersOrdersService struct{ resource }
