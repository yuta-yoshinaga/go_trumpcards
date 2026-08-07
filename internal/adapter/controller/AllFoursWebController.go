package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// AllFoursWebInput All Fours Webインプット
type AllFoursWebInput struct {
	BaseWebInput
	Beg       *bool              `json:"beg,omitempty"`
	Run       *bool              `json:"run,omitempty"`
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *AllFoursWebConfig `json:"config,omitempty"`
}

// AllFoursWebConfig All Fours Web設定
type AllFoursWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// AllFoursWebOutputPlayer All Fours Webアウトプットプレイヤー
type AllFoursWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// AllFoursWebOutputHint ヒント出力
type AllFoursWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Beg       *bool  `json:"beg,omitempty"`
	Run       *bool  `json:"run,omitempty"`
	Reason    string `json:"reason"`
}

// AllFoursWebOutputBreakdownAward は High / Low 項目の獲得者と該当トランプ札を表す。
// WinnerIdx が -1 のときは付与なし (トランプ未確定など)。
type AllFoursWebOutputBreakdownAward struct {
	WinnerIdx int            `json:"winnerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// AllFoursWebOutputBreakdownJack は Jack 項目 (J トランプの捕獲) の獲得者を表す。
// WinnerIdx が -1 のときは、そのディールで J トランプが場に出なかったことを示す。
type AllFoursWebOutputBreakdownJack struct {
	WinnerIdx int `json:"winnerIdx"`
}

// AllFoursWebOutputBreakdownGame は Game 項目 (カードピップ合計) の内訳を表す。
// WinnerIdx が -1 のときは付与なし (同点または全員 0 点)。Points はプレイヤーごとのピップ合計。
type AllFoursWebOutputBreakdownGame struct {
	WinnerIdx int   `json:"winnerIdx"`
	Points    []int `json:"points"`
}

// AllFoursWebOutputRoundBreakdown はラウンド確定時の High / Low / Jack / Game 得点内訳。
// ROUND_END / GAME_END フェーズでのみ出力される。
type AllFoursWebOutputRoundBreakdown struct {
	High AllFoursWebOutputBreakdownAward `json:"high"`
	Low  AllFoursWebOutputBreakdownAward `json:"low"`
	Jack AllFoursWebOutputBreakdownJack  `json:"jack"`
	Game AllFoursWebOutputBreakdownGame  `json:"game"`
	// Provisional はラウンド途中の暫定値であることを表す (#4771)。
	//
	// **途中経過を確定値として見せてはいけない。**High も Low も「そのラウンドで
	// 出たトランプの中で」決まるので、まだ配られていない札で引っくり返る。
	Provisional bool `json:"provisional"`
}

// AllFoursWebOutput All Fours Webアウトプット
type AllFoursWebOutput struct {
	Players          []*AllFoursWebOutputPlayer       `json:"players"`
	Phase            int                              `json:"phase"`
	RoundNumber      int                              `json:"roundNumber"`
	TrickNumber      int                              `json:"trickNumber"`
	DealerIdx        int                              `json:"dealerIdx"`
	NonDealerIdx     int                              `json:"nonDealerIdx"`
	CurrentPlayerIdx int                              `json:"currentPlayerIdx"`
	TrumpSuit        int                              `json:"trumpSuit"`
	TurnUp           *WebOutputCard                   `json:"turnUp"`
	RunCount         int                              `json:"runCount"`
	CurrentTrick     []*WebOutputTrickCard            `json:"currentTrick"`
	GameEndFlag      bool                             `json:"gameEndFlag"`
	WinnerIdx        int                              `json:"winnerIdx"`
	LeadPlayerIdx    int                              `json:"leadPlayerIdx"`
	ValidPlayIndices []int                            `json:"validPlayIndices,omitempty"`
	RoundBreakdown   *AllFoursWebOutputRoundBreakdown `json:"roundBreakdown,omitempty"`
	Hint             *AllFoursWebOutputHint           `json:"hint,omitempty"`
	WebOutputBase
	Config AllFoursWebOutputConfig `json:"config"`
}

// AllFoursWebOutputConfig All Fours 設定アウトプット
type AllFoursWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds an AllFoursConfig from the nested web config, applying bounds checking.
func (c *AllFoursWebConfig) ToConfig() domain.AllFoursConfig {
	cfg := domain.DefaultAllFoursConfig()
	cfg.CpuDifficulty = domain.AllFoursCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.AllFoursCpuDifficultyEasy), int(domain.AllFoursCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds an AllFoursConfig from the web input.
func (in AllFoursWebInput) ToConfig() domain.AllFoursConfig {
	return configOrDefault(in.Config, (*AllFoursWebConfig).ToConfig, domain.DefaultAllFoursConfig())
}

// AllFoursWebController All Fours Webコントローラー型
type AllFoursWebController = GameWebController[usecase.AllFoursInteractorIF, AllFoursWebInput, *AllFoursWebOutput]

// NewAllFoursWebController and NewAllFoursWebControllerWithProvider are
// the standard and provider-backed constructors for AllFoursWebController.
var NewAllFoursWebController, NewAllFoursWebControllerWithProvider = webControllerPair[usecase.AllFoursInteractorIF, AllFoursWebInput, *AllFoursWebOutput](
	newAllFoursDefaultOutput, allFoursDispatch,
)

func newAllFoursDefaultOutput(msg string) *AllFoursWebOutput {
	return &AllFoursWebOutput{
		Players:       make([]*AllFoursWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func allFoursDispatch(bc *baseController, w http.ResponseWriter, ai usecase.AllFoursInteractorIF, param AllFoursWebInput, newDefault func(string) *AllFoursWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ai.ResetWithConfig(param.ToConfig()))
	case "beg":
		if !requireParam(bc, w, newDefault, param.Beg == nil, "param error: beg is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.Beg(*param.Beg))
	case "respond":
		if !requireParam(bc, w, newDefault, param.Run == nil, "param error: run is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.RespondBeg(*param.Run))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ai.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ai.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ai.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ai.Hint, ai.ActionLog)
	}
	return true
}
