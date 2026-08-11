//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MinibridgeWebInput ミニブリッジWebインプット
type MinibridgeWebInput struct {
	BaseWebInput
	CardIndex *int                 `json:"cardIndex,omitempty"`
	Level     *int                 `json:"level,omitempty"`
	Suit      *int                 `json:"suit,omitempty"`
	Config    *MinibridgeWebConfig `json:"config,omitempty"`
}

// MinibridgeWebConfig ミニブリッジWeb設定
type MinibridgeWebConfig struct {
	Rounds *int `json:"rounds,omitempty"`
}

// MinibridgeWebOutputPlayer ミニブリッジWebアウトプットプレイヤー
type MinibridgeWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Hcp はこの席が公開申告したハイカードポイント。**競りが無いこのゲームの
	// 唯一の公開情報**で、4 席の合計は必ず 40。
	Hcp        int `json:"hcp"`
	Team       int `json:"team"`
	TrickCount int `json:"trickCount"`
}

// MinibridgeWebOutputHint ヒント出力
type MinibridgeWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	// Level / Suit は選ぶべき契約（プレイ中は 0）。
	Level int `json:"level"`
	Suit  int `json:"suit"`
}

// MinibridgeWebOutput ミニブリッジWebアウトプット
type MinibridgeWebOutput struct {
	Players     []*MinibridgeWebOutputPlayer `json:"players"`
	Phase       int                          `json:"phase"`
	RoundNumber int                          `json:"roundNumber"`
	TrickNumber int                          `json:"trickNumber"`
	// ContractLevel は 0 のあいだ未決定、ContractSuit は 0 がノートランプ。
	ContractSuit  int `json:"contractSuit"`
	ContractLevel int `json:"contractLevel"`
	// RequiredTricks は契約の 6 + レベル。契約前は 0。
	RequiredTricks int `json:"requiredTricks"`
	DeclarerIdx    int `json:"declarerIdx"`
	DummyIdx       int `json:"dummyIdx"`
	// DummyHand は契約決定後に公開されるダミーの手札。
	DummyHand []*WebOutputCard `json:"dummyHand"`
	// LastMade / LastTricks は直前のディールの結果。
	LastMade         bool                     `json:"lastMade"`
	LastTricks       int                      `json:"lastTricks"`
	TeamScores       []int                    `json:"teamScores"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	ValidPlays       []int                    `json:"validPlays"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerTeam       int                      `json:"winnerTeam"`
	Hint             *MinibridgeWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config MinibridgeWebOutputConfig `json:"config"`
}

// MinibridgeWebOutputConfig ミニブリッジ設定アウトプット
type MinibridgeWebOutputConfig struct {
	Rounds int `json:"rounds"`
}

// ToConfig builds a MinibridgeConfig from the nested web config, applying bounds checking.
//
// **範囲に収めるだけでは足りない。** ラウンド数が 4 の倍数でないと親が一巡せず、
// ペア同点（実測 8.1%）で宣言側を取る回数が偏ります。エラーにせず丸めます。
func (c *MinibridgeWebConfig) ToConfig() domain.MinibridgeConfig {
	cfg := domain.DefaultMinibridgeConfig()
	rounds := webutil.BoundedIntPtr(c.Rounds,
		domain.MinibridgeRoundsMin, domain.MinibridgeRoundsMax, cfg.Rounds)
	cfg.Rounds = rounds - rounds%domain.MinibridgePlayerCnt
	if cfg.Rounds < domain.MinibridgeRoundsMin {
		cfg.Rounds = domain.MinibridgeRoundsMin
	}
	return cfg
}

// ToConfig builds a MinibridgeConfig from the web input.
func (p MinibridgeWebInput) ToConfig() domain.MinibridgeConfig {
	return configOrDefault(p.Config, (*MinibridgeWebConfig).ToConfig, domain.DefaultMinibridgeConfig())
}

// MinibridgeWebController ミニブリッジWebコントローラークラス
type MinibridgeWebController = GameWebController[usecase.MinibridgeInteractorIF, MinibridgeWebInput, *MinibridgeWebOutput]

// NewMinibridgeWebController and NewMinibridgeWebControllerWithProvider are
// the standard and provider-backed constructors for MinibridgeWebController.
var NewMinibridgeWebController, NewMinibridgeWebControllerWithProvider = webControllerPair[usecase.MinibridgeInteractorIF, MinibridgeWebInput, *MinibridgeWebOutput](
	newMinibridgeDefaultOutput, minibridgeDispatch,
)

func newMinibridgeDefaultOutput(msg string) *MinibridgeWebOutput {
	return &MinibridgeWebOutput{
		Players:       make([]*MinibridgeWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		DummyHand:     make([]*WebOutputCard, 0),
		TeamScores:    make([]int, 0),
		DeclarerIdx:   -1,
		DummyIdx:      -1,
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func minibridgeDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MinibridgeInteractorIF, param MinibridgeWebInput, newDefault func(string) *MinibridgeWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, mi.ResetWithConfig(param.ToConfig()))
	case "c", "contract":
		// **レベルもスートも既定値で埋めない。** 埋めると選んでいない契約を
		// 引き受けてしまう。ノートランプは suit=0 で、これは省略とは違う。
		if !requireParam(bc, w, newDefault, param.Level == nil, "param error: level is required.") {
			return true
		}
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.Contract(*param.Level, *param.Suit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, mi.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, mi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, mi.Hint, mi.ActionLog)
	}
	return true
}
