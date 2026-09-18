package classifier

import "strings"

var ignoredStringKeys = map[string]struct{}{
	"model": {}, "role": {}, "type": {}, "name": {}, "id": {}, "object": {},
	"status": {}, "finish_reason": {}, "finishreason": {}, "call_id": {}, "tool_call_id": {},
	"mime_type": {}, "mimetype": {}, "url": {}, "uri": {},
}

// ExtractTextSegments conservatively collects textual values that can be part of a
// model request. It intentionally does not care whether a request is USER_CHAT or
// INTERNAL. History, system/instructions, tool descriptions and nested content are
// all inspected independently so a benign policy string cannot mask explicit history.
func ExtractTextSegments(v any) []string {
	out := make([]string, 0, 16)
	walkText(v, "", &out)
	return out
}

func walkText(v any, key string, out *[]string) {
	switch x := v.(type) {
	case string:
		if _, ignored := ignoredStringKeys[strings.ToLower(key)]; ignored {
			return
		}
		if strings.TrimSpace(x) != "" {
			*out = append(*out, x)
		}
	case []any:
		for _, item := range x {
			walkText(item, key, out)
		}
	case map[string]any:
		for k, item := range x {
			walkText(item, k, out)
		}
	}
}
