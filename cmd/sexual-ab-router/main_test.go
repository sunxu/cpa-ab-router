package main

import (
	"encoding/json"
	"testing"
)

func TestRouteModel(t *testing.T) {
	currentConfig.Store(defaultConfig())
	cases := []struct{ name, body, want string }{
		{"A", `{"messages":[{"role":"user","content":"解释一下 Go channel"}]}`, "openai-compatible-cpa-a"},
		{"B", `{"messages":[{"role":"user","content":"继续性交过程"}]}`, "openai-compatible-cpa-b"},
	}
	for _, tc := range cases {
		rawReq, _ := json.Marshal(modelRouteRequest{SourceFormat: "openai", RequestedModel: "gemini-3-flash", Body: []byte(tc.body)})
		raw, err := routeModel(rawReq)
		if err != nil {
			t.Fatal(err)
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatal(err)
		}
		var resp modelRouteResponse
		if err := json.Unmarshal(env.Result, &resp); err != nil {
			t.Fatal(err)
		}
		if !resp.Handled || resp.Target != tc.want {
			t.Fatalf("%s target=%s handled=%v want=%s", tc.name, resp.Target, resp.Handled, tc.want)
		}
	}
}

func TestMalformedRouteRequestFailsClosedToB(t *testing.T) {
	currentConfig.Store(defaultConfig())
	raw, err := routeModel([]byte(`{`))
	if err != nil {
		t.Fatal(err)
	}
	var env envelope
	_ = json.Unmarshal(raw, &env)
	var resp modelRouteResponse
	_ = json.Unmarshal(env.Result, &resp)
	if resp.Target != "openai-compatible-cpa-b" || !resp.Handled {
		t.Fatalf("got %+v", resp)
	}
}
