//go:build !js || !wasm || extra || extra3 || extra10 || extra11

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// hanafudaMonthName は月番号に対応する花札の月名を現在の言語で返す。
func hanafudaMonthName(month int) string {
	return i18n.T("hanafudaMonth" + strconv.Itoa(month))
}

// hanafudaCategoryName はカテゴリコードに対応する花札のカテゴリ名を現在の言語で返す。
func hanafudaCategoryName(code string) string {
	return i18n.T("hanafudaCategory" + strings.ToUpper(code[:1]) + code[1:])
}
