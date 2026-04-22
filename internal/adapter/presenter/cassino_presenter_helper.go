package presenter

// cassinoPhaseLabel translates a Cassino phase enum into a Japanese label.
func cassinoPhaseLabel(phase string) string {
	switch phase {
	case "dealing":
		return "配札中"
	case "playerTurn":
		return "プレイ中"
	case "roundEnd":
		return "ラウンド終了"
	case "gameEnd":
		return "ゲーム終了"
	default:
		return phase
	}
}
