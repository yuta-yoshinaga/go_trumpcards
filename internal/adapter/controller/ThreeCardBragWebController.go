//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ThreeCardBragWebInput スリーカード・ブラグのWebインプット
type ThreeCardBragWebInput struct {
	BaseWebInput
	// RaiseStake raise コマンドで引き上げる賭け単位
	RaiseStake *int `json:"raiseStake,omitempty"`
	// Config ゲーム設定
	Config *ThreeCardBragWebConfig `json:"config,omitempty"`
}

// ThreeCardBragWebConfig スリーカード・ブラグのWeb設定
type ThreeCardBragWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	Ante          *int `json:"ante,omitempty"`
	StartingChips *int `json:"startingChips,omitempty"`
}

// ThreeCardBragWebOutputPlayer スリーカード・ブラグのWebアウトプットプレイヤー
type ThreeCardBragWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Chips     int              `json:"chips"`
	Seen      bool             `json:"seen"`
	Folded    bool             `json:"folded"`
	Out       bool             `json:"out"`
	RoundBet  int              `json:"roundBet"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// HandName ショーダウンで公開された手の役名 (非公開時は空文字)
	HandName string `json:"handName,omitempty"`
}

// ThreeCardBragWebOutputHint ヒント出力
type ThreeCardBragWebOutputHint struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// ThreeCardBragWebOutputConfig スリーカード・ブラグの設定アウトプット
type ThreeCardBragWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	Ante          int `json:"ante"`
	StartingChips int `json:"startingChips"`
}

// ThreeCardBragWebOutput スリーカード・ブラグのWebアウトプット
type ThreeCardBragWebOutput struct {
	Players          []*ThreeCardBragWebOutputPlayer `json:"players"`
	Pot              int                             `json:"pot"`
	Stake            int                             `json:"stake"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	DealerIdx        int                             `json:"dealerIdx"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	RoundWinnerIdx   int                             `json:"roundWinnerIdx"`
	MatchWinnerIdx   int                             `json:"matchWinnerIdx"`
	IsShowdown       bool                            `json:"isShowdown"`
	CanShow          bool                            `json:"canShow"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	IsHumanTurn      bool                            `json:"isHumanTurn"`
	Hint             *ThreeCardBragWebOutputHint     `json:"hint,omitempty"`
	Config           ThreeCardBragWebOutputConfig    `json:"config"`
	WebOutputBase
}

// ToConfig builds a ThreeCardBragConfig from the nested web config, applying bounds checking.
func (c *ThreeCardBragWebConfig) ToConfig() domain.ThreeCardBragConfig {
	cfg := domain.DefaultThreeCardBragConfig()
	cfg.CpuDifficulty = domain.ThreeCardBragCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.ThreeCardBragCpuDifficultyEasy), int(domain.ThreeCardBragCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, 1, 1000)
	webutil.ApplyBoundedInt(&cfg.StartingChips, c.StartingChips, 2, domain.ThreeCardBragMaxStartingChips)
	return cfg
}

// ToConfig builds a ThreeCardBragConfig from the web input.
func (p ThreeCardBragWebInput) ToConfig() domain.ThreeCardBragConfig {
	return configOrDefault(p.Config, (*ThreeCardBragWebConfig).ToConfig, domain.DefaultThreeCardBragConfig())
}

// ThreeCardBragWebController スリーカード・ブラグのWebコントローラークラス
type ThreeCardBragWebController = GameWebController[usecase.ThreeCardBragInteractorIF, ThreeCardBragWebInput, *ThreeCardBragWebOutput]

// NewThreeCardBragWebController and NewThreeCardBragWebControllerWithProvider are
// the standard and provider-backed constructors for ThreeCardBragWebController.
var NewThreeCardBragWebController, NewThreeCardBragWebControllerWithProvider = webControllerPair[usecase.ThreeCardBragInteractorIF, ThreeCardBragWebInput, *ThreeCardBragWebOutput](
	newThreeCardBragDefaultOutput, threeCardBragDispatch,
)

func newThreeCardBragDefaultOutput(msg string) *ThreeCardBragWebOutput {
	return &ThreeCardBragWebOutput{
		Players:        make([]*ThreeCardBragWebOutputPlayer, 0),
		RoundWinnerIdx: -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func threeCardBragDispatch(bc *baseController, w http.ResponseWriter, ti usecase.ThreeCardBragInteractorIF, param ThreeCardBragWebInput, newDefault func(string) *ThreeCardBragWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "see":
		bc.writePresenterResponse(w, ti.See())
	case "bet":
		bc.writePresenterResponse(w, ti.Bet())
	case "raise":
		if !requireParam(bc, w, newDefault, param.RaiseStake == nil, "param error: raiseStake is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Raise(*param.RaiseStake))
	case "fold":
		bc.writePresenterResponse(w, ti.Fold())
	case "show":
		bc.writePresenterResponse(w, ti.Show())
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
