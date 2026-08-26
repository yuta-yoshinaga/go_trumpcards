//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LingerLongerWebInput リンガーロンガーWebインプット
type LingerLongerWebInput struct {
	BaseWebInput
	CardIndex *int                   `json:"cardIndex,omitempty"`
	Config    *LingerLongerWebConfig `json:"config,omitempty"`
}

// LingerLongerWebConfig リンガーロンガーWeb設定
type LingerLongerWebConfig struct {
	PlayerCnt *int `json:"playerCnt,omitempty"`
}

// LingerLongerWebOutputPlayer リンガーロンガーWebアウトプットプレイヤー
type LingerLongerWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// TricksWon は取ったトリック数。**得点ではなく、補充できた回数。**
	TricksWon int `json:"tricksWon"`
	// EliminatedAt は脱落した順番（0 = まだ在席）。
	EliminatedAt int `json:"eliminatedAt"`
}

// LingerLongerWebOutputHint ヒント出力
type LingerLongerWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// LingerLongerWebOutput リンガーロンガーWebアウトプット
type LingerLongerWebOutput struct {
	Players    []*LingerLongerWebOutputPlayer `json:"players"`
	Phase      int                            `json:"phase"`
	ValidPlays []int                          `json:"validPlays"`
	// StockSize は山札の残り。**0 になると誰も補充できず、脱落が一気に進みます。**
	StockSize        int                   `json:"stockSize"`
	CurrentTrick     []*WebOutputTrickCard `json:"currentTrick"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                   `json:"leadPlayerIdx"`
	TrickNumber      int                   `json:"trickNumber"`
	LastDrawIdx      int                   `json:"lastDrawIdx"`
	EliminatedCnt    int                   `json:"eliminatedCnt"`
	Discarded        int                   `json:"discarded"`
	GameEndFlag      bool                  `json:"gameEndFlag"`
	WinnerIdx        int                   `json:"winnerIdx"`
	// WinReason は決着の理由のロケール非依存キー ("lasted" / "lastTrick" /
	// "giveUp")。決着前は空。ページは winnerIdx だけを見て「最後まで持ち続けた」
	// と書いていたが、全員が同時に出し切った局ではそれが事実に反する (#5765)。
	WinReason string                     `json:"winReason"`
	Hint      *LingerLongerWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config LingerLongerWebOutputConfig `json:"config"`
}

// LingerLongerWebOutputConfig リンガーロンガー設定アウトプット
type LingerLongerWebOutputConfig struct {
	PlayerCnt int `json:"playerCnt"`
}

// ToConfig builds a LingerLongerConfig from the nested web config, applying bounds checking.
func (c *LingerLongerWebConfig) ToConfig() domain.LingerLongerConfig {
	cfg := domain.DefaultLingerLongerConfig()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.LingerLongerPlayerCntMin, domain.LingerLongerPlayerCntMax, cfg.PlayerCnt)
	return cfg
}

// ToConfig builds a LingerLongerConfig from the web input.
func (p LingerLongerWebInput) ToConfig() domain.LingerLongerConfig {
	return configOrDefault(p.Config, (*LingerLongerWebConfig).ToConfig, domain.DefaultLingerLongerConfig())
}

// LingerLongerWebController リンガーロンガーWebコントローラークラス
type LingerLongerWebController = GameWebController[usecase.LingerLongerInteractorIF, LingerLongerWebInput, *LingerLongerWebOutput]

// NewLingerLongerWebController and NewLingerLongerWebControllerWithProvider are
// the standard and provider-backed constructors for LingerLongerWebController.
var NewLingerLongerWebController, NewLingerLongerWebControllerWithProvider = webControllerPair[usecase.LingerLongerInteractorIF, LingerLongerWebInput, *LingerLongerWebOutput](
	newLingerLongerDefaultOutput, lingerLongerDispatch,
)

func newLingerLongerDefaultOutput(msg string) *LingerLongerWebOutput {
	return &LingerLongerWebOutput{
		Players:       make([]*LingerLongerWebOutputPlayer, 0),
		ValidPlays:    make([]int, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		LastDrawIdx:   -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func lingerLongerDispatch(bc *baseController, w http.ResponseWriter, li usecase.LingerLongerInteractorIF, param LingerLongerWebInput, newDefault func(string) *LingerLongerWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, li.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, li.Play(*param.CardIndex))
	case "g", "giveup":
		bc.writePresenterResponse(w, li.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, li.Hint, li.ActionLog)
	}
	return true
}
