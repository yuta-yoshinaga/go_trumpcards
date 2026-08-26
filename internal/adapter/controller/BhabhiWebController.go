//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BhabhiWebInput バービーWebインプット
type BhabhiWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *BhabhiWebConfig `json:"config,omitempty"`
}

// BhabhiWebConfig バービーWeb設定
type BhabhiWebConfig struct {
	PlayerCnt *int `json:"playerCnt,omitempty"`
}

// BhabhiWebOutputPlayer バービーWebアウトプットプレイヤー
type BhabhiWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Rank は上がった順位 (-1: まだ手札が残っている)。**強さの順ではない。**
	Rank int `json:"rank"`
	// Pickups は場札を引き取った回数。
	Pickups int `json:"pickups"`
}

// BhabhiWebOutputHint ヒント出力
type BhabhiWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// BhabhiWebOutput バービーWebアウトプット
type BhabhiWebOutput struct {
	Players     []*BhabhiWebOutputPlayer `json:"players"`
	Phase       int                      `json:"phase"`
	TrickNumber int                      `json:"trickNumber"`
	// LeadSuit は 0 のあいだトリック未開始。フォロー義務はこのスートに掛かる。
	LeadSuit int `json:"leadSuit"`
	// Pile は場に出ている札。**フォローできない人がこれを全部引き取る。**
	Pile []*WebOutputTrickCard `json:"pile"`
	// LastPickupIdx / LastPickupSize は直前の引き取り (-1 / 0)。
	LastPickupIdx    int   `json:"lastPickupIdx"`
	LastPickupSize   int   `json:"lastPickupSize"`
	CurrentPlayerIdx int   `json:"currentPlayerIdx"`
	LeadPlayerIdx    int   `json:"leadPlayerIdx"`
	ValidPlays       []int `json:"validPlays"`
	AliveCount       int   `json:"aliveCount"`
	GameEndFlag      bool  `json:"gameEndFlag"`
	// BhabhiIdx は敗者 (-1: 未確定)。**勝者ではなく敗者を決めるゲーム。**
	BhabhiIdx int `json:"bhabhiIdx"`
	// Stalemate は膠着で打ち切ったかどうか。
	Stalemate bool `json:"stalemate"`
	// StalemateTricks は膠着と判定するトリック数。
	StalemateTricks int                  `json:"stalemateTricks"`
	Hint            *BhabhiWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config BhabhiWebOutputConfig `json:"config"`
}

// BhabhiWebOutputConfig バービー設定アウトプット
type BhabhiWebOutputConfig struct {
	PlayerCnt int `json:"playerCnt"`
}

// ToConfig builds a BhabhiConfig from the nested web config, applying bounds checking.
func (c *BhabhiWebConfig) ToConfig() domain.BhabhiConfig {
	cfg := domain.DefaultBhabhiConfig()
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.BhabhiMinPlayers, domain.BhabhiMaxPlayers, cfg.PlayerCnt)
	return cfg
}

// ToConfig builds a BhabhiConfig from the web input.
func (p BhabhiWebInput) ToConfig() domain.BhabhiConfig {
	return configOrDefault(p.Config, (*BhabhiWebConfig).ToConfig, domain.DefaultBhabhiConfig())
}

// BhabhiWebController バービーWebコントローラークラス
type BhabhiWebController = GameWebController[usecase.BhabhiInteractorIF, BhabhiWebInput, *BhabhiWebOutput]

// NewBhabhiWebController and NewBhabhiWebControllerWithProvider are
// the standard and provider-backed constructors for BhabhiWebController.
var NewBhabhiWebController, NewBhabhiWebControllerWithProvider = webControllerPair[usecase.BhabhiInteractorIF, BhabhiWebInput, *BhabhiWebOutput](
	newBhabhiDefaultOutput, bhabhiDispatch,
)

func newBhabhiDefaultOutput(msg string) *BhabhiWebOutput {
	return &BhabhiWebOutput{
		Players:         make([]*BhabhiWebOutputPlayer, 0),
		Pile:            make([]*WebOutputTrickCard, 0),
		ValidPlays:      make([]int, 0),
		LastPickupIdx:   -1,
		BhabhiIdx:       -1,
		StalemateTricks: domain.BhabhiStalemateTricks,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

// bhabhiDispatch はコマンドを振り分ける。
//
// **ハンドの区切りが無いので next は無い。** 配り切りの 1 ゲームで最後の 1 人が
// 決まるまで続きます。
func bhabhiDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BhabhiInteractorIF, param BhabhiWebInput, newDefault func(string) *BhabhiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "g", "giveup":
		bc.writePresenterResponse(w, bi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
