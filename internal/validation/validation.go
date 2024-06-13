package validation

import (
	"errors"
	"fmt"

	"github.com/PhilippNikitin/WBTECH-L0-task/tree/main/internal/models"
)

// Статически определяем ошибки, которая будут использоваться при валидации данных о заказе
var ErrOrderDataIncorrect = errors.New("ошибка в полученных данных")
var ErrOrderDataAbsent = errors.New("отсутствуют необходимые данные")

func ValidateOrder(order models.Order) error {
	switch {
	// Валидация OrderUID
	case len(order.OrderUID) != 19 && len(order.OrderUID) != 0:
		return fmt.Errorf("orderuid: %w", ErrOrderDataIncorrect)
	case len(order.OrderUID) == 0:
		return fmt.Errorf("orderuid: %w", ErrOrderDataAbsent)

	// Валидация Track_number
	case len(order.Track_number) != 14 && len(order.OrderUID) != 0:
		return fmt.Errorf("track_number: %w", ErrOrderDataIncorrect)
	case len(order.Track_number) == 0:
		return fmt.Errorf("track_number: %w", ErrOrderDataAbsent)

	// Валидация Customer_id
	case len(order.Customer_id) != 4 && len(order.OrderUID) != 0:
		return fmt.Errorf("customer_id: %w", ErrOrderDataIncorrect)
	case len(order.Customer_id) == 0:
		return fmt.Errorf("customer_id: %w", ErrOrderDataAbsent)

	// Валидация Delivery_service
	case len(order.Delivery_service) != 5 && len(order.OrderUID) != 0:
		return fmt.Errorf("delivery_service: %w", ErrOrderDataIncorrect)
	case len(order.Track_number) == 0:
		return fmt.Errorf("delivery_service: %w", ErrOrderDataAbsent)

	// Валидация имени покупателя
	case len(order.Delivery.Name) == 0:
		return fmt.Errorf("delivery.name: %w", ErrOrderDataAbsent)

	// Валидация транзакции
	case len(order.Payment.Transaction) == 0:
		return fmt.Errorf("payment.transaction: %w", ErrOrderDataAbsent)

	case order.Items.Price == 0.:
		return fmt.Errorf("items.price: %w", ErrOrderDataIncorrect)

	default:
		return nil
	}
}
