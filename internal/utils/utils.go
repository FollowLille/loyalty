// Package utils предоставляет функции для работы с программой лояльности
package utils

import "strconv"

// CheckLunar проверяет, является ли номер заказа корректным.
// Если номер заказа является корректным, возвращает true. В противном случае возвращает false.
//
// Параметры:
//   - orderNumber: номер заказа.
//
// Возвращает:
//   - bool: true, если номер заказа является корректным; false в противном случае.
func CheckLunar(orderNumber string) bool {
	var sum int
	var remainder = len(orderNumber) % 2
	if len(orderNumber) == 0 {
		return false
	}
	for i, num := range orderNumber {
		digit, err := strconv.Atoi(string(num))

		if err != nil {
			return false
		}

		if i%2 == remainder {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}
