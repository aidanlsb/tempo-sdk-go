package extension

import "testing"

func TestValidateHandshake(t *testing.T) {
	valid := Handshake{
		ABI: ABIV1,
		SDK: SDK{Language: "go", Module: "example.com/sdk", Version: "v1.2.3"},
	}
	if err := ValidateHandshake(valid); err != nil {
		t.Fatalf("ValidateHandshake() error = %v", err)
	}
	for name, mutate := range map[string]func(*Handshake){
		"abi":      func(value *Handshake) { value.ABI = "" },
		"language": func(value *Handshake) { value.SDK.Language = "" },
		"module":   func(value *Handshake) { value.SDK.Module = "" },
		"version":  func(value *Handshake) { value.SDK.Version = "" },
	} {
		t.Run(name, func(t *testing.T) {
			value := valid
			mutate(&value)
			if err := ValidateHandshake(value); err == nil {
				t.Fatal("ValidateHandshake() accepted incomplete provenance")
			}
		})
	}
}
