package controller

// WebOutputCard 共通Webアウトプットカード。
//
// 標準52枚デッキのカードは Design と Value のみを送出し、フロントエンドは
// 静的PNG (/images/{prefix}{NN}.png) で描画する。タロット・花札・カブ札・
// ウィザードのような非52枚デッキの札は、専用PNGアートを持たないため、
// 以下の任意フィールド（Glyph/Label/Color/Deck）で自己記述し、フロント
// エンドはこれらから手続き的（CSS/SVG）に描画する。Deck が非空のとき、
// フロントエンドは手続き描画パスへ切り替える。詳細は ADR-0033 を参照。
type WebOutputCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
	// Month は花札の月 (1..12) を表す。
	Month int `json:"month,omitempty"`
	// Kind は花札の種別コード。bright/animal/ribbon/chaff/gwang/yeol/tti/pi の 8 値で、
	// ゴーストップだけ韓国語由来の種別名を使う。
	Kind string `json:"kind,omitempty"`
	// Glyph は中央に描画する記号（例 "✦"）。非52枚デッキの札のみ設定する。
	Glyph string `json:"glyph,omitempty"`
	// Label は隅に描画するランク／名称ラベル（例 "Wizard", "XXI"）。
	Label string `json:"label,omitempty"`
	// Color は色調トークン（例 "red", "black", "purple", "green"）。
	Color string `json:"color,omitempty"`
	// Deck はデッキ系統ID（例 "wizard"）。非空なら手続き描画へ切り替える。
	Deck string `json:"deck,omitempty"`
	// Points はその札 1 枚の点数。**点数の合計で競うゲームだけが設定する**
	// (さくら: 20/10/5/1)。役で競うゲームでは意味が無いので省く。
	Points *int `json:"points,omitempty"`
}
