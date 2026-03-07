package validator_test

import (
	"testing"

	"stockanalyzr/pkg/validator"
)

type TestStruct struct {
	Name string `validate:"required"`
	Age  int    `validate:"gte=18"`
}

func TestValidator_New(t *testing.T) {
	v := validator.New()
	if v == nil {
		t.Fatal("Expected validator instance, got nil")
	}
	if v.Validate == nil {
		t.Fatal("Expected underlying validate instance, got nil")
	}
}

func TestValidator_Struct(t *testing.T) {
	v := validator.New()

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name: "John Doe",
				Age:  25,
			},
			wantErr: false,
		},
		{
			name: "missing required name",
			input: TestStruct{
				Age: 25,
			},
			wantErr: true,
		},
		{
			name: "age below minimum",
			input: TestStruct{
				Name: "Jane Doe",
				Age:  17,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validator.Struct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
