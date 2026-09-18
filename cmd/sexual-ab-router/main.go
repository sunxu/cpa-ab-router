package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct { void* ptr; size_t len; } cliproxy_buffer;
typedef int (*cliproxy_host_call_fn)(void*, const char*, const uint8_t*, size_t, cliproxy_buffer*);
typedef void (*cliproxy_host_free_fn)(void*, size_t);
typedef struct { uint32_t abi_version; void* host_ctx; cliproxy_host_call_fn call; cliproxy_host_free_fn free_buffer; } cliproxy_host_api;
typedef int (*cliproxy_plugin_call_fn)(char*, uint8_t*, size_t, cliproxy_buffer*);
typedef void (*cliproxy_plugin_free_fn)(void*, size_t);
typedef void (*cliproxy_plugin_shutdown_fn)(void);
typedef struct { uint32_t abi_version; cliproxy_plugin_call_fn call; cliproxy_plugin_free_fn free_buffer; cliproxy_plugin_shutdown_fn shutdown; } cliproxy_plugin_api;

extern int cliproxyPluginCall(char*, uint8_t*, size_t, cliproxy_buffer*);
extern void cliproxyPluginFree(void*, size_t);
extern void cliproxyPluginShutdown(void);
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/sunxu/cpa-ab-router/internal/classifier"
)

const (
	abiVersion        uint32 = 1
	schemaVersion     uint32 = 6
	pluginID                 = "sexual-ab-router"
	methodRegister           = "plugin.register"
	methodReconfigure        = "plugin.reconfigure"
	methodModelRoute         = "model.route"
)

type envelope struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}
type rpcError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type lifecycleRequest struct {
	ConfigYAML []byte `json:"config_yaml"`
}
type pluginConfig struct {
	Enabled   bool
	ProviderA string
	ProviderB string
}
type modelRouteRequest struct {
	SourceFormat   string
	RequestedModel string
	Stream         bool
	Body           []byte
}
type modelRouteResponse struct {
	Handled     bool
	TargetKind  string
	Target      string
	TargetModel string
	Reason      string
}
type configField struct {
	Name        string `json:"Name"`
	Type        string `json:"Type"`
	Description string `json:"Description"`
}
type metadata struct {
	Name             string        `json:"Name"`
	Version          string        `json:"Version"`
	Author           string        `json:"Author"`
	GitHubRepository string        `json:"GitHubRepository,omitempty"`
	ConfigFields     []configField `json:"ConfigFields,omitempty"`
}
type registration struct {
	SchemaVersion uint32         `json:"schema_version"`
	Metadata      metadata       `json:"metadata"`
	Capabilities  map[string]any `json:"capabilities"`
}

var currentConfig atomic.Value

func main() {}

//export cliproxy_plugin_init
func cliproxy_plugin_init(_ *C.cliproxy_host_api, plugin *C.cliproxy_plugin_api) C.int {
	if plugin == nil {
		return 1
	}
	plugin.abi_version = C.uint32_t(abiVersion)
	plugin.call = C.cliproxy_plugin_call_fn(C.cliproxyPluginCall)
	plugin.free_buffer = C.cliproxy_plugin_free_fn(C.cliproxyPluginFree)
	plugin.shutdown = C.cliproxy_plugin_shutdown_fn(C.cliproxyPluginShutdown)
	return 0
}

//export cliproxyPluginCall
func cliproxyPluginCall(method *C.char, request *C.uint8_t, requestLen C.size_t, response *C.cliproxy_buffer) C.int {
	if response != nil {
		response.ptr = nil
		response.len = 0
	}
	if method == nil {
		writeResponse(response, errorEnvelope("invalid_method", "method is required"))
		return 1
	}
	var requestBytes []byte
	if request != nil && requestLen > 0 {
		requestBytes = C.GoBytes(unsafe.Pointer(request), C.int(requestLen))
	}
	raw, err := handleMethod(C.GoString(method), requestBytes)
	if err != nil {
		writeResponse(response, errorEnvelope("plugin_error", err.Error()))
		return 1
	}
	writeResponse(response, raw)
	return 0
}

//export cliproxyPluginFree
func cliproxyPluginFree(ptr unsafe.Pointer, _ C.size_t) {
	if ptr != nil {
		C.free(ptr)
	}
}

//export cliproxyPluginShutdown
func cliproxyPluginShutdown() {}

func defaultConfig() pluginConfig {
	return pluginConfig{Enabled: true, ProviderA: "openai-compatible-cpa-a", ProviderB: "openai-compatible-cpa-b"}
}
func loadedConfig() pluginConfig {
	if v := currentConfig.Load(); v != nil {
		if c, ok := v.(pluginConfig); ok {
			return c
		}
	}
	return defaultConfig()
}

func handleMethod(method string, request []byte) ([]byte, error) {
	switch method {
	case methodRegister, methodReconfigure:
		if err := configure(request); err != nil {
			return nil, err
		}
		return okEnvelope(pluginRegistration())
	case methodModelRoute:
		return routeModel(request)
	default:
		return errorEnvelope("unknown_method", "unknown method: "+method), nil
	}
}

func configure(raw []byte) error {
	cfg := defaultConfig()
	if len(raw) > 0 {
		var req lifecycleRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return err
		}
		if len(req.ConfigYAML) > 0 {
			parseSimpleYAML(req.ConfigYAML, &cfg)
		}
	}
	if strings.TrimSpace(cfg.ProviderA) == "" || strings.TrimSpace(cfg.ProviderB) == "" {
		return fmt.Errorf("provider_a and provider_b are required")
	}
	currentConfig.Store(cfg)
	return nil
}

func parseSimpleYAML(raw []byte, cfg *pluginConfig) {
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		switch k {
		case "enabled":
			cfg.Enabled = !strings.EqualFold(v, "false")
		case "provider_a":
			cfg.ProviderA = v
		case "provider_b":
			cfg.ProviderB = v
		}
	}
}

func pluginRegistration() registration {
	return registration{SchemaVersion: schemaVersion, Metadata: metadata{Name: pluginID, Version: "0.1.0", Author: "sunxu", GitHubRepository: "https://github.com/sunxu/cpa-ab-router", ConfigFields: []configField{
		{Name: "provider_a", Type: "string", Description: "CPA-A OpenAI-compatible provider key"},
		{Name: "provider_b", Type: "string", Description: "CPA-B OpenAI-compatible provider key"},
	}}, Capabilities: map[string]any{"model_router": true}}
}

func routeModel(raw []byte) ([]byte, error) {
	cfg := loadedConfig()
	if !cfg.Enabled { // disabled is the only case where the plugin declines routing.
		return okEnvelope(modelRouteResponse{Handled: false})
	}
	var req modelRouteRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return okEnvelope(modelRouteResponse{Handled: true, TargetKind: "provider", Target: cfg.ProviderB, Reason: "router_request_decode_failed"})
	}
	result := classifier.ClassifyJSON(req.Body)
	target := cfg.ProviderB
	if result.Route == "A" {
		target = cfg.ProviderA
	}
	return okEnvelope(modelRouteResponse{Handled: true, TargetKind: "provider", Target: target, TargetModel: req.RequestedModel, Reason: result.Reason})
}

func okEnvelope(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope{OK: true, Result: raw})
}
func errorEnvelope(code, msg string) []byte {
	raw, _ := json.Marshal(envelope{OK: false, Error: &rpcError{Code: code, Message: msg}})
	return raw
}
func writeResponse(response *C.cliproxy_buffer, raw []byte) {
	if response == nil || len(raw) == 0 {
		return
	}
	ptr := C.CBytes(raw)
	if ptr == nil {
		return
	}
	response.ptr = ptr
	response.len = C.size_t(len(raw))
}
