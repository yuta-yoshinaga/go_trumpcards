//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DesmocheWebInput デスモチェWebインプット
type DesmocheWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices はメルドに出す手札の添字集合。
	CardIndices []int `json:"cardIndices,omitempty"`
	MeldIndex   *int  `json:"meldIndex,omitempty"`
	// FromMeldIndex / ToMeldIndex は desmoche (メルドの組み替え) の移動元と移動先。
	FromMeldIndex *int               `json:"fromMeldIndex,omitempty"`
	ToMeldIndex   *int               `json:"toMeldIndex,omitempty"`
	Config        *DesmocheWebConfig `json:"config,omitempty"`
}

// DesmocheWebConfig デスモチェWeb設定
type DesmocheWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// DesmocheWebOutputPlayer デスモチェWebアウトプットプレイヤー
type DesmocheWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。伏せている間も送る。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Score は収支。ポットを取ると増え、掛け金を出すと減る。
	Score int `json:"score"`
	// MeldedCount は場に出している総枚数。10 で上がりなので進捗そのもの。
	MeldedCount int  `json:"meldedCount"`
	Hidden      bool `json:"hidden"`
}

// DesmocheWebOutputMeld 場のメルド
type DesmocheWebOutputMeld struct {
	Owner int `json:"owner"`
	// Kind は 0=セット (同ランク), 1=ラン (同スートの並び)。
	Kind  int              `json:"kind"`
	Cards []*WebOutputCard `json:"cards"`
}

// DesmocheWebOutputHint ヒント出力
type DesmocheWebOutputHint struct {
	// CardIndices はメルドに出せる手札の添字 (あれば)。
	CardIndices []int `json:"cardIndices,omitempty"`
	CardIndex   *int  `json:"cardIndex,omitempty"`
	// DrawStock が真なら山札から引くのが推奨手。
	DrawStock bool   `json:"drawStock"`
	Reason    string `json:"reason"`
}

// DesmocheWebOutput デスモチェWebアウトプット
type DesmocheWebOutput struct {
	Players          []*DesmocheWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	StockCount       int                        `json:"stockCount"`
	// DiscardTop は捨て札の一番上。取るかどうかの判断材料。
	DiscardTop *WebOutputCard           `json:"discardTop,omitempty"`
	Melds      []*DesmocheWebOutputMeld `json:"melds"`
	RoundNo    int                      `json:"roundNo"`
	// Pot は場の掛け金。勝者なしのラウンドで持ち越されるので、伸びていく。
	Pot int `json:"pot"`
	// GoOutSize は上がりに要るメルドの総枚数 (10)。クライアントに書き写させない。
	GoOutSize int `json:"goOutSize"`
	// RoundWinner は直近ラウンドの勝者 (-1: 勝者なし)。
	RoundWinner int `json:"roundWinner"`
	// RoundExhausted は山札が尽きて勝者なしで終わったか。ポット持ち越しの根拠。
	RoundExhausted bool                   `json:"roundExhausted"`
	GameEndFlag    bool                   `json:"gameEndFlag"`
	WinnerIdx      int                    `json:"winnerIdx"`
	Hint           *DesmocheWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config DesmocheWebOutputConfig `json:"config"`
}

// DesmocheWebOutputConfig デスモチェ設定アウトプット
type DesmocheWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a DesmocheConfig from the nested web config, applying bounds checking.
func (c *DesmocheWebConfig) ToConfig() domain.DesmocheConfig {
	cfg := domain.DefaultDesmocheConfig()
	cfg.CpuDifficulty = domain.DesmocheCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.DesmocheCpuDifficultyNormal), int(domain.DesmocheCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a DesmocheConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *DesmocheWebConfig and calling the method on it would
// dereference nil.
func (i DesmocheWebInput) ToConfig() domain.DesmocheConfig {
	return configOrDefault(i.Config, (*DesmocheWebConfig).ToConfig, domain.DefaultDesmocheConfig())
}

// DesmocheWebController デスモチェWebコントローラ
type DesmocheWebController = GameWebController[usecase.DesmocheInteractorIF, DesmocheWebInput, *DesmocheWebOutput]

// NewDesmocheWebController and NewDesmocheWebControllerWithProvider are the
// standard and provider-backed constructors for DesmocheWebController.
var NewDesmocheWebController, NewDesmocheWebControllerWithProvider = webControllerPair[usecase.DesmocheInteractorIF, DesmocheWebInput, *DesmocheWebOutput](
	newDesmocheDefaultOutput, desmocheDispatch,
)

func newDesmocheDefaultOutput(msg string) *DesmocheWebOutput {
	return &DesmocheWebOutput{
		Players:       make([]*DesmocheWebOutputPlayer, 0),
		Melds:         make([]*DesmocheWebOutputMeld, 0),
		GoOutSize:     domain.DesmocheGoOutSize,
		RoundWinner:   -1,
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func desmocheDispatch(bc *baseController, w http.ResponseWriter, di usecase.DesmocheInteractorIF, param DesmocheWebInput, newDefault func(string) *DesmocheWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, di.DrawStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, di.DrawDiscard())
	case "m", "meld":
		if !requireParam(bc, w, newDefault, len(param.CardIndices) == 0, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Meld(param.CardIndices))
	case "o", "layoff":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil || param.MeldIndex == nil,
			"param error: cardIndex and meldIndex are required.") {
			return true
		}
		bc.writePresenterResponse(w, di.LayOff(*param.CardIndex, *param.MeldIndex))
	case "x", "desmoche":
		if !requireParam(bc, w, newDefault,
			param.FromMeldIndex == nil || param.CardIndex == nil || param.ToMeldIndex == nil,
			"param error: fromMeldIndex, cardIndex and toMeldIndex are required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Desmoche(*param.FromMeldIndex, *param.CardIndex, *param.ToMeldIndex))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}

// NewDesmocheDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
func NewDesmocheDefaultOutputForTest(msg string) *DesmocheWebOutput {
	return newDesmocheDefaultOutput(msg)
}
