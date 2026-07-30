//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MushiWebInput 虫Webインプット
type MushiWebInput struct {
	BaseWebInput
	CardIndex  *int            `json:"cardIndex,omitempty"`
	FieldIndex *int            `json:"fieldIndex,omitempty"`
	Config     *MushiWebConfig `json:"config,omitempty"`
}

// MushiWebConfig 虫Web設定
type MushiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// MushiWebOutputCard 場札・取り札 1 枚の出力。
type MushiWebOutputCard struct {
	*WebOutputCard
	// Month は札の月 (1..12、6 と 7 は無い)。design は花札では月を表すが、
	// ワイヤ上はスート名 ("CLOVER" 等) に変換されてしまうため、月を別に送る。
	// クライアントにスート名から月を復号させるのは筋が悪い。
	Month int `json:"month"`
	// Index は月内の番号 (1..4)。
	Index int `json:"index"`
	// Category は 0=カス / 1=短冊 / 2=種 / 3=光。得点計算をクライアントに
	// やり直させないために送る。
	Category int `json:"category"`
	Points   int `json:"points"`
	// IsWild は柳の雷札 (任意の札を取れる) かどうか。
	IsWild bool `json:"isWild"`
}

// MushiWebOutputYaku 成立した役 1 件。
type MushiWebOutputYaku struct {
	Key    string `json:"key"`
	Points string `json:"points"`
}

// MushiWebOutputPlayer 虫Webアウトプットプレイヤー
type MushiWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。伏せている間も送る -- 何枚持っているかは公開情報。
	CardCount int                   `json:"cardCount"`
	Cards     []*MushiWebOutputCard `json:"cards"`
	Captured  []*MushiWebOutputCard `json:"captured"`
	// CapturedPoints は取り札の合計点。
	CapturedPoints int `json:"capturedPoints"`
	Score          int `json:"score"`
	RoundResult    int `json:"roundResult"`
	// Hidden は手札が伏せられていることを示す。
	Hidden bool `json:"hidden"`
}

// MushiWebOutputHint ヒント出力
type MushiWebOutputHint struct {
	CardIndex  *int   `json:"cardIndex,omitempty"`
	FieldIndex *int   `json:"fieldIndex,omitempty"`
	Reason     string `json:"reason"`
}

// MushiWebOutput 虫Webアウトプット
type MushiWebOutput struct {
	Players          []*MushiWebOutputPlayer `json:"players"`
	Field            []*MushiWebOutputCard   `json:"field"`
	Phase            int                     `json:"phase"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	DealerIdx        int                     `json:"dealerIdx"`
	RoundNumber      int                     `json:"roundNumber"`
	TargetRounds     int                     `json:"targetRounds"`
	StockCount       int                     `json:"stockCount"`
	// PendingCard は選択待ちの札。SelectableIndices と対で使う。
	PendingCard *MushiWebOutputCard `json:"pendingCard,omitempty"`
	// SelectableIndices は選択フェーズで取れる場札の添字。ワイルドの
	// 「柳は取れない」規則をクライアントに再実装させないために送る。
	SelectableIndices []int               `json:"selectableIndices"`
	GameEndFlag       bool                `json:"gameEndFlag"`
	WinnerIdx         int                 `json:"winnerIdx"`
	Hint              *MushiWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config MushiWebOutputConfig `json:"config"`
}

// MushiWebOutputConfig 虫設定アウトプット
type MushiWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds a MushiConfig from the nested web config, applying bounds checking.
func (c *MushiWebConfig) ToConfig() domain.MushiConfig {
	cfg := domain.DefaultMushiConfig()
	cfg.CpuDifficulty = domain.MushiCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.MushiCpuDifficultyNormal), int(domain.MushiCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	cfg.TargetRounds = webutil.BoundedIntPtr(c.TargetRounds, 1, domain.MushiMaxRounds, cfg.TargetRounds)
	return cfg
}

// ToConfig builds a MushiConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *MushiWebConfig and calling the method on it would
// dereference nil.
func (i MushiWebInput) ToConfig() domain.MushiConfig {
	return configOrDefault(i.Config, (*MushiWebConfig).ToConfig, domain.DefaultMushiConfig())
}

// MushiWebController 虫Webコントローラ
type MushiWebController = GameWebController[usecase.MushiInteractorIF, MushiWebInput, *MushiWebOutput]

// NewMushiWebController and NewMushiWebControllerWithProvider are
// the standard and provider-backed constructors for MushiWebController.
var NewMushiWebController, NewMushiWebControllerWithProvider = webControllerPair[usecase.MushiInteractorIF, MushiWebInput, *MushiWebOutput](
	newMushiDefaultOutput, mushiDispatch,
)

func newMushiDefaultOutput(msg string) *MushiWebOutput {
	return &MushiWebOutput{
		Players:           make([]*MushiWebOutputPlayer, 0),
		Field:             make([]*MushiWebOutputCard, 0),
		SelectableIndices: make([]int, 0),
		WinnerIdx:         -1,
		TargetRounds:      domain.MushiMaxRounds,
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

func mushiDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MushiInteractorIF, param MushiWebInput, newDefault func(string) *MushiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, mi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.Play(*param.CardIndex))
	case "s", "select":
		if !requireParam(bc, w, newDefault, param.FieldIndex == nil, "param error: fieldIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.Select(*param.FieldIndex))
	case "n", "next":
		bc.writePresenterResponse(w, mi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, mi.Hint, mi.ActionLog)
	}
	return true
}

// NewMushiDefaultOutputForTest exposes the default-output builder to the
// external controller_test package, which cannot reach the unexported one.
func NewMushiDefaultOutputForTest(msg string) *MushiWebOutput { return newMushiDefaultOutput(msg) }
