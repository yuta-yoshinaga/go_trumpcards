//go:build !js || !wasm || extra11

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// mushiCategoryCode は虫の札種別を共通コードへ変換する。
func mushiCategoryCode(category domain.MushiCategory) string {
	switch category {
	case domain.MushiBright:
		return "bright"
	case domain.MushiAnimal:
		return "animal"
	case domain.MushiRibbon:
		return "ribbon"
	default:
		return "chaff"
	}
}
