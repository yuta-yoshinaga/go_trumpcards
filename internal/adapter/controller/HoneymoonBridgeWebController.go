//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HoneymoonBridgeWebInput ハネムーンブリッジWebインプット
type HoneymoonBridgeWebInput struct {
	BaseWebInput
	CardIndex *int                      `json:"cardIndex,omitempty"`
	Level     *int                      `json:"level,omitempty"`
	Suit      *int                      `json:"suit,omitempty"`
	Config    *HoneymoonBridgeWebConfig `json:"config,omitempty"`
}

// HoneymoonBridgeWebConfig ハネムーンブリッジWeb設定
type HoneymoonBridgeWebConfig struct {
	Target *int `json:"target,omitempty"`
}

// HoneymoonBridgeWebOutputPlayer ハネムーンブリッジWebアウトプットプレイヤー
type HoneymoonBridgeWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// BidLevel / BidSuit はこの席の直近の宣言 (レベル 0 はパス)。
	BidLevel   int `json:"bidLevel"`
	BidSuit    int `json:"bidSuit"`
	TrickCount int `json:"trickCount"`
	Score      int `json:"score"`
}

// HoneymoonBridgeWebOutputHint ヒント出力
type HoneymoonBridgeWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	// Level / Suit は競りで勧める契約 (それ以外は 0)。
	Level int `json:"level"`
	Suit  int `json:"suit"`
}

// HoneymoonBridgeWebOutput ハネムーンブリッジWebアウトプット
type HoneymoonBridgeWebOutput struct {
	Players     []*HoneymoonBridgeWebOutputPlayer `json:"players"`
	Phase       int                               `json:"phase"`
	RoundNumber int                               `json:"roundNumber"`
	TrickNumber int                               `json:"trickNumber"`
	// StockSize は引き合いフェーズの山札の残り。競り以降は 0。
	StockSize int `json:"stockSize"`
	// TrumpSuit は 0 のあいだノートランプ。落札するまでも 0。
	TrumpSuit     int `json:"trumpSuit"`
	DeclarerIdx   int `json:"declarerIdx"`
	ContractLevel int `json:"contractLevel"`
	// RequiredTricks は契約の 6 + レベル。契約前は 0。
	RequiredTricks int `json:"requiredTricks"`
	// MinBidLevel / MinBidSuit は次に通る最小の宣言。**サーバが必ず拒否する値を
	// クライアントに出させないためにワイヤへ載せる。** 競り以外では 0。
	MinBidLevel int `json:"minBidLevel"`
	MinBidSuit  int `json:"minBidSuit"`
	// LastMade / LastTricks は直前のディールの結果。
	LastMade         bool                          `json:"lastMade"`
	LastTricks       int                           `json:"lastTricks"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                           `json:"leadPlayerIdx"`
	DealerIdx        int                           `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard         `json:"currentTrick"`
	ValidPlays       []int                         `json:"validPlays"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerIdx        int                           `json:"winnerIdx"`
	Hint             *HoneymoonBridgeWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config HoneymoonBridgeWebOutputConfig `json:"config"`
}

// HoneymoonBridgeWebOutputConfig ハネムーンブリッジ設定アウトプット
type HoneymoonBridgeWebOutputConfig struct {
	Target int `json:"target"`
}

// ToConfig builds a HoneymoonBridgeConfig from the nested web config, applying bounds checking.
func (c *HoneymoonBridgeWebConfig) ToConfig() domain.HoneymoonBridgeConfig {
	cfg := domain.DefaultHoneymoonBridgeConfig()
	cfg.Target = webutil.BoundedIntPtr(c.Target,
		domain.HoneymoonBridgeTargetMin, domain.HoneymoonBridgeTargetMax, cfg.Target)
	return cfg
}

// ToConfig builds a HoneymoonBridgeConfig from the web input.
func (p HoneymoonBridgeWebInput) ToConfig() domain.HoneymoonBridgeConfig {
	return configOrDefault(p.Config, (*HoneymoonBridgeWebConfig).ToConfig, domain.DefaultHoneymoonBridgeConfig())
}

// HoneymoonBridgeWebController ハネムーンブリッジWebコントローラークラス
type HoneymoonBridgeWebController = GameWebController[usecase.HoneymoonBridgeInteractorIF, HoneymoonBridgeWebInput, *HoneymoonBridgeWebOutput]

// NewHoneymoonBridgeWebController and NewHoneymoonBridgeWebControllerWithProvider are
// the standard and provider-backed constructors for HoneymoonBridgeWebController.
var NewHoneymoonBridgeWebController, NewHoneymoonBridgeWebControllerWithProvider = webControllerPair[usecase.HoneymoonBridgeInteractorIF, HoneymoonBridgeWebInput, *HoneymoonBridgeWebOutput](
	newHoneymoonBridgeDefaultOutput, honeymoonBridgeDispatch,
)

func newHoneymoonBridgeDefaultOutput(msg string) *HoneymoonBridgeWebOutput {
	return &HoneymoonBridgeWebOutput{
		Players:       make([]*HoneymoonBridgeWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		DeclarerIdx:   -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func honeymoonBridgeDispatch(bc *baseController, w http.ResponseWriter, hi usecase.HoneymoonBridgeInteractorIF, param HoneymoonBridgeWebInput, newDefault func(string) *HoneymoonBridgeWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, hi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		// **レベルもスートも既定値で埋めない。** 埋めると宣言していない契約を
		// 落札してしまう。ノートランプは suit=0 で、これは省略とは違う。
		if !requireParam(bc, w, newDefault, param.Level == nil, "param error: level is required.") {
			return true
		}
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, hi.Bid(*param.Level, *param.Suit))
	case "pass":
		bc.writePresenterResponse(w, hi.Pass())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, hi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, hi.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, hi.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, hi.Hint, hi.ActionLog)
	}
	return true
}
