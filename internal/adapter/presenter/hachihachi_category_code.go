//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// hachiHachiCategoryCode は八八の札種別を共通コードへ変換する。
func hachiHachiCategoryCode(category domain.HachiHachiCategory) string {
	switch category {
	case domain.HachiHachiBright:
		return "bright"
	case domain.HachiHachiAnimal:
		return "animal"
	case domain.HachiHachiRibbon:
		return "ribbon"
	default:
		return "chaff"
	}
}
