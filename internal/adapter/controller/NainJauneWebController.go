//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NainJauneWebInput ル・ナン・ジョーヌWebインプット
type NainJauneWebInput struct {
	BaseWebInput
	CardIndex *int                `json:"cardIndex,omitempty"`
	Config    *NainJauneWebConfig `json:"config,omitempty"`
}

// NainJauneWebConfig ル・ナン・ジョーヌWeb設定
type NainJauneWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// NainJauneWebOutputBox は盤の 1 区画。
type NainJauneWebOutputBox struct {
	// Name は "ten" / "jack" / "queen" / "king" / "dwarf"。
	Name string `json:"name"`
	// Chips は残高。**取られなかった区画は持ち越される**ので貯まっていく。
	Chips int `json:"chips"`
	// Card はこの区画を取る札。**スートまで一致していなければ取れない**ので、
	// クライアントに書き写させずサーバーから送る。
	Card *WebOutputCard `json:"card"`
}

// NainJauneWebOutputAward は区画が誰にいくら渡ったかの記録。
type NainJauneWebOutputAward struct {
	Box    string `json:"box"`
	Player int    `json:"player"`
	Chips  int    `json:"chips"`
}

// NainJauneWebOutputPlayer ル・ナン・ジョーヌWebアウトプットプレイヤー
type NainJauneWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	Chips     int              `json:"chips"`
	// Points は手札に残っている失点。**支払いは枚数ではなく点数**なので、
	// 枚数だけでは負債額が読めない。
	Points int  `json:"points"`
	Hidden bool `json:"hidden"`
}

// NainJauneWebOutputHint ヒント出力
type NainJauneWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// NainJauneWebOutput ル・ナン・ジョーヌWebアウトプット
type NainJauneWebOutput struct {
	Players []*NainJauneWebOutputPlayer `json:"players"`
	Phase   int                         `json:"phase"`
	// ValidPlays は人間の手番でのみ埋まる、今出せる手札のインデックス。
	// **並びに従う義務がある**ので、出す前に示さないと押して初めて弾かれる。
	ValidPlays       []int                    `json:"validPlays"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	Boxes            []*NainJauneWebOutputBox `json:"boxes"`
	// TalonCount は配り切らなかった残りの枚数。誰も使わない。
	TalonCount int                        `json:"talonCount"`
	Awards     []*NainJauneWebOutputAward `json:"awards"`
	PlayedPile []*WebOutputCard           `json:"playedPile"`
	// RunRank は今の並びの最高ランク。**0 は「好きな札で始められる」。**
	// スートは持たない -- この game は無関係。
	RunRank     int                     `json:"runRank"`
	DealNo      int                     `json:"dealNo"`
	TargetDeals int                     `json:"targetDeals"`
	DealWinner  int                     `json:"dealWinner"`
	GameEndFlag bool                    `json:"gameEndFlag"`
	WinnerIdx   int                     `json:"winnerIdx"`
	Hint        *NainJauneWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config NainJauneWebOutputConfig `json:"config"`
}

// NainJauneWebOutputConfig ル・ナン・ジョーヌ設定アウトプット
type NainJauneWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig builds a NainJauneConfig from the nested web config, applying bounds checking.
func (c *NainJauneWebConfig) ToConfig() domain.NainJauneConfig {
	cfg := domain.DefaultNainJauneConfig()
	cfg.CpuDifficulty = domain.NainJauneCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.NainJauneCpuDifficultyNormal), int(domain.NainJauneCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals, 1, 100)
	return cfg
}

// ToConfig builds a NainJauneConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *NainJauneWebConfig and calling the method on it
// would dereference nil.
func (i NainJauneWebInput) ToConfig() domain.NainJauneConfig {
	return configOrDefault(i.Config, (*NainJauneWebConfig).ToConfig, domain.DefaultNainJauneConfig())
}

// NainJauneWebController ル・ナン・ジョーヌWebコントローラ
type NainJauneWebController = GameWebController[usecase.NainJauneInteractorIF, NainJauneWebInput, *NainJauneWebOutput]

// NewNainJauneWebController and NewNainJauneWebControllerWithProvider are the
// standard and provider-backed constructors for NainJauneWebController.
var NewNainJauneWebController, NewNainJauneWebControllerWithProvider = webControllerPair[usecase.NainJauneInteractorIF, NainJauneWebInput, *NainJauneWebOutput](
	newNainJauneDefaultOutput, nainJauneDispatch,
)

func newNainJauneDefaultOutput(msg string) *NainJauneWebOutput {
	return &NainJauneWebOutput{
		Players:       make([]*NainJauneWebOutputPlayer, 0),
		Boxes:         make([]*NainJauneWebOutputBox, 0),
		Awards:        make([]*NainJauneWebOutputAward, 0),
		PlayedPile:    make([]*WebOutputCard, 0),
		TargetDeals:   domain.DefaultNainJauneConfig().TargetDeals,
		DealWinner:    -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func nainJauneDispatch(bc *baseController, w http.ResponseWriter, ni usecase.NainJauneInteractorIF, param NainJauneWebInput, newDefault func(string) *NainJauneWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ni.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ni.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ni.NextDeal())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ni.Hint, ni.ActionLog)
	}
	return true
}
