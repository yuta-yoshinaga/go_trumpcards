//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SergeantMajorWebInput サージェントメジャーWebインプット
type SergeantMajorWebInput struct {
	BaseWebInput
	CardIndex *int                    `json:"cardIndex,omitempty"`
	Discards  []int                   `json:"discards,omitempty"`
	Suit      *int                    `json:"suit,omitempty"`
	Config    *SergeantMajorWebConfig `json:"config,omitempty"`
}

// SergeantMajorWebConfig サージェントメジャーWeb設定
type SergeantMajorWebConfig struct {
	Rounds *int `json:"rounds,omitempty"`
}

// SergeantMajorWebOutputPlayer サージェントメジャーWebアウトプットプレイヤー
type SergeantMajorWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Target はこのラウンドのノルマ (8 / 5 / 3)。**席順で決まり、宣言しません。**
	Target     int `json:"target"`
	TrickCount int `json:"trickCount"`
	// Score は「ノルマとの差」の累計。**勝敗はこれで決まる。**
	Score int `json:"score"`
}

// SergeantMajorWebOutputHint ヒント出力
type SergeantMajorWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
	// Suit は切り札に勧めるスート（それ以外は 0）。
	Suit int `json:"suit"`
	// Indices は捨てるべき札（捨て札フェーズ以外は空）。
	Indices []int `json:"indices"`
}

// SergeantMajorWebOutput サージェントメジャーWebアウトプット
type SergeantMajorWebOutput struct {
	Players     []*SergeantMajorWebOutputPlayer `json:"players"`
	Phase       int                             `json:"phase"`
	RoundNumber int                             `json:"roundNumber"`
	TrickNumber int                             `json:"trickNumber"`
	// TrumpSuit は 0 のあいだ未宣言。決めるのは親（ノルマ 8 の席）だけ。
	TrumpSuit int `json:"trumpSuit"`
	// KittyIndices は人間の手札のうち、今回のキティから入ってきた札の位置。
	//
	// **取り込むと手札に紛れて見分けが付かなくなる** (#5759)。捨て終われば空。
	KittyIndices []int `json:"kittyIndices"`
	// KittySize は親が取り込む余り札、DiscardCount は捨てる枚数。
	KittySize    int `json:"kittySize"`
	DiscardCount int `json:"discardCount"`
	// LastExchange は直前のラウンド間で動いた札の枚数。
	LastExchange     int                         `json:"lastExchange"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	DealerIdx        int                         `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	ValidPlays       []int                       `json:"validPlays"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerIdx        int                         `json:"winnerIdx"`
	Hint             *SergeantMajorWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config SergeantMajorWebOutputConfig `json:"config"`
}

// SergeantMajorWebOutputConfig サージェントメジャー設定アウトプット
type SergeantMajorWebOutputConfig struct {
	Rounds int `json:"rounds"`
}

// ToConfig builds a SergeantMajorConfig from the nested web config, applying bounds checking.
//
// **範囲に収めるだけでは足りない。** ラウンド数は 3 の倍数でなければ親の役が
// 一巡せず、1 人だけノルマ 8 を多く引き受けることになる（レビュー指摘 PR #5311）。
// クライアントが 4 を送ってきたらエラーにするのではなく、直近の倍数へ丸める。
func (c *SergeantMajorWebConfig) ToConfig() domain.SergeantMajorConfig {
	cfg := domain.DefaultSergeantMajorConfig()
	rounds := webutil.BoundedIntPtr(c.Rounds,
		domain.SergeantMajorRoundsMin, domain.SergeantMajorRoundsMax, cfg.Rounds)
	cfg.Rounds = rounds - rounds%domain.SergeantMajorPlayerCnt
	if cfg.Rounds < domain.SergeantMajorRoundsMin {
		cfg.Rounds = domain.SergeantMajorRoundsMin
	}
	return cfg
}

// ToConfig builds a SergeantMajorConfig from the web input.
func (p SergeantMajorWebInput) ToConfig() domain.SergeantMajorConfig {
	return configOrDefault(p.Config, (*SergeantMajorWebConfig).ToConfig, domain.DefaultSergeantMajorConfig())
}

// SergeantMajorWebController サージェントメジャーWebコントローラークラス
type SergeantMajorWebController = GameWebController[usecase.SergeantMajorInteractorIF, SergeantMajorWebInput, *SergeantMajorWebOutput]

// NewSergeantMajorWebController and NewSergeantMajorWebControllerWithProvider are
// the standard and provider-backed constructors for SergeantMajorWebController.
var NewSergeantMajorWebController, NewSergeantMajorWebControllerWithProvider = webControllerPair[usecase.SergeantMajorInteractorIF, SergeantMajorWebInput, *SergeantMajorWebOutput](
	newSergeantMajorDefaultOutput, sergeantMajorDispatch,
)

func newSergeantMajorDefaultOutput(msg string) *SergeantMajorWebOutput {
	return &SergeantMajorWebOutput{
		Players:       make([]*SergeantMajorWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		ValidPlays:    make([]int, 0),
		DiscardCount:  domain.SergeantMajorKittySize,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sergeantMajorDispatch(bc *baseController, w http.ResponseWriter, si usecase.SergeantMajorInteractorIF, param SergeantMajorWebInput, newDefault func(string) *SergeantMajorWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "t", "trump":
		// **スート無しの宣言は通さない。** 既定値で埋めると選んでいないスートが
		// そのラウンドの切り札になる。
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.DeclareTrump(*param.Suit))
	case "d", "discard":
		// **枚数はドメインが検証する。** ここでは「指定が無い」だけを弾く。
		if !requireParam(bc, w, newDefault, len(param.Discards) == 0, "param error: discards is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Discard(param.Discards))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextRound())
	case "g", "giveup":
		bc.writePresenterResponse(w, si.GiveUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
