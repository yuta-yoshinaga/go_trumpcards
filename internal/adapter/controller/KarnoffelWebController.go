//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KarnoffelWebInput カルニッフェル Webインプット
type KarnoffelWebInput struct {
	BaseWebInput
	CardIndex *int                `json:"cardIndex,omitempty"`
	Config    *KarnoffelWebConfig `json:"config,omitempty"`
}

// KarnoffelWebConfig カルニッフェル Web設定
type KarnoffelWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetHands   *int `json:"targetHands,omitempty"`
}

// KarnoffelWebOutputPlayer カルニッフェル Webアウトプットプレイヤー
type KarnoffelWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0/2 が 0、1/3 が 1。
	Team      int `json:"team"`
	CardCount int `json:"cardCount"`
	// Cards は自分の手札のみ。
	Cards []*WebOutputCard `json:"cards"`
	// UpCard は配り始めに表向きで置かれた 1 枚。**切札はこの 4 枚のうち
	// 最も低い札が決める**ので、全員ぶん公開される。
	UpCard        *WebOutputCard `json:"upCard"`
	TricksWon     int            `json:"tricksWon"`
	IsDealer      bool           `json:"isDealer"`
	IsCurrentTurn bool           `json:"isCurrentTurn"`
}

// KarnoffelWebOutputResult カルニッフェル Webアウトプット局結果
type KarnoffelWebOutputResult struct {
	// WinnerTeam はこの局を取ったチーム (未決着なら -1)。
	WinnerTeam int                          `json:"winnerTeam"`
	Tricks     [domain.KarnoffelTeamCnt]int `json:"tricks"`
	ChosenSuit int                          `json:"chosenSuit"`
}

// KarnoffelWebOutput カルニッフェル Webアウトプット
type KarnoffelWebOutput struct {
	Players []*KarnoffelWebOutputPlayer `json:"players"`
	Phase   int                         `json:"phase"`
	// HandNumber は何局目か。
	HandNumber       int `json:"handNumber"`
	CurrentPlayerIdx int `json:"currentPlayerIdx"`
	DealerIdx        int `json:"dealerIdx"`
	// ChosenSuit は選ばれたスート。
	ChosenSuit int              `json:"chosenSuit"`
	Trick      []*WebOutputCard `json:"trick"`
	// ValidPlays は人間が出せる手札インデックス。**追随の義務は無いが、
	// 第 1 トリックのリードに悪魔は使えない。**
	ValidPlays     []int `json:"validPlays"`
	TrickLeaderIdx int   `json:"trickLeaderIdx"`
	TrickNumber    int   `json:"trickNumber"`
	// TeamTricks / HandsWon は [team0, team1]。
	TeamTricks [domain.KarnoffelTeamCnt]int `json:"teamTricks"`
	HandsWon   [domain.KarnoffelTeamCnt]int `json:"handsWon"`
	LastResult *KarnoffelWebOutputResult    `json:"lastResult"`
	// TricksToWin は 1 局を取るのに要るトリック数 (3)。
	TricksToWin int `json:"tricksToWin"`
	// HandSize は 1 人あたりの手札枚数 (**5**)。
	HandSize    int  `json:"handSize"`
	TargetHands int  `json:"targetHands"`
	GameEndFlag bool `json:"gameEndFlag"`
	WinnerTeam  int  `json:"winnerTeam"`
	WebOutputBase
	Config KarnoffelWebOutputConfig `json:"config"`
}

// KarnoffelWebOutputConfig カルニッフェル設定アウトプット
type KarnoffelWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetHands   int `json:"targetHands"`
}

// ToConfig builds a KarnoffelConfig from the nested web config, applying bounds checking.
func (c *KarnoffelWebConfig) ToConfig() domain.KarnoffelConfig {
	cfg := domain.DefaultKarnoffelConfig()
	cfg.CpuDifficulty = domain.KarnoffelCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.KarnoffelCpuDifficultyNormal), int(domain.KarnoffelCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	cfg.TargetHands = webutil.BoundedIntPtr(c.TargetHands,
		domain.KarnoffelMinTarget, domain.KarnoffelMaxTarget, cfg.TargetHands)
	return cfg
}

// ToConfig builds a KarnoffelConfig from the web input.
func (p KarnoffelWebInput) ToConfig() domain.KarnoffelConfig {
	return configOrDefault(p.Config, (*KarnoffelWebConfig).ToConfig, domain.DefaultKarnoffelConfig())
}

// KarnoffelWebController カルニッフェル Webコントローラークラス
type KarnoffelWebController = GameWebController[usecase.KarnoffelInteractorIF, KarnoffelWebInput, *KarnoffelWebOutput]

// NewKarnoffelWebController and NewKarnoffelWebControllerWithProvider are
// the standard and provider-backed constructors for KarnoffelWebController.
var NewKarnoffelWebController, NewKarnoffelWebControllerWithProvider = webControllerPair[usecase.KarnoffelInteractorIF, KarnoffelWebInput, *KarnoffelWebOutput](
	newKarnoffelDefaultOutput, karnoffelDispatch,
)

func newKarnoffelDefaultOutput(msg string) *KarnoffelWebOutput {
	return &KarnoffelWebOutput{
		Players:       make([]*KarnoffelWebOutputPlayer, 0),
		Trick:         make([]*WebOutputCard, 0),
		ValidPlays:    make([]int, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func karnoffelDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KarnoffelInteractorIF, param KarnoffelWebInput, newOut func(string) *KarnoffelWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ki.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newOut, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ki.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ki.NextHand())
	default:
		return dispatchLog(param.Command, bc, w, ki.ActionLog)
	}
	return true
}
