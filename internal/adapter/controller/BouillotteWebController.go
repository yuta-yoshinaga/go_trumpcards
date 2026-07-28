//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BouillotteWebConfig はブイヨット (Bouillotte) の Web 設定。
type BouillotteWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	Ante          *int `json:"ante,omitempty"`
	StartingChips *int `json:"startingChips,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// ToConfig は BouillotteWebConfig を domain.BouillotteConfig に変換する (境界チェック付き)。
func (c *BouillotteWebConfig) ToConfig() domain.BouillotteConfig {
	cfg := domain.DefaultBouillotteConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.BouillotteMinPlayerCount, domain.BouillotteMaxPlayerCount)
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, domain.BouillotteMinAnte, domain.BouillotteMaxAnte)
	webutil.ApplyBoundedInt(&cfg.StartingChips, c.StartingChips, domain.BouillotteMinStartingChips, domain.BouillotteMaxStartingChips)
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.BouillotteMinTargetRounds, domain.BouillotteMaxTargetRounds)
	return cfg
}

// BouillotteWebInput はブイヨット Web インプット。
type BouillotteWebInput struct {
	BaseWebInput
	// Action はベッティングアクション ("call"/"raise"/"fold")。bet コマンドで必須。
	Action *string              `json:"action,omitempty"`
	Config *BouillotteWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.BouillotteConfig を構築する。
func (p BouillotteWebInput) ToConfig() domain.BouillotteConfig {
	return configOrDefault(p.Config, (*BouillotteWebConfig).ToConfig, domain.DefaultBouillotteConfig())
}

// BouillotteWebOutputPlayer は 1 プレイヤーの出力。
type BouillotteWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Chips     int              `json:"chips"`
	RoundBet  int              `json:"roundBet"`
	Folded    bool             `json:"folded"`
	Out       bool             `json:"out"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// HandName は公開された手の役名キー ("brelan"/"highcard"、非公開時は空文字)。
	HandName string `json:"handName,omitempty"`
	IsWinner bool   `json:"isWinner"`
}

// BouillotteWebOutputHint はヒント出力。
type BouillotteWebOutputHint struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// BouillotteWebOutputConfig は設定アウトプット。
type BouillotteWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	Ante          int `json:"ante"`
	StartingChips int `json:"startingChips"`
	TargetRounds  int `json:"targetRounds"`
}

// BouillotteWebOutput はブイヨット Web アウトプット。
type BouillotteWebOutput struct {
	Players          []*BouillotteWebOutputPlayer `json:"players"`
	Phase            int                          `json:"phase"`
	RoundNumber      int                          `json:"roundNumber"`
	Pot              int                          `json:"pot"`
	Ante             int                          `json:"ante"`
	Chips            int                          `json:"chips"`
	CurrentBet       int                          `json:"currentBet"`
	RaiseCount       int                          `json:"raiseCount"`
	MaxRaises        int                          `json:"maxRaises"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	DealerIdx        int                          `json:"dealerIdx"`
	Retourne         *WebOutputCard               `json:"retourne"`
	IsHumanTurn      bool                         `json:"isHumanTurn"`
	CanRaise         bool                         `json:"canRaise"`
	WinnerIdx        int                          `json:"winnerIdx"`
	MatchWinnerIdx   int                          `json:"matchWinnerIdx"`
	Result           int                          `json:"result"`
	GameEndFlag      bool                         `json:"gameEndFlag"`
	Hint             *BouillotteWebOutputHint     `json:"hint,omitempty"`
	Config           BouillotteWebOutputConfig    `json:"config"`
	WebOutputBase
}

// BouillotteWebController はブイヨット Web コントローラークラス。
type BouillotteWebController = GameWebController[usecase.BouillotteInteractorIF, BouillotteWebInput, *BouillotteWebOutput]

// NewBouillotteWebController, NewBouillotteWebControllerWithProvider are the
// standard and provider-backed constructors for BouillotteWebController.
var NewBouillotteWebController, NewBouillotteWebControllerWithProvider = webControllerPair[usecase.BouillotteInteractorIF, BouillotteWebInput, *BouillotteWebOutput](
	newBouillotteDefaultOutput, bouillotteDispatch,
)

func newBouillotteDefaultOutput(msg string) *BouillotteWebOutput {
	return &BouillotteWebOutput{
		Players:        make([]*BouillotteWebOutputPlayer, 0),
		WinnerIdx:      -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func bouillotteDispatch(bc *baseController, w http.ResponseWriter, ti usecase.BouillotteInteractorIF, param BouillotteWebInput, newDefault func(string) *BouillotteWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "bet":
		if !requireParam(bc, w, newDefault, param.Action == nil, "param error: action is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Bet(*param.Action))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
