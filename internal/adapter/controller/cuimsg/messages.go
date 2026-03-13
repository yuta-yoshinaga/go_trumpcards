// Package cuimsg provides common error message helpers for CUI controllers.
package cuimsg

import "fmt"

// Required returns a "<field> is required." error message.
func Required(field string) string {
	return field + " is required."
}

// InvalidNotANumber returns an "Invalid <field>. Please enter a number." error message.
func InvalidNotANumber(field string) string {
	return "Invalid " + field + ". Please enter a number."
}

// InvalidOutOfRange returns an "Invalid <field>: <val>. <hint>" error message.
func InvalidOutOfRange(field, val, hint string) string {
	return fmt.Sprintf("Invalid %s: %s. %s", field, val, hint)
}
