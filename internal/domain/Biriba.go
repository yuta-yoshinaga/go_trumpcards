//go:build !js || !wasm || extra

package domain

// Biriba (ビリバ) is implemented as a configured Canasta: the Pozzetto
// reserve-hand mechanic, the 11-card deal, and the take-pozzetto-before-going-out
// rule all live in the Canasta domain, gated by CanastaConfig.UsePozzetto.
//
// Biriba is exposed through type aliases rather than a second domain type on
// purpose. The domain package is linked into every Cloudflare Worker WASM
// binary, and TinyGo conservatively retains every json.Marshaler /
// json.Unmarshaler implementation it finds — so a standalone Biriba type would
// ship its serialisation code in all three workers and push the classic worker
// (already at the 1 MB gzip free-tier limit) over the edge. Aliasing keeps the
// footprint at zero new types.

// Biriba はビリバゲーム（= ポゼット有効化した Canasta）。
type Biriba = Canasta

// BiribaPlayer はビリバプレイヤー。
type BiribaPlayer = CanastaPlayer

// BiribaConfig はビリバ設定。
type BiribaConfig = CanastaConfig

// BiribaMeld はビリバのメルド。
type BiribaMeld = CanastaMeld

// BiribaPhase はビリバのフェーズ型。
type BiribaPhase = CanastaPhase

// BiribaCpuDifficulty はビリバの CPU 難易度型。
type BiribaCpuDifficulty = CanastaCpuDifficulty

// BiribaHint はビリバのヒント情報。
type BiribaHint = CanastaHint

// ビリバのフェーズ定数（Canasta と同一値）。
const (
	BiribaPhaseDraw     = CanastaPhaseDraw
	BiribaPhaseMeld     = CanastaPhaseMeld
	BiribaPhaseDiscard  = CanastaPhaseDiscard
	BiribaPhaseRoundEnd = CanastaPhaseRoundEnd
	BiribaPhaseGameEnd  = CanastaPhaseGameEnd
)

// ビリバの CPU 難易度定数（Canasta と同一値）。
const (
	BiribaCpuDifficultyEasy   = CanastaCpuDifficultyEasy
	BiribaCpuDifficultyNormal = CanastaCpuDifficultyNormal
	BiribaCpuDifficultyHard   = CanastaCpuDifficultyHard
)

// BiribaHandSize ビリバの初期配布枚数。
const BiribaHandSize = CanastaBiribaHandSize

// BiribaPozzettoSize ビリバのポゼット1山の枚数。
const BiribaPozzettoSize = CanastaPozzettoSize

// BiribaDefaultPointLimit ビリバのデフォルト目標スコア。
const BiribaDefaultPointLimit = 2005

// DefaultBiribaConfig はビリバのデフォルト設定（ポゼット有効, 2005点）を返す。
func DefaultBiribaConfig() CanastaConfig {
	cfg := DefaultCanastaConfig()
	cfg.UsePozzetto = true
	cfg.PointLimit = BiribaDefaultPointLimit
	return cfg
}

// NewBiribaPlayer はビリバプレイヤーを生成する（CanastaPlayer と同一）。
func NewBiribaPlayer(isHuman bool) *CanastaPlayer { return NewCanastaPlayer(isHuman) }

// NewBiriba はビリバゲームを生成する（ポゼット有効の Canasta）。
func NewBiriba(trumpCards *TrumpCards, players []*CanastaPlayer, config CanastaConfig) *Canasta {
	return NewCanasta(trumpCards, players, config)
}

// NewDefaultBiriba は標準的な2人ビリバ（人間1 + CPU1, 108枚デッキ）を生成する。
func NewDefaultBiriba() *Canasta {
	players := []*CanastaPlayer{
		NewCanastaPlayer(true),
		NewCanastaPlayer(false),
	}
	return NewCanasta(NewTrumpCardsWithDecks(2, 4), players, DefaultBiribaConfig())
}
