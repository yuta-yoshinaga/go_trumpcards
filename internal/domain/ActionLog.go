package domain

// ActionLogEntry 棋譜エントリ
type ActionLogEntry struct {
	TurnNumber int     // ターン番号
	PlayerIdx  int     // プレイヤーインデックス (-1 = システム/ディーラー)
	ActionType string  // アクション種別 (ゲーム固有)
	Detail     string  // 人間が読める説明
	Cards      []*Card // 関連カード (常に公開)
}
