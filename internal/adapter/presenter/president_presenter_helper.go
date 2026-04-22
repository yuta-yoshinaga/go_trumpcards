package presenter

// presidentRankName は President のプレイヤーランク (1-4) に対応する日本語名を返します。
// 範囲外のランクには「不明」を返します。
func presidentRankName(rank int) string {
	switch rank {
	case 1:
		return "大統領"
	case 2:
		return "副大統領"
	case 3:
		return "副スカム"
	case 4:
		return "スカム"
	default:
		return "不明"
	}
}
