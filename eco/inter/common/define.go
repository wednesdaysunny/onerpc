package common


var TraceHeaders = []string{
	"x-request-id",
	"x-b3-traceid",
	"x-b3-spanid",
	"x-b3-parentspanid",
	"x-b3-sampled",
	"x-b3-flags",
	"x-ot-span-context",
}

const (
	Success        = "success"
	Failed         = "failed"
	CommaSeparator = ","
	DefaultLimit   = 20
	MaxLimit       = 200
	OrderMaxLimit  = 10000
)

const (
	LanguageEn = "en"
	LanguageZh = "zh"
)