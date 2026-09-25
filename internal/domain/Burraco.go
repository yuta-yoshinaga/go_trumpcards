//go:build !js || !wasm || extra7

package domain

// Burraco (ブラーコ) is implemented as a configured Canasta: the Pozzetto
// reserve-hand mechanic, the 11-card deal, and the take-pozzetto-before-going-out
// rule all live in the Canasta domain, gated by CanastaConfig.UsePozzetto.
//
// Burraco is exposed through type aliases rather than a second domain type on
// purpose. Burraco can be expressed entirely through Canasta's existing
// configuration surface (UsePozzetto), so introducing a standalone type would
// add serialisation code without adding any new behaviour.

// Burraco はブラーコゲーム（= ポゼット有効化した Canasta）。
type Burraco = Canasta

// BurracoPlayer はブラーコプレイヤー。
type BurracoPlayer = CanastaPlayer

// BurracoConfig はブラーコ設定。
type BurracoConfig = CanastaConfig

// BurracoMeld はブラーコのメルド。
type BurracoMeld = CanastaMeld

// BurracoPhase はブラーコのフェーズ型。
type BurracoPhase = CanastaPhase

// BurracoCpuDifficulty はブラーコの CPU 難易度型。
type BurracoCpuDifficulty = CanastaCpuDifficulty

// BurracoHint はブラーコのヒント情報。
type BurracoHint = CanastaHint

// ブラーコのフェーズ定数（Canasta と同一値）。
const (
	BurracoPhaseDraw     = CanastaPhaseDraw
	BurracoPhaseMeld     = CanastaPhaseMeld
	BurracoPhaseDiscard  = CanastaPhaseDiscard
	BurracoPhaseRoundEnd = CanastaPhaseRoundEnd
	BurracoPhaseGameEnd  = CanastaPhaseGameEnd
)

// ブラーコの CPU 難易度定数（Canasta と同一値）。
const (
	BurracoCpuDifficultyEasy   = CanastaCpuDifficultyEasy
	BurracoCpuDifficultyNormal = CanastaCpuDifficultyNormal
	BurracoCpuDifficultyHard   = CanastaCpuDifficultyHard
)

// BurracoHandSize ブラーコの初期配布枚数。
const BurracoHandSize = CanastaBurracoHandSize

// BurracoPozzettoSize ブラーコのポゼット1山の枚数。
const BurracoPozzettoSize = CanastaPozzettoSize

// BurracoDefaultPointLimit ブラーコのデフォルト目標スコア。
const BurracoDefaultPointLimit = 2005

// DefaultBurracoConfig はブラーコのデフォルト設定（ポゼット有効, 2005点）を返す。
func DefaultBurracoConfig() CanastaConfig {
	cfg := DefaultCanastaConfig()
	cfg.UsePozzetto = true
	cfg.PointLimit = BurracoDefaultPointLimit
	return cfg
}

// NewBurracoPlayer はブラーコプレイヤーを生成する（CanastaPlayer と同一）。
func NewBurracoPlayer(isHuman bool) *CanastaPlayer { return NewCanastaPlayer(isHuman) }

// NewBurraco はブラーコゲームを生成する（ポゼット有効の Canasta）。
func NewBurraco(trumpCards *TrumpCards, players []*CanastaPlayer, config CanastaConfig) *Canasta {
	return NewCanasta(trumpCards, players, config)
}

// NewDefaultBurraco は標準的な2人ブラーコ（人間1 + CPU1, 108枚デッキ）を生成する。
func NewDefaultBurraco() *Canasta {
	players := []*CanastaPlayer{
		NewCanastaPlayer(true),
		NewCanastaPlayer(false),
	}
	return NewCanasta(NewTrumpCardsWithDecks(2, 4), players, DefaultBurracoConfig())
}
