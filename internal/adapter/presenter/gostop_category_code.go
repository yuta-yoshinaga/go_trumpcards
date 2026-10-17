//go:build !js || !wasm || extra10

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// goStopCategoryCode はゴーストップの韓国語札種別を共通コードへ変換する。
func goStopCategoryCode(category domain.GoStopCategory) string {
	switch category {
	case domain.GoStopGwang:
		return "gwang"
	case domain.GoStopYeol:
		return "yeol"
	case domain.GoStopTti:
		return "tti"
	default:
		return "pi"
	}
}
