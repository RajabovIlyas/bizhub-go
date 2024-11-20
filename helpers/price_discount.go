package helpers

import (
	"fmt"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/models"
)

func ValidateDiscountData(data models.DiscountData) error {
	errStr := ""
	if data.Duration < 1 {
		// fmt.Printf("\nDuration: %v\n", data.Duration)
		errStr = errStr + " [duration not valid]"
	}
	if !SliceContains(config.DISCOUNT_TYPES, data.Type) {
		// fmt.Printf("\nDiscount type: %v\n", data.Type)
		errStr = errStr + " [discount type not valid]"
	}
	if !SliceContains(config.DURATION_TYPES, data.DurationType) {
		// fmt.Printf("\nDuration type: %v\n", data.DurationType)
		errStr = errStr + " [duration type not valid]"
	}
	if data.Percent <= 0 || data.Percent > 100 {
		// fmt.Printf("\nPercent: %v\n", data.Percent)
		errStr = errStr + " [discount percentage not valid]"
	}
	if data.Price <= 0 {
		// fmt.Printf("\nPrice: %v\n", data.Price)
		errStr = errStr + " [discount price not valid]"
	}
	if len(errStr) > 0 {
		return fmt.Errorf("Error: %v", errStr)
	}
	return nil
}
