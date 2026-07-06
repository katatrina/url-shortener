package request

// Normalizer is implemented by request DTOs that clean up their own fields
// (trim whitespace, lowercase, etc.) before validation. ShouldBindJSON calls
// Normalize after unmarshaling and before validating, if obj implements it.
type Normalizer interface {
	Normalize()
}
