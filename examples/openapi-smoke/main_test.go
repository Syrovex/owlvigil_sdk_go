package main

import (
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

func TestSmokeEnvironment(t *testing.T) {
	tests := []struct {
		value string
		want  owlvigil.Environment
	}{
		{value: "", want: owlvigil.EnvironmentProduction},
		{value: "production", want: owlvigil.EnvironmentProduction},
		{value: "staging", want: owlvigil.EnvironmentStaging},
		{value: "local", want: owlvigil.EnvironmentLocal},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := smokeEnvironment(tt.value); got != tt.want {
				t.Errorf("smokeEnvironment(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestSmokeEnabled(t *testing.T) {
	if !smokeEnabled("1") {
		t.Error(`smokeEnabled("1") = false, want true`)
	}
	if smokeEnabled("true") {
		t.Error(`smokeEnabled("true") = true, want false`)
	}
}
