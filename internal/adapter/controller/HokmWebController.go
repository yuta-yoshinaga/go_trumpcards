//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HokmWebInput ホクムWebインプット
type HokmWebInput struct {
	BaseWebInput
	CardIndex *int           `json:"cardIndex,omitempty"`
	Suit      *int           `json:"suit,omitempty"`
	Config    *HokmWebConfig `json:"config,omitempty"`
}

// HokmWebConfig ホクムWeb設定
type HokmWebConfig struct {
	Target *int `json:"target,omitempty"`
}

// HokmWebOutputPlayer ホクムWebアウトプットプレイヤー
type HokmWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Team      int              `json:"team"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// IsHakem は切り札を宣言する親か。勝っているあいだ交代しない。
	IsHakem    bool `json:"isHakem"`
	TrickCount int  `json:"trickCount"`
}

// HokmWebOutputHint ヒント出力
type HokmWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	Suit      int    `json:"suit"`
}

// HokmWebOutput ホクムWebアウトプット
type HokmWebOutput struct {
	Players     []*HokmWebOutputPlayer `json:"players"`
	Phase       int                    `json:"phase"`
	HandNumber  int                    `json:"handNumber"`
	TrickNumber int                    `json:"trickNumber"`
	TrumpSuit   int                    `json:"trumpSuit"`
	HakemIdx    int                    `json:"hakemIdx"`
	// Scores はハンド勝ち点、TeamTricks はこのハンドの獲得トリック数。
	// **7 で即終了なので、進捗はトリック数のほうに出る。**
	Scores      []int `json:"scores"`
	TeamTricks  []int `json:"teamTricks"`
	TricksToWin int   `json:"tricksToWin"`
	// LastHandKot は直前のハンドが Kot だったか、LastHandWinner はその勝者 (-1: 無し)。
	LastHandKot bool `json:"lastHandKot"`
	// LastHandHakemChanged は直前のハンドで親が交代したか。
	//
	// **次に自分が切り札を選べるかを左右する。**これまでは次ハンドが始まって
	// 親バッジが動くのを見るまで分からなかった (#5753)。
	LastHandHakemChanged bool                  `json:"lastHandHakemChanged"`
	LastHandWinner       int                   `json:"lastHandWinner"`
	CurrentPlayerIdx     int                   `json:"currentPlayerIdx"`
	LeadPlayerIdx        int                   `json:"leadPlayerIdx"`
	CurrentTrick         []*WebOutputTrickCard `json:"currentTrick"`
	ValidPlays           []int                 `json:"validPlays"`
	GameEndFlag          bool                  `json:"gameEndFlag"`
	WinnerTeam           int                   `json:"winnerTeam"`
	Hint                 *HokmWebOutputHint    `json:"hint,omitempty"`
	WebOutputBase
	Config HokmWebOutputConfig `json:"config"`
}

// HokmWebOutputConfig ホクム設定アウトプット
type HokmWebOutputConfig struct {
	Target int `json:"target"`
}

// ToConfig builds a HokmConfig from the nested web config, applying bounds checking.
func (c *HokmWebConfig) ToConfig() domain.HokmConfig {
	cfg := domain.DefaultHokmConfig()
	cfg.Target = webutil.BoundedIntPtr(c.Target,
		domain.HokmTargetMin, domain.HokmTargetMax, cfg.Target)
	return cfg
}

// ToConfig builds a HokmConfig from the web input.
func (p HokmWebInput) ToConfig() domain.HokmConfig {
	return configOrDefault(p.Config, (*HokmWebConfig).ToConfig, domain.DefaultHokmConfig())
}

// HokmWebController ホクムWebコントローラークラス
type HokmWebController = GameWebController[usecase.HokmInteractorIF, HokmWebInput, *HokmWebOutput]

// NewHokmWebController and NewHokmWebControllerWithProvider are
// the standard and provider-backed constructors for HokmWebController.
var NewHokmWebController, NewHokmWebControllerWithProvider = webControllerPair[usecase.HokmInteractorIF, HokmWebInput, *HokmWebOutput](
	newHokmDefaultOutput, hokmDispatch,
)

func newHokmDefaultOutput(msg string) *HokmWebOutput {
	return &HokmWebOutput{
		Players:        make([]*HokmWebOutputPlayer, 0),
		CurrentTrick:   make([]*WebOutputTrickCard, 0),
		ValidPlays:     make([]int, 0),
		Scores:         make([]int, 0),
		TeamTricks:     make([]int, 0),
		TricksToWin:    domain.HokmTricksToWin,
		LastHandWinner: -1,
		WinnerTeam:     -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func hokmDispatch(bc *baseController, w http.ResponseWriter, hi usecase.HokmInteractorIF, param HokmWebInput, newDefault func(string) *HokmWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, hi.ResetWithConfig(param.ToConfig()))
	case "t", "trump":
		// **切り札は既定値で埋めない。** 埋めると親が選んでいないスートが
		// そのハンドの切り札になる。
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, hi.DeclareTrump(*param.Suit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, hi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, hi.NextHand())
	case "g", "giveup":
		bc.writePresenterResponse(w, hi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, hi.Hint, hi.ActionLog)
	}
	return true
}
