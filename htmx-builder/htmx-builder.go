package ghtmx

import (
	"fmt"
	"strings"
)

type HtmxElement struct {
	attributes map[string]string
	styles     map[string]string
	classes    string
}

type HtmxBuilder struct {
	content strings.Builder
}

func Div(name string, children ...HtmxElement) {
}

func NewHtmxBuilder() *HtmxBuilder {
	return &HtmxBuilder{}
}

func (hb *HtmxBuilder) Build() string {
	return hb.content.String()
}

func test() {
	builder := NewHtmxBuilder()

	fmt.Println(builder.Build())
}
