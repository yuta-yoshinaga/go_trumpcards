//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LobaWebInput ロバWebインプット
type LobaWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices はメルドに出す手札の添字集合。
	CardIndices []int          `json:"cardIndices,omitempty"`
	MeldIndex   *int           `json:"meldIndex,omitempty"`
	Config      *LobaWebConfig `json:"config,omitempty"`
}

// LobaWebConfig ロバWeb設定
type LobaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// LobaWebOutputPlayer ロバWebアウトプットプレイヤー
type LobaWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。伏せている間も送る。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Score は累計失点。101 で脱落するので、常に見えている必要がある。
	Score int `json:"score"`
	// Eliminated は 101 点に達して脱落したか。
	Eliminated bool `json:"eliminated"`
	// HasMelded はこのラウンドで既に出したか。レイオフの前提。
	HasMelded bool `json:"hasMelded"`
	Hidden    bool `json:"hidden"`
}

// LobaWebOutputMeld 場のメルド
type LobaWebOutputMeld struct {
	Owner int `json:"owner"`
	// Kind は 0=ピエルナ (同ランク・異なる3スート), 1=エスカレラ (同スートの並び)。
	Kind  int              `json:"kind"`
	Cards []*WebOutputCard `json:"cards"`
}

// LobaWebOutputHint ヒント出力
type LobaWebOutputHint struct {
	// CardIndices はメルドに出せる手札の添字 (あれば)。
	CardIndices []int `json:"cardIndices,omitempty"`
	CardIndex   *int  `json:"cardIndex,omitempty"`
	// DrawStock が真なら山札から引くのが推奨手。
	DrawStock bool   `json:"drawStock"`
	Reason    string `json:"reason"`
}

// LobaWebOutput ロバWebアウトプット
type LobaWebOutput struct {
	Players          []*LobaWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	StockCount       int                    `json:"stockCount"`
	// DiscardTop は捨て札の一番上。取るかどうかの判断材料。
	DiscardTop *WebOutputCard       `json:"discardTop,omitempty"`
	Melds      []*LobaWebOutputMeld `json:"melds"`
	RoundNo    int                  `json:"roundNo"`
	// KnockOut は脱落する失点 (101)。クライアントに書き写させない。
	KnockOut int `json:"knockOut"`
	// RoundWinner は直近ラウンドで上がった人 (-1: なし)。
	RoundWinner int `json:"roundWinner"`
	// RoundClean は直近の上がりが「一度も出さずに一気に」だったか。-10 の根拠。
	RoundClean  bool               `json:"roundClean"`
	GameEndFlag bool               `json:"gameEndFlag"`
	WinnerIdx   int                `json:"winnerIdx"`
	Hint        *LobaWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config LobaWebOutputConfig `json:"config"`
}

// LobaWebOutputConfig ロバ設定アウトプット
type LobaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a LobaConfig from the nested web config, applying bounds checking.
func (c *LobaWebConfig) ToConfig() domain.LobaConfig {
	cfg := domain.DefaultLobaConfig()
	cfg.CpuDifficulty = domain.LobaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.LobaCpuDifficultyNormal), int(domain.LobaCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a LobaConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *LobaWebConfig and calling the method on it would
// dereference nil.
func (i LobaWebInput) ToConfig() domain.LobaConfig {
	return configOrDefault(i.Config, (*LobaWebConfig).ToConfig, domain.DefaultLobaConfig())
}

// LobaWebController ロバWebコントローラ
type LobaWebController = GameWebController[usecase.LobaInteractorIF, LobaWebInput, *LobaWebOutput]

// NewLobaWebController and NewLobaWebControllerWithProvider are the standard
// and provider-backed constructors for LobaWebController.
var NewLobaWebController, NewLobaWebControllerWithProvider = webControllerPair[usecase.LobaInteractorIF, LobaWebInput, *LobaWebOutput](
	newLobaDefaultOutput, lobaDispatch,
)

func newLobaDefaultOutput(msg string) *LobaWebOutput {
	return &LobaWebOutput{
		Players:       make([]*LobaWebOutputPlayer, 0),
		Melds:         make([]*LobaWebOutputMeld, 0),
		KnockOut:      domain.LobaKnockOut,
		RoundWinner:   -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func lobaDispatch(bc *baseController, w http.ResponseWriter, li usecase.LobaInteractorIF, param LobaWebInput, newDefault func(string) *LobaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, li.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, li.DrawStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, li.DrawDiscard())
	case "m", "meld":
		if !requireParam(bc, w, newDefault, len(param.CardIndices) == 0, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.Meld(param.CardIndices))
	case "o", "layoff":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil || param.MeldIndex == nil,
			"param error: cardIndex and meldIndex are required.") {
			return true
		}
		bc.writePresenterResponse(w, li.LayOff(*param.CardIndex, *param.MeldIndex))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.Discard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, li.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, li.Hint, li.ActionLog)
	}
	return true
}

// NewLobaDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
func NewLobaDefaultOutputForTest(msg string) *LobaWebOutput {
	return newLobaDefaultOutput(msg)
}
