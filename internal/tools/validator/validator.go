package validator

import "github.com/go-playground/validator/v10"

// Validate is a single instance of Validate, it caches struct info.
// Use this globally across the application.
var Validate *validator.Validate

func init() {
	// Initialize the validator with WithRequiredStructEnabled.
	// This option enables the new behavior that will become default in v11+,
	// ensuring that any missing `validate:"required"` fields in inner structs
	// correctly fail validation, rather than skipping the inner struct validation.
	Validate = validator.New(validator.WithRequiredStructEnabled())
}
