# Tempo Go Extension SDK

Public Go guest SDK for authoring sandboxed WASM extensions for
[Tempo](https://github.com/aidanlsb/tempo).

The SDK is independently versioned from the Tempo CLI. Compatibility is
determined by the emitted extension ABI (`tempo.extension/v1`), not by matching
CLI and SDK versions.

## Packages

- `extension`: descriptors, invocation types, Effects, Events, and validation
- `guest`: Go/WASI export implementation and request dispatch
- `rng`: deterministic, recordable guest randomness

## Universe setup

```toml
# universe.toml
[extension]
language = "go"
abi = "tempo.extension/v1"
source = "extension"
```

```go
require github.com/aidanlsb/tempo-sdk-go v0.1.0
```

A guest configures its descriptor and exports the SDK bridge functions:

```go
package main

import (
    "github.com/aidanlsb/tempo-sdk-go/extension"
    "github.com/aidanlsb/tempo-sdk-go/guest"
)

func init() {
    guest.Configure(extension.Descriptor{ID: "example"}, invoke)
}

func invoke(request extension.InvokeRequest) extension.InvokeResult {
    return extension.InvokeResult{}
}

//go:wasmexport tempo_alloc
func tempoAlloc(size uint32) uint32 { return guest.Alloc(size) }

//go:wasmexport tempo_free
func tempoFree(pointer, _ uint32) { guest.Free(pointer) }

//go:wasmexport tempo_handshake
func tempoHandshake() uint64 { return guest.Handshake() }

//go:wasmexport tempo_describe
func tempoDescribe() uint64 { return guest.Describe() }

//go:wasmexport tempo_invoke
func tempoInvoke(pointer, length uint32) uint64 { return guest.Invoke(pointer, length) }

func main() {}
```

Builds use ordinary Go module resolution. For unreleased local SDK development,
use an ignored `go.work` rather than committing a path-based `replace`.
