package export

import (
	"strings"
)

const (
	FormatterNamePlainText = "plain-text"
	FormatterNamePlainJson = "plain-json"
	fourSpaces             = "    "
)

var (
	AllFormattersNames = strings.Join([]string{FormatterNamePlainText, FormatterNamePlainText}, ", ")
)

type formatterArgs struct {
}
