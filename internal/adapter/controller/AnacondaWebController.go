//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AnacondaWebConfig はアナコンダ (Anaconda) の Web 設定。
type AnacondaWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	Ante          *int `json:"ante,omitempty"`
	StartingChips *int `json:"startingChips,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// ToConfig は AnacondaWebConfig を domain.AnacondaConfig に変換する (境界チェック付き)。
func (c *AnacondaWebConfig) ToConfig() domain.AnacondaConfig {
	cfg := domain.DefaultAnacondaConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.AnacondaMinPlayerCount, domain.AnacondaMaxPlayerCount)
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, domain.AnacondaMinAnte, domain.AnacondaMaxAnte)
	webutil.ApplyBoundedInt(&cfg.StartingChips, c.StartingChips, domain.AnacondaMinStartingChips, domain.AnacondaMaxStartingChips)
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.AnacondaMinTargetRounds, domain.AnacondaMaxTargetRounds)
	return cfg
}

// AnacondaWebInput はアナコンダ Web インプット。
type AnacondaWebInput struct {
	BaseWebInput
	// CardIndices は pass / keep コマンドで選択するカードインデックス。
	CardIndices []int `json:"cardIndices,omitempty"`
	// Action は bet コマンドのアクション種別 ("call"/"raise"/"fold")。
	Action *string            `json:"action,omitempty"`
	Config *AnacondaWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.AnacondaConfig を構築する。
func (p AnacondaWebInput) ToConfig() domain.AnacondaConfig {
	return configOrDefault(p.Config, (*AnacondaWebConfig).ToConfig, domain.DefaultAnacondaConfig())
}

// AnacondaWebOutputPlayer は 1 プレイヤーの出力。
type AnacondaWebOutputPlayer struct {
	ID        int  `json:"id"`
	IsHuman   bool `json:"isHuman"`
	Chips     int  `json:"chips"`
	Folded    bool `json:"folded"`
	Out       bool `json:"out"`
	RoundBet  int  `json:"roundBet"`
	StreetBet int  `json:"streetBet"`
	CardCount int  `json:"cardCount"`
	// Cards は公開済みカード (人間は全手札、CPU はロールで公開された分、結果で全 5 枚)。
	Cards []*WebOutputCard `json:"cards"`
	// HandName は完全公開時の役名キー (非公開時は空文字)。
	HandName string `json:"handName,omitempty"`
	IsWinner bool   `json:"isWinner"`
}

// AnacondaWebOutputHint はヒント出力。
type AnacondaWebOutputHint struct {
	Action      string `json:"action"`
	CardIndices []int  `json:"cardIndices,omitempty"`
	Reason      string `json:"reason"`
}

// AnacondaWebOutputConfig は設定アウトプット。
type AnacondaWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	Ante          int `json:"ante"`
	StartingChips int `json:"startingChips"`
	TargetRounds  int `json:"targetRounds"`
}

// AnacondaWebOutput はアナコンダ Web アウトプット。
type AnacondaWebOutput struct {
	Players        []*AnacondaWebOutputPlayer `json:"players"`
	Phase          int                        `json:"phase"`
	RoundNumber    int                        `json:"roundNumber"`
	DealerIdx      int                        `json:"dealerIdx"`
	CurrentPlayer  int                        `json:"currentPlayer"`
	PassCount      int                        `json:"passCount"`
	RollIndex      int                        `json:"rollIndex"`
	Pot            int                        `json:"pot"`
	CurrentBet     int                        `json:"currentBet"`
	RaiseCount     int                        `json:"raiseCount"`
	MaxRaises      int                        `json:"maxRaises"`
	Ante           int                        `json:"ante"`
	Chips          int                        `json:"chips"`
	WinnerIdx      int                        `json:"winnerIdx"`
	MatchWinnerIdx int                        `json:"matchWinnerIdx"`
	Result         int                        `json:"result"`
	GameEndFlag    bool                       `json:"gameEndFlag"`
	IsHumanTurn    bool                       `json:"isHumanTurn"`
	CanRaise       bool                       `json:"canRaise"`
	Hint           *AnacondaWebOutputHint     `json:"hint,omitempty"`
	Config         AnacondaWebOutputConfig    `json:"config"`
	WebOutputBase
}

// AnacondaWebController はアナコンダ Web コントローラークラス。
type AnacondaWebController = GameWebController[usecase.AnacondaInteractorIF, AnacondaWebInput, *AnacondaWebOutput]

// NewAnacondaWebController, NewAnacondaWebControllerWithProvider are the standard and
// provider-backed constructors for AnacondaWebController.
var NewAnacondaWebController, NewAnacondaWebControllerWithProvider = webControllerPair[usecase.AnacondaInteractorIF, AnacondaWebInput, *AnacondaWebOutput](
	newAnacondaDefaultOutput, anacondaDispatch,
)

func newAnacondaDefaultOutput(msg string) *AnacondaWebOutput {
	return &AnacondaWebOutput{
		Players:        make([]*AnacondaWebOutputPlayer, 0),
		WinnerIdx:      -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func anacondaDispatch(bc *baseController, w http.ResponseWriter, ti usecase.AnacondaInteractorIF, param AnacondaWebInput, newDefault func(string) *AnacondaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "pass":
		if !requireParam(bc, w, newDefault, param.CardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Pass(param.CardIndices))
	case "keep":
		if !requireParam(bc, w, newDefault, param.CardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Keep(param.CardIndices))
	case "bet":
		if !requireParam(bc, w, newDefault, param.Action == nil, "param error: action is required.") {
			return true
		}
		bc.writePresenterResponse(w, anacondaBetResponse(ti, *param.Action))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}

// anacondaBetResponse は bet アクション ("call"/"raise"/"fold") をインタラクター呼び出しへ振り分ける。
func anacondaBetResponse(ti usecase.AnacondaInteractorIF, action string) string {
	switch action {
	case "raise":
		return ti.Raise()
	case "fold":
		return ti.Fold()
	default:
		return ti.Call()
	}
}
