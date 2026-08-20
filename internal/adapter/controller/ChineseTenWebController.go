//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ChineseTenWebInput 撿紅點Webインプット
type ChineseTenWebInput struct {
	BaseWebInput
	CardIndex   *int                 `json:"cardIndex,omitempty"`
	LayoutIndex *int                 `json:"layoutIndex,omitempty"`
	Config      *ChineseTenWebConfig `json:"config,omitempty"`
}

// ChineseTenWebConfig 撿紅點Web設定
type ChineseTenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// ChineseTenWebOutputCard 1 枚の出力。
type ChineseTenWebOutputCard struct {
	*WebOutputCard
	// Points はこの札の得点。赤札だけが 0 でない。得点表をクライアントに
	// もう一部持たせないために送る。
	Points int `json:"points"`
	// IsRed は得点する側の札か。UI が赤札を目立たせるために使う。
	IsRed bool `json:"isRed"`
}

// ChineseTenWebOutputPlayer 撿紅點Webアウトプットプレイヤー
type ChineseTenWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。伏せている間も送る -- 何枚持っているかは公開情報。
	CardCount int                        `json:"cardCount"`
	Cards     []*ChineseTenWebOutputCard `json:"cards"`
	// Captured は取り札。**両者とも公開**する -- 何が取られたかを見て残りを
	// 読むのがフィッシング系の骨格で、隠すとゲームが成立しない。
	Captured []*ChineseTenWebOutputCard `json:"captured"`
	Score    int                        `json:"score"`
	Hidden   bool                       `json:"hidden"`
}

// ChineseTenWebOutputHint ヒント出力
type ChineseTenWebOutputHint struct {
	CardIndex   *int   `json:"cardIndex,omitempty"`
	LayoutIndex *int   `json:"layoutIndex,omitempty"`
	Reason      string `json:"reason"`
}

// ChineseTenWebOutput 撿紅點Webアウトプット
type ChineseTenWebOutput struct {
	Players          []*ChineseTenWebOutputPlayer `json:"players"`
	Layout           []*ChineseTenWebOutputCard   `json:"layout"`
	Phase            int                          `json:"phase"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	StockCount       int                          `json:"stockCount"`
	// PendingCard は選択待ちの札。SelectableIndices と対で使う。
	PendingCard *ChineseTenWebOutputCard `json:"pendingCard,omitempty"`
	// SelectableIndices は選択フェーズで取れる場札の添字。捕獲規則
	// (A〜9 は合計10・10〜K は同ランク) をクライアントに再実装させない。
	SelectableIndices []int `json:"selectableIndices"`
	// TieScore は引き分けとなる点 (赤札総点の半分)。
	TieScore    int                      `json:"tieScore"`
	GameEndFlag bool                     `json:"gameEndFlag"`
	WinnerIdx   int                      `json:"winnerIdx"`
	Hint        *ChineseTenWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config ChineseTenWebOutputConfig `json:"config"`
}

// ChineseTenWebOutputConfig 撿紅點設定アウトプット
type ChineseTenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a ChineseTenConfig from the nested web config, applying bounds checking.
func (c *ChineseTenWebConfig) ToConfig() domain.ChineseTenConfig {
	cfg := domain.DefaultChineseTenConfig()
	cfg.CpuDifficulty = domain.ChineseTenCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.ChineseTenCpuDifficultyNormal), int(domain.ChineseTenCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a ChineseTenConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *ChineseTenWebConfig and calling the method on it
// would dereference nil.
func (i ChineseTenWebInput) ToConfig() domain.ChineseTenConfig {
	return configOrDefault(i.Config, (*ChineseTenWebConfig).ToConfig, domain.DefaultChineseTenConfig())
}

// ChineseTenWebController 撿紅點Webコントローラ
type ChineseTenWebController = GameWebController[usecase.ChineseTenInteractorIF, ChineseTenWebInput, *ChineseTenWebOutput]

// NewChineseTenWebController and NewChineseTenWebControllerWithProvider are
// the standard and provider-backed constructors for ChineseTenWebController.
var NewChineseTenWebController, NewChineseTenWebControllerWithProvider = webControllerPair[usecase.ChineseTenInteractorIF, ChineseTenWebInput, *ChineseTenWebOutput](
	newChineseTenDefaultOutput, chineseTenDispatch,
)

func newChineseTenDefaultOutput(msg string) *ChineseTenWebOutput {
	return &ChineseTenWebOutput{
		Players:           make([]*ChineseTenWebOutputPlayer, 0),
		Layout:            make([]*ChineseTenWebOutputCard, 0),
		SelectableIndices: make([]int, 0),
		TieScore:          domain.ChineseTenTieScore,
		WinnerIdx:         -1,
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

func chineseTenDispatch(bc *baseController, w http.ResponseWriter, ci usecase.ChineseTenInteractorIF, param ChineseTenWebInput, newDefault func(string) *ChineseTenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.CardIndex))
	case "s", "select":
		if !requireParam(bc, w, newDefault, param.LayoutIndex == nil, "param error: layoutIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Select(*param.LayoutIndex))
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
