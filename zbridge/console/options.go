package console

// MaxTextLen is the maximum length of the message text of a single-line WTO, in
// bytes.
//
// The value is 124, not 126. That correction came from reading GC28-0683-2
// directly rather than from a summary; the audit is in
// docs/evidence/DOC-001-wto-wpl-primary-source-2026-07-25.md. Whether z/OS
// carries the same limit for a single-line WTO is question Q3 of research brief
// 003 and is NOT settled — this constant is the ancestor's documented value and
// is used as a conservative bound until z/OS's is cited.
const MaxTextLen = 124

// RouteCode selects which console groups receive a message. Route codes are
// 1-based, matching the MVS documentation.
type RouteCode uint8

// Route codes named by the operations they select. This is not the full set;
// the ones here are those a bridge library realistically needs.
const (
	RouteMasterConsole      RouteCode = 1
	RouteMasterConsoleInfo  RouteCode = 2
	RouteTapePool           RouteCode = 3
	RouteDirectAccessPool   RouteCode = 4
	RouteTapeLibrary        RouteCode = 5
	RouteDiskLibrary        RouteCode = 6
	RouteUnitRecordPool     RouteCode = 7
	RouteTeleprocessing     RouteCode = 8
	RouteSystemSecurity     RouteCode = 9
	RouteSystemError        RouteCode = 10
	RouteProgrammerInfo     RouteCode = 11
	RouteEmulation          RouteCode = 12
)

// DescriptorCode describes the nature of a message: whether it is an immediate
// action request, an eventual action request, or informational. The descriptor
// code determines how the console presents and retains the message.
type DescriptorCode uint8

const (
	DescSystemFailure     DescriptorCode = 1
	DescImmediateAction   DescriptorCode = 2
	DescEventualAction    DescriptorCode = 3
	DescSystemStatus      DescriptorCode = 4
	DescImmediateCommand  DescriptorCode = 5
	DescJobStatus         DescriptorCode = 6
	DescApplicationStatus DescriptorCode = 7
	DescOutOfLineMessage  DescriptorCode = 8
)

// options is the resolved configuration for one WTO call.
type options struct {
	routes      []RouteCode
	descriptors []DescriptorCode
}

// Option configures a WTO call.
type Option func(*options)

// WithRoute adds route codes, selecting which console groups receive the
// message. With no route codes the system applies its defaults.
func WithRoute(codes ...RouteCode) Option {
	return func(o *options) { o.routes = append(o.routes, codes...) }
}

// WithDescriptor adds descriptor codes, describing the nature of the message.
func WithDescriptor(codes ...DescriptorCode) Option {
	return func(o *options) { o.descriptors = append(o.descriptors, codes...) }
}

func resolve(opts []Option) *options {
	o := &options{}
	for _, fn := range opts {
		fn(o)
	}
	return o
}

// routeMask returns the 16-bit route-code mask, where route code N sets the bit
// for that code. Route codes are 1-based; code 1 is the most significant bit.
//
// This function is pure arithmetic and is tested on every platform. It does NOT
// depend on the unverified WPL layout — where the mask SITS in the parameter
// list is unknown, but what the mask CONTAINS is documented and stable.
func (o *options) routeMask() uint16 {
	var mask uint16
	for _, rc := range o.routes {
		if rc >= 1 && rc <= 16 {
			mask |= 1 << (16 - uint(rc))
		}
	}
	return mask
}

// descriptorMask returns the 16-bit descriptor-code mask, on the same 1-based,
// most-significant-bit-first convention as routeMask.
func (o *options) descriptorMask() uint16 {
	var mask uint16
	for _, dc := range o.descriptors {
		if dc >= 1 && dc <= 16 {
			mask |= 1 << (16 - uint(dc))
		}
	}
	return mask
}
