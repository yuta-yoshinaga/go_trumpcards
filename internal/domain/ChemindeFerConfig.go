//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// 卓の大きさとシューの構成。
const (
	// ChemindeFerSeatCnt は卓に着く人数。席 0 が人間で、残りは CPU。
	//
	// **バンク (親) が席を順に回る**のがこのゲームの本体なので、卓は広めに取る。
	ChemindeFerSeatCnt = 6
	// ChemindeFerDeckCnt はシューに使うデッキ数 (6 デッキ 312 枚)。
	ChemindeFerDeckCnt = 6
	// ChemindeFerHandSize は最初に配る枚数 (親・子とも 2 枚)。
	ChemindeFerHandSize = 2
	// ChemindeFerMaxHandSize は 3 枚目を引いた後の上限。**4 枚目は無い。**
	ChemindeFerMaxHandSize = 3
)

// 合計値まわりの定数。**すべて mod 10** で、9 に近い方が勝つ。
const (
	// ChemindeFerNaturalMin はナチュラル (即決着) の下限。8 か 9 がナチュラル。
	ChemindeFerNaturalMin = 8
	// ChemindeFerPunterMustDrawMax はこれ以下なら子が**必ず引く**合計。
	ChemindeFerPunterMustDrawMax = 4
	// ChemindeFerPunterFreeTotal は子が**引くか引かないかを選べる**唯一の合計。
	ChemindeFerPunterFreeTotal = 5
	// ChemindeFerPunterMustStandMin はこれ以上なら子が**必ず立つ**合計。
	ChemindeFerPunterMustStandMin = 6
)

// チップとベットの範囲。
const (
	// ChemindeFerChipsMin / Max / Default は初期チップの範囲。
	ChemindeFerChipsMin     = 100
	ChemindeFerChipsMax     = 100000
	ChemindeFerDefaultChips = 1000

	// ChemindeFerRoundsMin / Max / Default は 1 ゲームのラウンド数の範囲。
	ChemindeFerRoundsMin     = 4
	ChemindeFerRoundsMax     = 50
	ChemindeFerDefaultRounds = 12

	// ChemindeFerStakeMin は親が張れる最小のバンク額。
	ChemindeFerStakeMin = 10
)

// ChemindeFerConfig はシュマン・ド・フェールのゲーム設定。
type ChemindeFerConfig struct {
	// Rounds は 1 ゲームのラウンド数。
	Rounds int
	// InitialChips は各席の初期チップ。**全席同額**で始める。
	InitialChips int
}

// DefaultChemindeFerConfig はデフォルト設定を返す。
func DefaultChemindeFerConfig() ChemindeFerConfig {
	return ChemindeFerConfig{
		Rounds:       ChemindeFerDefaultRounds,
		InitialChips: ChemindeFerDefaultChips,
	}
}

// Validate は設定値の妥当性を検証する。
func (c ChemindeFerConfig) Validate() error {
	if err := ValidateRange("rounds", c.Rounds, ChemindeFerRoundsMin, ChemindeFerRoundsMax); err != nil {
		return err
	}
	return ValidateRange("chips", c.InitialChips, ChemindeFerChipsMin, ChemindeFerChipsMax)
}

// chemindeFerConfigJSON is the JSON wire format for ChemindeFerConfig.
type chemindeFerConfigJSON struct {
	Rounds       int `json:"rd"`
	InitialChips int `json:"ic"`
}

// MarshalJSON implements json.Marshaler.
func (c ChemindeFerConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(chemindeFerConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *ChemindeFerConfig) UnmarshalJSON(data []byte) error {
	var j chemindeFerConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = ChemindeFerConfig(j)
	return c.Validate()
}

// ChemindeFerPhase は進行フェーズ。
type ChemindeFerPhase int

// フェーズ定数。
const (
	// ChemindeFerPhaseStake 親がバンク額 (張り金) を決める
	ChemindeFerPhaseStake ChemindeFerPhase = iota
	// ChemindeFerPhaseBet 子がバンク額を分け合って賭ける
	ChemindeFerPhaseBet
	// ChemindeFerPhasePunterDraw 子側が 3 枚目を引くか決める
	ChemindeFerPhasePunterDraw
	// ChemindeFerPhaseBankerDraw 親が 3 枚目を引くか決める
	ChemindeFerPhaseBankerDraw
	// ChemindeFerPhaseRoundEnd 決着後、次のラウンドを待つ
	ChemindeFerPhaseRoundEnd
)

// ChemindeFerPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const ChemindeFerPhaseMax = ChemindeFerPhaseRoundEnd

// ChemindeFerPhaseName はフェーズの識別子を返す (i18n キーの一部に使う)。
func ChemindeFerPhaseName(p ChemindeFerPhase) string {
	switch p {
	case ChemindeFerPhaseStake:
		return "stake"
	case ChemindeFerPhaseBet:
		return "bet"
	case ChemindeFerPhasePunterDraw:
		return "punterDraw"
	case ChemindeFerPhaseBankerDraw:
		return "bankerDraw"
	default:
		return "roundEnd"
	}
}

// ChemindeFerResult は 1 ラウンドの決着。
type ChemindeFerResult int

// 決着の種類。
const (
	// ChemindeFerResultNone まだ決着していない
	ChemindeFerResultNone ChemindeFerResult = iota
	// ChemindeFerResultBanker 親の勝ち
	ChemindeFerResultBanker
	// ChemindeFerResultPunter 子側の勝ち
	ChemindeFerResultPunter
	// ChemindeFerResultTie 引き分け (égalité)。**チップは動かない。**
	ChemindeFerResultTie
)

// ChemindeFerResultMax は最大の決着値 (復元時の範囲検査に使う)。
const ChemindeFerResultMax = ChemindeFerResultTie

// ChemindeFerResultName は決着の識別子を返す (i18n キーの一部に使う)。
func ChemindeFerResultName(r ChemindeFerResult) string {
	switch r {
	case ChemindeFerResultBanker:
		return "banker"
	case ChemindeFerResultPunter:
		return "punter"
	case ChemindeFerResultTie:
		return "tie"
	default:
		return "none"
	}
}

// ChemindeFerPunterMustDraw は子側が**引かされる**合計かを返す。
//
// バカラ (プント・バンコ) では親も子も引き方が表で固定されているが、
// シュマン・ド・フェールで固定されているのは**子の 0-4 と 6-7 だけ**で、
// 5 は子の自由、そして**親は常に自由**。ここがこのゲームの本体なので、
// 規則と戦略をはっきり分けて、規則側だけをこの 2 つの述語に閉じ込める。
func ChemindeFerPunterMustDraw(total int) bool {
	return total <= ChemindeFerPunterMustDrawMax
}

// ChemindeFerPunterMustStand は子側が**立たされる**合計かを返す。
func ChemindeFerPunterMustStand(total int) bool {
	return total >= ChemindeFerPunterMustStandMin
}

// ChemindeFerPunterMayChoose は子側に**選択の余地がある**合計かを返す。
// 真になるのは合計 5 のときだけ。
func ChemindeFerPunterMayChoose(total int) bool {
	return !ChemindeFerPunterMustDraw(total) && !ChemindeFerPunterMustStand(total)
}

// ChemindeFerIsNatural は 2 枚の合計がナチュラル (8 か 9) かを返す。
func ChemindeFerIsNatural(total int) bool { return total >= ChemindeFerNaturalMin }
