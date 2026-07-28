//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PrimeroWebConfig はプリメロ (Primero) の Web 設定。
type PrimeroWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	Ante          *int `json:"ante,omitempty"`
	StartingChips *int `json:"startingChips,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// ToConfig は PrimeroWebConfig を domain.PrimeroConfig に変換する (境界チェック付き)。
func (c *PrimeroWebConfig) ToConfig() domain.PrimeroConfig {
	cfg := domain.DefaultPrimeroConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.PrimeroMinPlayerCount, domain.PrimeroMaxPlayerCount)
	webutil.ApplyBoundedInt(&cfg.Ante, c.Ante, domain.PrimeroMinAnte, domain.PrimeroMaxAnte)
	webutil.ApplyBoundedInt(&cfg.StartingChips, c.StartingChips, domain.PrimeroMinStartingChips, domain.PrimeroMaxStartingChips)
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.PrimeroMinTargetRounds, domain.PrimeroMaxTargetRounds)
	return cfg
}

// PrimeroWebInput はプリメロ Web インプット。
type PrimeroWebInput struct {
	BaseWebInput
	// Action はベッティングアクション ("call"/"raise"/"fold")。bet コマンドで必須。
	Action *string           `json:"action,omitempty"`
	Config *PrimeroWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.PrimeroConfig を構築する。
func (p PrimeroWebInput) ToConfig() domain.PrimeroConfig {
	return configOrDefault(p.Config, (*PrimeroWebConfig).ToConfig, domain.DefaultPrimeroConfig())
}

// PrimeroWebOutputPlayer は 1 プレイヤーの出力。
type PrimeroWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Chips     int              `json:"chips"`
	RoundBet  int              `json:"roundBet"`
	Folded    bool             `json:"folded"`
	Out       bool             `json:"out"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// HandName は公開された手の役名キー ("fluxus"/"supremus"/"primero"/"numerus"、非公開時は空文字)。
	HandName string `json:"handName,omitempty"`
	IsWinner bool   `json:"isWinner"`
}

// PrimeroWebOutputHint はヒント出力。
type PrimeroWebOutputHint struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// PrimeroWebOutputConfig は設定アウトプット。
type PrimeroWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	Ante          int `json:"ante"`
	StartingChips int `json:"startingChips"`
	TargetRounds  int `json:"targetRounds"`
}

// PrimeroWebOutput はプリメロ Web アウトプット。
type PrimeroWebOutput struct {
	Players          []*PrimeroWebOutputPlayer `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	Pot              int                       `json:"pot"`
	Ante             int                       `json:"ante"`
	Chips            int                       `json:"chips"`
	CurrentBet       int                       `json:"currentBet"`
	RaiseCount       int                       `json:"raiseCount"`
	MaxRaises        int                       `json:"maxRaises"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	DealerIdx        int                       `json:"dealerIdx"`
	IsHumanTurn      bool                      `json:"isHumanTurn"`
	CanRaise         bool                      `json:"canRaise"`
	WinnerIdx        int                       `json:"winnerIdx"`
	MatchWinnerIdx   int                       `json:"matchWinnerIdx"`
	Result           int                       `json:"result"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	Hint             *PrimeroWebOutputHint     `json:"hint,omitempty"`
	Config           PrimeroWebOutputConfig    `json:"config"`
	WebOutputBase
}

// PrimeroWebController はプリメロ Web コントローラークラス。
type PrimeroWebController = GameWebController[usecase.PrimeroInteractorIF, PrimeroWebInput, *PrimeroWebOutput]

// NewPrimeroWebController, NewPrimeroWebControllerWithProvider are the
// standard and provider-backed constructors for PrimeroWebController.
var NewPrimeroWebController, NewPrimeroWebControllerWithProvider = webControllerPair[usecase.PrimeroInteractorIF, PrimeroWebInput, *PrimeroWebOutput](
	newPrimeroDefaultOutput, primeroDispatch,
)

func newPrimeroDefaultOutput(msg string) *PrimeroWebOutput {
	return &PrimeroWebOutput{
		Players:        make([]*PrimeroWebOutputPlayer, 0),
		WinnerIdx:      -1,
		MatchWinnerIdx: -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func primeroDispatch(bc *baseController, w http.ResponseWriter, ti usecase.PrimeroInteractorIF, param PrimeroWebInput, newDefault func(string) *PrimeroWebOutput) bool {
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
