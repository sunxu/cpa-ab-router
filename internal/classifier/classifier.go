package classifier

import (
	"encoding/json"
	"strings"
	"unicode"
)

type Result struct {
	Route          string `json:"route"`
	Classification string `json:"classification"`
	Reason         string `json:"reason"`
	Score          int    `json:"score"`
	Segments       int    `json:"segments"`
}

// The router is an A-admission gate, not a general-purpose content taxonomy.
// It is deliberately asymmetric: uncertain traffic goes to B.
var strongRiskTerms = []string{
	"做爱", "交媾", "口交", "肛交", "自慰", "手淫", "潮吹", "内射", "抽插",
	"后穴", "花穴", "肉穴", "鸡巴", "几把", "肉棒", "操逼", "草逼", "骚逼", "逼里", "磨逼", "扣喷",
	"强行标记", "结合热", "全塞进去", "沉腰进入", "射进去", "射进里面", "射在里面",
	"聊黄", "聊骚", "dirty talk",
}

// These terms can occur in medical/policy/research discussion. Outside a clear
// discussion context, any one is enough to keep the request out of A.
var contextualRiskTerms = []string{
	"性交", "性行为", "性关系", "射精", "高潮", "性欲", "性瘾", "发情", "交配", "性器官", "性器",
	"阴道", "阴茎", "龟头", "肛门", "私处", "下体", "乳房", "乳头", "奶子", "巨乳",
	"裸体", "赤裸", "没穿衣服", "湿透",
	"呻吟", "娇喘", "喘息", "哼唧", "好痒", "难耐", "欲火", "情欲", "越来越湿",
	"太大了", "塞进去", "里面越来越硬", "还没软", "还是硬的", "喷了", "喷出来", "水声",
	"脱衣服", "解开衣服", "不要停", "再深", "再快", "床上等", "抱去床上", "去床上", "调教", "双修",
}

var anatomyTerms = []string{
	"后穴", "肛门", "阴道", "花穴", "肉穴", "私处", "下体", "阴茎", "龟头", "鸡巴", "几把", "肉棒", "性器官", "性器",
	"乳房", "乳头", "奶子", "巨乳",
}

var actionTerms = []string{
	"抽插", "插入", "开拓", "扩张", "顶弄", "抽送", "撞击", "塞入", "塞进去", "摩擦", "勃起", "舔", "含", "揉", "捏", "玩弄", "射",
}

var discussionTerms = []string{
	"性教育", "性健康", "生殖健康", "医学", "医疗", "疾病", "感染", "避孕", "安全套",
	"研究", "论文", "统计", "新闻", "政策", "审核", "风控", "过滤", "分类器", "检测", "识别", "占比",
}

func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\u200b", "")
	return strings.Join(strings.FieldsFunc(s, unicode.IsSpace), " ")
}

func ClassifyJSON(body []byte) Result {
	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return Result{Route: "B", Classification: "uncertain", Reason: "invalid_json"}
	}
	segments := ExtractTextSegments(root)
	if len(segments) == 0 {
		return Result{Route: "B", Classification: "uncertain", Reason: "no_classifiable_text"}
	}

	for _, segment := range segments {
		if r, ok := classifySegment(segment); ok {
			r.Segments = len(segments)
			return r
		}
	}
	return Result{Route: "A", Classification: "ordinary", Reason: "high_confidence_ordinary", Segments: len(segments)}
}

func classifySegment(raw string) (Result, bool) {
	text := normalizeText(raw)
	if text == "" {
		return Result{}, false
	}
	lower := strings.ToLower(text)

	if containsAny(lower, strongRiskTerms) {
		return Result{Route: "B", Classification: "sexual", Reason: "strong_risk_term", Score: 10}, true
	}

	// Euphemistic combinations retained from the historical v3 gate.
	if strings.Contains(text, "里面") && containsAny(text, []string{"越来越硬", "射", "抽插", "顶弄", "疯狂捣", "捣"}) {
		return Result{Route: "B", Classification: "sexual", Reason: "euphemistic_inside_action", Score: 8}, true
	}
	if containsAny(text, []string{"想要", "快点"}) && containsAny(text, []string{"难耐", "好痒", "湿透", "湿漉漉", "大开着", "呻吟", "娇喘", "后穴", "性器", "太大"}) {
		return Result{Route: "B", Classification: "sexual", Reason: "euphemistic_desire_combo", Score: 8}, true
	}

	// Anatomy + action is risky even when the surrounding text is educational.
	// The experiment protects A purity, so quoted explicit detail is still B exposure.
	if containsAny(text, anatomyTerms) && containsAny(text, actionTerms) {
		return Result{Route: "B", Classification: "sexual", Reason: "anatomy_action_combination", Score: 8}, true
	}

	discussion := containsAny(text, discussionTerms)
	if containsAny(text, contextualRiskTerms) && !discussion {
		return Result{Route: "B", Classification: "uncertain", Reason: "contextual_risk_term", Score: 3}, true
	}
	return Result{}, false
}

func containsAny(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}
