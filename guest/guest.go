// Package guest provides the JSON ABI implementation used by Go-authored WASM
// extensions. A package main exports thin go:wasmexport wrappers around these
// functions.
package guest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/aidanlsb/tempo-sdk-go/extension"
)

type Handler func(extension.InvokeRequest) extension.InvokeResult

var state struct {
	sync.Mutex
	handshake  extension.Handshake
	descriptor extension.Descriptor
	handler    Handler
	allocated  map[uint32][]byte
}

// Configure retains the development bridge for guests importing Tempo's
// original extension package. New guest SDKs should call ConfigureWithHandshake
// with their own module and version provenance.
func Configure(descriptor extension.Descriptor, handler Handler) {
	ConfigureWithHandshake(extension.Handshake{
		ABI: extension.ABIV1,
		SDK: extension.SDK{
			Language: "go",
			Module:   "github.com/aidanlsb/tempo-sdk-go",
			Version:  "v0.1.0",
		},
	}, descriptor, handler)
}

func ConfigureWithHandshake(
	handshake extension.Handshake,
	descriptor extension.Descriptor,
	handler Handler,
) {
	state.Lock()
	defer state.Unlock()
	state.handshake = handshake
	state.descriptor = descriptor
	state.handler = handler
	if state.allocated == nil {
		state.allocated = make(map[uint32][]byte)
	}
}

func Alloc(size uint32) uint32 {
	state.Lock()
	defer state.Unlock()
	if size == 0 {
		return 0
	}
	data := make([]byte, size)
	pointer := uint32(uintptr(unsafe.Pointer(&data[0])))
	state.allocated[pointer] = data
	return pointer
}

func Free(pointer uint32) {
	state.Lock()
	defer state.Unlock()
	delete(state.allocated, pointer)
}

func Handshake() uint64 {
	state.Lock()
	defer state.Unlock()
	data, err := json.Marshal(state.handshake)
	if err != nil {
		return outputLocked(marshalError(err))
	}
	return outputLocked(data)
}

func Describe() uint64 {
	state.Lock()
	defer state.Unlock()
	data, err := json.Marshal(state.descriptor)
	if err != nil {
		return outputLocked(marshalError(err))
	}
	return outputLocked(data)
}

func Invoke(pointer, length uint32) uint64 {
	state.Lock()
	defer state.Unlock()
	input := state.allocated[pointer]
	if uint32(len(input)) < length {
		return outputLocked(marshalErrorString("invocation input is out of bounds"))
	}
	var request extension.InvokeRequest
	if err := decodeJSON(input[:length], &request); err != nil {
		return outputLocked(marshalError(err))
	}
	if state.handler == nil {
		return outputLocked(marshalErrorString("guest handler is not configured"))
	}
	result := state.handler(request)
	data, err := json.Marshal(result)
	if err != nil {
		return outputLocked(marshalError(err))
	}
	return outputLocked(data)
}

func outputLocked(data []byte) uint64 {
	if len(data) == 0 {
		return 0
	}
	pointer := uint32(uintptr(unsafe.Pointer(&data[0])))
	state.allocated[pointer] = data
	return uint64(pointer)<<32 | uint64(len(data))
}

func marshalError(err error) []byte {
	return marshalErrorString(err.Error())
}

func marshalErrorString(message string) []byte {
	data, _ := json.Marshal(extension.InvokeResult{Rejected: message})
	return data
}

func decodeJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing JSON content")
	}
	return nil
}
