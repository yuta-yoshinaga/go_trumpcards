//go:build !js || !wasm || casino

package domain

import "encoding/json"

// シューの構成。
const (
	// DoubleAttackDeckCount はシューに使うデッキ数。
	//
	// **10 を抜いた 48 枚のスパニッシュデッキ**を 8 組 = 384 枚。
	DoubleAttackDeckCount = 8
	// DoubleAttackDeckSize は 1 組の枚数 (10 を除いた 48 枚)。
	DoubleAttackDeckSize = 48
)

// チップとベットの範囲。
const (
	DoubleAttackChipsMin     = 100
	DoubleAttackChipsMax     = 100000
	DoubleAttackDefaultChips = 1000

	DoubleAttackAnteMin     = 10
	DoubleAttackAnteMax     = 500
	DoubleAttackDefaultAnte = 50
)

// DoubleAttackBustItMin / Max は Bust It サイドベットの範囲 (0 は置かない)。
const (
	DoubleAttackBustItMax = 200
)

// DoubleAttackBlackjackPayoutNum / Den はプレイヤーのブラックジャック配当。
//
// **1:1 に抑えられている。** 通常のブラックジャックの 3:2 ではない ── アップカードを
// 見てから賭け増しできる有利さの対価で、ここを 3:2 にすると控除率が消える。
const (
	DoubleAttackBlackjackPayoutNum = 1
	DoubleAttackBlackjackPayoutDen = 1
)

// DoubleAttackDealerHitsSoft17 はディーラーがソフト 17 でヒットするか。
const DoubleAttackDealerHitsSoft17 = true

// doubleAttackBustItPayouts は Bust It の配当 (X:1)。キーはバスト時の手札枚数。
//
// **issue の配当表は使えなかった。** issue は「3 枚バスト 15:1 程度〜8 枚バスト
// 200:1 程度」としているが、実際に 48 枚 × 8 のシューでディーラーを 200 万回
// 進行させて数えると分布はこうなる (ソフト 17 ヒット):
//
//	バスト率       27.5620%
//	 3 枚バスト    15.30965%   公正オッズ    5.53:1
//	 4 枚バスト     9.40075%   公正オッズ    9.64:1
//	 5 枚バスト     2.45730%   公正オッズ   39.70:1
//	 6 枚バスト     0.35905%   公正オッズ  277.51:1
//	 7 枚バスト     0.03280%   公正オッズ 3047.78:1
//	 8 枚以上       0.00235%   公正オッズ 42552:1
//
// **3 枚バストは 15% 起きる**ので、そこに 15:1 を払うと期待値は +476% ──
// 置くだけで勝てるボタンになる。issue の表は端が入れ替わっているとみられる
// (15:1 は 6 枚あたり、200:1 は 8 枚の値に近い)。
//
// この表での**ハウスエッジは 5.212%** (期待値 -0.0521)。サイドベットとして妥当な水準。
var doubleAttackBustItPayouts = map[int]int{
	3: 1,
	4: 2,
	5: 8,
	6: 25,
	7: 100,
	8: 500, // 8 枚以上
}

// DoubleAttackBustItMaxCards は配当表が個別に持つ最大の枚数。これ以上は同じ配当。
const DoubleAttackBustItMaxCards = 8

// DoubleAttackBustItPayout はバスト枚数に対する Bust It の配当倍率を返す。
//
// **8 枚以上はすべて同じ**。個別に刻んでも実測で 200 万回に 2 回しか出ない。
func DoubleAttackBustItPayout(cards int) int {
	if cards >= DoubleAttackBustItMaxCards {
		return doubleAttackBustItPayouts[DoubleAttackBustItMaxCards]
	}
	return doubleAttackBustItPayouts[cards]
}

// DoubleAttackBlackjackConfig は追加ベット・ブラックジャックのゲーム設定。
type DoubleAttackBlackjackConfig struct {
	// InitialChips は初期チップ。
	InitialChips int
	// DefaultAnte はラウンド開始時に選んでおくアンティ額。
	DefaultAnte int
}

// DefaultDoubleAttackBlackjackConfig はデフォルト設定を返す。
func DefaultDoubleAttackBlackjackConfig() DoubleAttackBlackjackConfig {
	return DoubleAttackBlackjackConfig{
		InitialChips: DoubleAttackDefaultChips,
		DefaultAnte:  DoubleAttackDefaultAnte,
	}
}

// Validate は設定値の妥当性を検証する。
func (c DoubleAttackBlackjackConfig) Validate() error {
	if err := ValidateRange("chips", c.InitialChips,
		DoubleAttackChipsMin, DoubleAttackChipsMax); err != nil {
		return err
	}
	return ValidateRange("ante", c.DefaultAnte, DoubleAttackAnteMin, DoubleAttackAnteMax)
}

// doubleAttackConfigJSON is the JSON wire format for DoubleAttackBlackjackConfig.
type doubleAttackConfigJSON struct {
	InitialChips int `json:"ic"`
	DefaultAnte  int `json:"da"`
}

// MarshalJSON implements json.Marshaler.
func (c DoubleAttackBlackjackConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(doubleAttackConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *DoubleAttackBlackjackConfig) UnmarshalJSON(data []byte) error {
	var j doubleAttackConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = DoubleAttackBlackjackConfig(j)
	return c.Validate()
}

// DoubleAttackPhase は進行フェーズ。
type DoubleAttackPhase int

// フェーズ定数。
const (
	// DoubleAttackPhaseBet アンティと任意の Bust It を賭ける
	DoubleAttackPhaseBet DoubleAttackPhase = iota
	// DoubleAttackPhaseAttack アップカードを見て追加ベットを決める
	DoubleAttackPhaseAttack
	// DoubleAttackPhasePlay ヒット / スタンド / ダブル / スプリット
	DoubleAttackPhasePlay
	// DoubleAttackPhaseResult 決着後
	DoubleAttackPhaseResult
)

// DoubleAttackPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const DoubleAttackPhaseMax = DoubleAttackPhaseResult

// DoubleAttackPhaseName はフェーズの識別子を返す (i18n キーの一部に使う)。
func DoubleAttackPhaseName(p DoubleAttackPhase) string {
	switch p {
	case DoubleAttackPhaseBet:
		return "bet"
	case DoubleAttackPhaseAttack:
		return "attack"
	case DoubleAttackPhasePlay:
		return "play"
	default:
		return "result"
	}
}
