//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MendikotWebInput メンディコットWebインプット
type MendikotWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *MendikotWebConfig `json:"config,omitempty"`
}

// MendikotWebConfig メンディコットWeb設定
type MendikotWebConfig struct {
	Target *int `json:"target,omitempty"`
}

// MendikotWebOutputPlayer メンディコットWebアウトプットプレイヤー
type MendikotWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	Team      int              `json:"team"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Tens はこのハンドで取った 10 の枚数。**勝敗そのもの。**
	Tens       int `json:"tens"`
	TrickCount int `json:"trickCount"`
}

// MendikotWebOutputHint ヒント出力
type MendikotWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// MendikotWebOutput メンディコットWebアウトプット
type MendikotWebOutput struct {
	Players     []*MendikotWebOutputPlayer `json:"players"`
	Phase       int                        `json:"phase"`
	HandNumber  int                        `json:"handNumber"`
	TrickNumber int                        `json:"trickNumber"`
	// TrumpSuit は 0 のあいだ切り札なし。TrumpChooserIdx は決めた席 (-1: 未決定)。
	TrumpSuit       int `json:"trumpSuit"`
	TrumpChooserIdx int `json:"trumpChooserIdx"`
	// TeamTens は 10 の獲得枚数、TeamTricks はトリック数。**勝敗はこの2つで決まる。**
	TeamTens   []int `json:"teamTens"`
	TeamTricks []int `json:"teamTricks"`
	TensInDeck int   `json:"tensInDeck"`
	Scores     []int `json:"scores"`
	// LastHandWinner / LastHandKind は直前のハンドの結末 (-1 / "")。
	LastHandWinner   int                    `json:"lastHandWinner"`
	LastHandKind     string                 `json:"lastHandKind"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                    `json:"leadPlayerIdx"`
	DealerIdx        int                    `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard  `json:"currentTrick"`
	ValidPlays       []int                  `json:"validPlays"`
	GameEndFlag      bool                   `json:"gameEndFlag"`
	WinnerTeam       int                    `json:"winnerTeam"`
	Hint             *MendikotWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config MendikotWebOutputConfig `json:"config"`
}

// MendikotWebOutputConfig メンディコット設定アウトプット
type MendikotWebOutputConfig struct {
	Target int `json:"target"`
}

// ToConfig builds a MendikotConfig from the nested web config, applying bounds checking.
func (c *MendikotWebConfig) ToConfig() domain.MendikotConfig {
	cfg := domain.DefaultMendikotConfig()
	cfg.Target = webutil.BoundedIntPtr(c.Target,
		domain.MendikotTargetMin, domain.MendikotTargetMax, cfg.Target)
	return cfg
}

// ToConfig builds a MendikotConfig from the web input.
func (p MendikotWebInput) ToConfig() domain.MendikotConfig {
	return configOrDefault(p.Config, (*MendikotWebConfig).ToConfig, domain.DefaultMendikotConfig())
}

// MendikotWebController メンディコットWebコントローラークラス
type MendikotWebController = GameWebController[usecase.MendikotInteractorIF, MendikotWebInput, *MendikotWebOutput]

// NewMendikotWebController and NewMendikotWebControllerWithProvider are
// the standard and provider-backed constructors for MendikotWebController.
var NewMendikotWebController, NewMendikotWebControllerWithProvider = webControllerPair[usecase.MendikotInteractorIF, MendikotWebInput, *MendikotWebOutput](
	newMendikotDefaultOutput, mendikotDispatch,
)

func newMendikotDefaultOutput(msg string) *MendikotWebOutput {
	return &MendikotWebOutput{
		Players:         make([]*MendikotWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		ValidPlays:      make([]int, 0),
		Scores:          make([]int, 0),
		TeamTens:        make([]int, 0),
		TeamTricks:      make([]int, 0),
		TensInDeck:      domain.MendikotTensInDeck,
		TrumpChooserIdx: -1,
		LastHandWinner:  -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func mendikotDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MendikotInteractorIF, param MendikotWebInput, newDefault func(string) *MendikotWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, mi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, mi.NextHand())
	case "g", "giveup":
		bc.writePresenterResponse(w, mi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, mi.Hint, mi.ActionLog)
	}
	return true
}
