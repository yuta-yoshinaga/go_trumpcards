//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BuraWebInput ブラWebインプット
type BuraWebInput struct {
	BaseWebInput
	// CardIndices は出すカードの手札添字。ブラはリードで最大 3 枚まとめて
	// 出せるため、単一の cardIndex ではなく配列を受け取る。
	CardIndices []int          `json:"cardIndices,omitempty"`
	Config      *BuraWebConfig `json:"config,omitempty"`
}

// BuraWebConfig ブラWeb設定
type BuraWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// BuraWebOutputPlayer ブラWebアウトプットプレイヤー
type BuraWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	Points    int              `json:"points"`
	// Hidden は手札が伏せられていることを示す。伏せている間 Cards は空だが
	// CardCount は残るので、UI は枚数だけ裏向きで描ける。
	Hidden bool `json:"hidden"`
}

// BuraWebOutputHint ヒント出力
type BuraWebOutputHint struct {
	CardIndices []int  `json:"cardIndices,omitempty"`
	Reason      string `json:"reason"`
}

// BuraWebOutput ブラWebアウトプット
type BuraWebOutput struct {
	Players          []*BuraWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	TrickNumber      int                    `json:"trickNumber"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                    `json:"leadPlayerIdx"`
	CurrentLead      []*WebOutputCard       `json:"currentLead"`
	TrumpSuit        int                    `json:"trumpSuit"`
	TrumpCard        *WebOutputCard         `json:"trumpCard,omitempty"`
	StockRemaining   int                    `json:"stockRemaining"`
	WinThreshold     int                    `json:"winThreshold"`
	GameEndFlag      bool                   `json:"gameEndFlag"`
	WinnerIdx        int                    `json:"winnerIdx"`
	IsDraw           bool                   `json:"isDraw"`
	Hint             *BuraWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config BuraWebOutputConfig `json:"config"`
}

// BuraWebOutputConfig ブラ設定アウトプット
type BuraWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a BuraConfig from the nested web config, applying bounds checking.
func (c *BuraWebConfig) ToConfig() domain.BuraConfig {
	cfg := domain.DefaultBuraConfig()
	cfg.CpuDifficulty = domain.BuraCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.BuraCpuDifficultyNormal), int(domain.BuraCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a BuraConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *BuraWebConfig, and calling the method on it
// dereferences nil. Reaching straight for i.Config.ToConfig() panicked on every
// reset -- a 500 on the request that starts the game.
func (i BuraWebInput) ToConfig() domain.BuraConfig {
	return configOrDefault(i.Config, (*BuraWebConfig).ToConfig, domain.DefaultBuraConfig())
}

// BuraWebController ブラWebコントローラ
type BuraWebController = GameWebController[usecase.BuraInteractorIF, BuraWebInput, *BuraWebOutput]

// NewBuraWebController and NewBuraWebControllerWithProvider are
// the standard and provider-backed constructors for BuraWebController.
var NewBuraWebController, NewBuraWebControllerWithProvider = webControllerPair[usecase.BuraInteractorIF, BuraWebInput, *BuraWebOutput](
	newBuraDefaultOutput, buraDispatch,
)

func newBuraDefaultOutput(msg string) *BuraWebOutput {
	return &BuraWebOutput{
		Players:       make([]*BuraWebOutputPlayer, 0),
		CurrentLead:   make([]*WebOutputCard, 0),
		WinnerIdx:     -1,
		WinThreshold:  domain.BuraWinThreshold,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func buraDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BuraInteractorIF, param BuraWebInput, newDefault func(string) *BuraWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, len(param.CardIndices) == 0, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(param.CardIndices))
	case "c", "claim":
		bc.writePresenterResponse(w, bi.Claim())
	case "d", "declare":
		bc.writePresenterResponse(w, bi.Declare())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
