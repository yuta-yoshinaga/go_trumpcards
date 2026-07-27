package controller

// WebOutputTrickCard トリック中の1枚の共通Web出力。
//
// 60ゲームが完全に同一の構造体を各 *WebController.go で再宣言していたものを
// 統合した（issue #4432、#4363 の adapter 層分）。json タグは移行前の各ゲームの
// 型と同一（playerIdx / card）なので、**RESTレスポンスの形状は不変**であり、
// フロントエンドの型定義・テスト・E2E は無変更で通る。
//
// Mighty は LeadDemandSuit / IsJokerLead を追加で持つ正当な別形状なので、
// domain.TrickCard と同じ理由で MightyWebOutputTrickCard を維持する。
type WebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// WebOutputCardHint 手札のカードを指すヒントの共通Web出力。
//
// 22ゲームが同一形状で再宣言していたものを統合した。
//
// 注意: *WebOutputHint という名前の型は全部で115個あり、そのうちこの形状
// （CardIndices + Reason）を共有するのは22個だけである。残り93個はソリティア
// 系の移動ヒント（fromZone/fromCol/toZone/toCol）やビッドヒントなど本当に
// 別物なので、統合してはならない。#4363 は「Hint 22型が同一形状」と読める
// 書き方をしているが、実際には「115型のうち22型」である。
type WebOutputCardHint struct {
	CardIndices []int  `json:"cardIndices"`
	Reason      string `json:"reason"`
}
