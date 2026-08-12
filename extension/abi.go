package extension

import "fmt"

// ABIV1 identifies the first stable Tempo extension wire contract. ABI
// versions are independent of Tempo CLI and guest SDK versions.
const ABIV1 = "tempo.extension/v1"

// SupportedABIs is the extension compatibility window implemented by this
// package. Return a copy so callers cannot mutate process-wide capabilities.
func SupportedABIs() []string {
	return []string{ABIV1}
}

// SDK identifies the guest adapter that emitted a WASM module. It is build
// provenance only; ABI is the compatibility authority.
type SDK struct {
	Language string `json:"language"`
	Module   string `json:"module"`
	Version  string `json:"version"`
}

// Handshake is emitted by a guest before its descriptor is accepted.
type Handshake struct {
	ABI string `json:"abi"`
	SDK SDK    `json:"sdk"`
}

// ValidateHandshake checks the stable, language-neutral handshake shape.
func ValidateHandshake(value Handshake) error {
	if value.ABI == "" {
		return fmt.Errorf("ABI is required")
	}
	if value.SDK.Language == "" {
		return fmt.Errorf("SDK language is required")
	}
	if value.SDK.Module == "" {
		return fmt.Errorf("SDK module is required")
	}
	if value.SDK.Version == "" {
		return fmt.Errorf("SDK version is required")
	}
	return nil
}
