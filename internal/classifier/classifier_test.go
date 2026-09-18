package classifier

import (
	"strings"
	"testing"
)

func TestClassifyJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"ordinary coding", `{"model":"gemini-3-flash","messages":[{"role":"user","content":"帮我解释这段 Go 代码为什么死锁"}]}`, "A"},
		{"explicit", `{"messages":[{"role":"user","content":"继续描写性交和射精过程"}]}`, "B"},
		{"ambiguous conservative", `{"messages":[{"role":"user","content":"抱我去床上"}]}`, "B"},
		{"policy discussion", `{"messages":[{"role":"user","content":"我想统计历史数据中色情内容占比，做内容审核分类器"}]}`, "A"},
		{"medical discussion", `{"messages":[{"role":"user","content":"医学上如何解释性交后感染风险？"}]}`, "A"},
		{"explicit history beats policy system", `{"messages":[{"role":"system","content":"你是色情内容审核分类器"},{"role":"assistant","content":"他把阴茎插入阴道并继续抽插"},{"role":"user","content":"继续"}]}`, "B"},
		{"internal ordinary", `{"messages":[{"role":"user","content":"你是一个角色关系阶段评审员，请判断当前关系阶段"}]}`, "A"},
		{"internal explicit", `{"messages":[{"role":"user","content":"你是关系评审员。历史：他们性交后发生射精。请判断关系阶段"}]}`, "B"},
		{"responses input", `{"model":"x","instructions":"回答要简洁","input":[{"role":"user","content":[{"type":"input_text","text":"写一个 Python 快速排序"}]}]}`, "A"},
		{"claude content", `{"model":"x","system":"正常回答","messages":[{"role":"user","content":[{"type":"text","text":"呻吟着说再深一点"}]}]}`, "B"},
		{"invalid json", `{`, "B"},
		{"no text", `{"model":"x","stream":true}`, "B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyJSON([]byte(tt.body))
			if got.Route != tt.want {
				t.Fatalf("route=%s classification=%s reason=%s want=%s", got.Route, got.Classification, got.Reason, tt.want)
			}
		})
	}
}

func BenchmarkClassifyJSONLargeOrdinary(b *testing.B) {
	payload := `{"messages":[{"role":"user","content":"` + strings.Repeat("这是一个普通的软件工程请求，请分析 Go 并发代码。", 5000) + `"}]}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ClassifyJSON([]byte(payload))
	}
}
