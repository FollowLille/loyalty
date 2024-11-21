package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckLunar(t *testing.T) {
	type args struct {
		orderNumber string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid_order_number",
			args: args{
				orderNumber: "79927398713", // Пример валидного номера по алгоритму Луна
			},
			want: true,
		},
		{
			name: "invalid_order_number",
			args: args{
				orderNumber: "79927398710", // Пример невалидного номера
			},
			want: false,
		},
		{
			name: "empty_string",
			args: args{
				orderNumber: "", // Пустая строка
			},
			want: false,
		},
		{
			name: "non_numeric_characters",
			args: args{
				orderNumber: "7992abc8713", // Строка с нечисловыми символами
			},
			want: false,
		},
		{
			name: "single_digit_valid",
			args: args{
				orderNumber: "0", // Валидный случай для одного символа
			},
			want: true,
		},
		{
			name: "single_digit_invalid",
			args: args{
				orderNumber: "1", // Невалидный случай для одного символа
			},
			want: false,
		},
		{
			name: "leading_zeros_invalid",
			args: args{
				orderNumber: "00000012345671",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckLunar(tt.args.orderNumber)
			assert.Equal(t, tt.want, got, "CheckLunar() failed for test case: %v", tt.name)
		})
	}
}
