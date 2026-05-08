package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

// presidentRankName resolves a President player rank (1-4) to its localized
// label. Out-of-range ranks return the "unknown" label.
func presidentRankName(rank int) string {
	switch rank {
	case 1:
		return i18n.T("president.rankPresident")
	case 2:
		return i18n.T("president.rankVicePresident")
	case 3:
		return i18n.T("president.rankViceScum")
	case 4:
		return i18n.T("president.rankScum")
	default:
		return i18n.T("president.rankUnknown")
	}
}
