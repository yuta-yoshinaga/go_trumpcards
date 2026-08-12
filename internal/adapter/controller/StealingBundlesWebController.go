//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// StealingBundlesWebInput スティーリングバンドルWebインプット
type StealingBundlesWebInput struct {
	BaseWebInput
	CardIndex *int                      `json:"cardIndex,omitempty"`
	VictimIdx *int                      `json:"victimIdx,omitempty"`
	Config    *StealingBundlesWebConfig `json:"config,omitempty"`
}

// StealingBundlesWebConfig スティーリングバンドルWeb設定
type StealingBundlesWebConfig struct {
	PlayerCnt *int `json:"playerCnt,omitempty"`
}

// StealingBundlesWebOutputPlayer スティーリングバンドルWebアウトプットプレイヤー
type StealingBundlesWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// BundleSize は獲得した束の枚数。**そのまま得点です。**
	BundleSize int `json:"bundleSize"`
	// BundleTop は束の一番上。**ここが狙われる場所**なので、伏せずに出します。
	BundleTop *WebOutputCard `json:"bundleTop,omitempty"`
}

// StealingBundlesWebOutputHint ヒント出力
type StealingBundlesWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	VictimIdx int    `json:"victimIdx"`
	Reason    string `json:"reason"`
}

// StealingBundlesWebOutput スティーリングバンドルWebアウトプット
type StealingBundlesWebOutput struct {
	Players    []*StealingBundlesWebOutputPlayer `json:"players"`
	Phase      int                               `json:"phase"`
	TableCards []*WebOutputCard                  `json:"tableCards"`
	// TableMatches は手札の位置ごとに取れる場札の位置。
	TableMatches map[string][]int `json:"tableMatches"`
	// StealTargets は手札の位置ごとに奪える相手の席。
	StealTargets map[string][]int `json:"stealTargets"`
	// CanCapture は取れる手があるか。**偽のときだけ場に置けます。**
	CanCapture       bool                          `json:"canCapture"`
	DeckRemaining    int                           `json:"deckRemaining"`
	LastCaptureIdx   int                           `json:"lastCaptureIdx"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	TurnNumber       int                           `json:"turnNumber"`
	PacksDealt       int                           `json:"packsDealt"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerIdx        int                           `json:"winnerIdx"`
	Hint             *StealingBundlesWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config StealingBundlesWebOutputConfig `json:"config"`
}

// StealingBundlesWebOutputConfig スティーリングバンドル設定アウトプット
type StealingBundlesWebOutputConfig struct {
	PlayerCnt int `json:"playerCnt"`
}

// ToConfig builds a StealingBundlesConfig from the nested web config, applying bounds checking.
func (c *StealingBundlesWebConfig) ToConfig() domain.StealingBundlesConfig {
	cfg := domain.DefaultStealingBundlesConfig()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.StealingBundlesPlayerCntMin, domain.StealingBundlesPlayerCntMax, cfg.PlayerCnt)
	return cfg
}

// ToConfig builds a StealingBundlesConfig from the web input.
func (p StealingBundlesWebInput) ToConfig() domain.StealingBundlesConfig {
	return configOrDefault(p.Config, (*StealingBundlesWebConfig).ToConfig, domain.DefaultStealingBundlesConfig())
}

// StealingBundlesWebController スティーリングバンドルWebコントローラークラス
type StealingBundlesWebController = GameWebController[usecase.StealingBundlesInteractorIF, StealingBundlesWebInput, *StealingBundlesWebOutput]

// NewStealingBundlesWebController and NewStealingBundlesWebControllerWithProvider are
// the standard and provider-backed constructors for StealingBundlesWebController.
var NewStealingBundlesWebController, NewStealingBundlesWebControllerWithProvider = webControllerPair[usecase.StealingBundlesInteractorIF, StealingBundlesWebInput, *StealingBundlesWebOutput](
	newStealingBundlesDefaultOutput, stealingBundlesDispatch,
)

func newStealingBundlesDefaultOutput(msg string) *StealingBundlesWebOutput {
	return &StealingBundlesWebOutput{
		Players:        make([]*StealingBundlesWebOutputPlayer, 0),
		TableCards:     make([]*WebOutputCard, 0),
		TableMatches:   map[string][]int{},
		StealTargets:   map[string][]int{},
		LastCaptureIdx: -1,
		WinnerIdx:      -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func stealingBundlesDispatch(bc *baseController, w http.ResponseWriter, si usecase.StealingBundlesInteractorIF, param StealingBundlesWebInput, newDefault func(string) *StealingBundlesWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "t", "take":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Take(*param.CardIndex))
	case "s", "steal":
		// **略奪は相手を指名します。** どちらか欠けても盤面は動かしません。
		if !requireParam(bc, w, newDefault, param.CardIndex == nil || param.VictimIdx == nil,
			"param error: cardIndex and victimIdx are required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Steal(*param.CardIndex, *param.VictimIdx))
	case "d", "trail":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Trail(*param.CardIndex))
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
