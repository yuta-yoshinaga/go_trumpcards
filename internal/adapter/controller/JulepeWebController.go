//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// JulepeWebInput フレペWebインプット
type JulepeWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *JulepeWebConfig `json:"config,omitempty"`
}

// JulepeWebConfig フレペWeb設定
type JulepeWebConfig struct {
	// PlayerCnt は 3〜5 人。**フレペは可変人数が特徴。**
	PlayerCnt *int `json:"playerCnt,omitempty"`
	Rounds    *int `json:"rounds,omitempty"`
}

// JulepeWebOutputPlayer フレペWebアウトプットプレイヤー
type JulepeWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Chips は持ちチップ。**多いほど良い。**
	Chips int `json:"chips"`
	// InRound はこのラウンドに参加しているか、Decided は選び終えたか。
	InRound     bool `json:"inRound"`
	Decided     bool `json:"decided"`
	RoundTricks int  `json:"roundTricks"`
	TrickCount  int  `json:"trickCount"`
}

// JulepeWebOutputHint ヒント出力
type JulepeWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// JulepeWebOutput フレペWebアウトプット
type JulepeWebOutput struct {
	Players     []*JulepeWebOutputPlayer `json:"players"`
	Phase       int                      `json:"phase"`
	RoundNumber int                      `json:"roundNumber"`
	TrickNumber int                      `json:"trickNumber"`
	Pot         int                      `json:"pot"`
	// RequiredTricks は現在の参加人数に対する規定トリック数。**人数で変わる**
	// ので、画面が固定値で説明するとルールを取り違える。
	RequiredTricks int `json:"requiredTricks"`
	// Beast は次ラウンドのアンティが倍になる席。
	Beast            []bool                `json:"beast"`
	TrumpSuit        int                   `json:"trumpSuit"`
	UpCard           *WebOutputCard        `json:"upCard,omitempty"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                   `json:"leadPlayerIdx"`
	DealerIdx        int                   `json:"dealerIdx"`
	ActiveCount      int                   `json:"activeCount"`
	CurrentTrick     []*WebOutputTrickCard `json:"currentTrick"`
	ValidPlays       []int                 `json:"validPlays"`
	GameEndFlag      bool                  `json:"gameEndFlag"`
	WinnerIdx        int                   `json:"winnerIdx"`
	Hint             *JulepeWebOutputHint  `json:"hint,omitempty"`
	WebOutputBase
	Config JulepeWebOutputConfig `json:"config"`
}

// JulepeWebOutputConfig フレペ設定アウトプット
type JulepeWebOutputConfig struct {
	PlayerCnt int `json:"playerCnt"`
	Rounds    int `json:"rounds"`
}

// ToConfig builds a JulepeConfig from the nested web config, applying bounds checking.
func (c *JulepeWebConfig) ToConfig() domain.JulepeConfig {
	cfg := domain.DefaultJulepeConfig()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.JulepePlayerCntMin, domain.JulepePlayerCntMax, cfg.PlayerCnt)
	cfg.Rounds = webutil.BoundedIntPtr(c.Rounds,
		domain.JulepeRoundsMin, domain.JulepeRoundsMax, cfg.Rounds)
	return cfg
}

// ToConfig builds a JulepeConfig from the web input.
func (p JulepeWebInput) ToConfig() domain.JulepeConfig {
	return configOrDefault(p.Config, (*JulepeWebConfig).ToConfig, domain.DefaultJulepeConfig())
}

// JulepeWebController フレペWebコントローラークラス
type JulepeWebController = GameWebController[usecase.JulepeInteractorIF, JulepeWebInput, *JulepeWebOutput]

// NewJulepeWebController and NewJulepeWebControllerWithProvider are
// the standard and provider-backed constructors for JulepeWebController.
var NewJulepeWebController, NewJulepeWebControllerWithProvider = webControllerPair[usecase.JulepeInteractorIF, JulepeWebInput, *JulepeWebOutput](
	newJulepeDefaultOutput, julepeDispatch,
)

func newJulepeDefaultOutput(msg string) *JulepeWebOutput {
	return &JulepeWebOutput{
		Players:       make([]*JulepeWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func julepeDispatch(bc *baseController, w http.ResponseWriter, ri usecase.JulepeInteractorIF, param JulepeWebInput, newDefault func(string) *JulepeWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ri.ResetWithConfig(param.ToConfig()))
	case "in", "play":
		bc.writePresenterResponse(w, ri.Play())
	case "out", "pass":
		bc.writePresenterResponse(w, ri.Pass())
	case "c", "card":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ri.PlayCard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ri.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, ri.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ri.Hint, ri.ActionLog)
	}
	return true
}
