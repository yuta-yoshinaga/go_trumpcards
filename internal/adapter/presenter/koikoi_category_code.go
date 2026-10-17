//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// koiKoiCategoryCode はこいこいの札種別を共通コードへ変換する。
func koiKoiCategoryCode(category domain.KoiKoiCategory) string {
	switch category {
	case domain.KoiKoiBright:
		return "bright"
	case domain.KoiKoiAnimal:
		return "animal"
	case domain.KoiKoiRibbon:
		return "ribbon"
	default:
		return "chaff"
	}
}
