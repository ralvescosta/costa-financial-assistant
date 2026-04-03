package errors

// PointerPolicyReason identifies approved reasons for value-semantics exceptions.
type PointerPolicyReason string

const (
	PointerPolicyReasonImmutableSmallValue PointerPolicyReason = "immutable_small_value"
	PointerPolicyReasonSafetyCopy          PointerPolicyReason = "safety_copy"
	PointerPolicyReasonCompatibilityBridge PointerPolicyReason = "compatibility_bridge"
)

// PointerPolicyException documents a boundary where value semantics are
// intentionally preserved.
type PointerPolicyException struct {
	StructName    string
	Boundary      string
	Reason        PointerPolicyReason
	Justification string
}

// IsValid reports whether the exception is sufficiently documented.
func (e PointerPolicyException) IsValid() bool {
	if e.StructName == "" || e.Boundary == "" || e.Justification == "" {
		return false
	}

	switch e.Reason {
	case PointerPolicyReasonImmutableSmallValue,
		PointerPolicyReasonSafetyCopy,
		PointerPolicyReasonCompatibilityBridge:
		return true
	default:
		return false
	}
}
