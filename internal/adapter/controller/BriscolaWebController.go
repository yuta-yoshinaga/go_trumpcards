package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BriscolaWebInput ブリスコラWebインプット
type BriscolaWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *BriscolaWebConfig `json:"config,omitempty"`
}

// BriscolaWebConfig ブリスコラWeb設定
type BriscolaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// BriscolaWebOutputPlayer ブリスコラWebアウトプットプレイヤー
type BriscolaWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Points     int              `json:"points"`
	TrickCount int              `json:"trickCount"`
}

// BriscolaWebOutputHint ヒント出力
type BriscolaWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// BriscolaWebOutput ブリスコラWebアウトプット
type BriscolaWebOutput struct {
	Players          []*BriscolaWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	TrickNumber      int                        `json:"trickNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	TrumpSuit        int                        `json:"trumpSuit"`
	TrumpCard        *WebOutputCard             `json:"trumpCard,omitempty"`
	DealerIdx        int                        `json:"dealerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	StockRemaining   int                        `json:"stockRemaining"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	Hint             *BriscolaWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config BriscolaWebOutputConfig `json:"config"`
}

// BriscolaWebOutputConfig ブリスコラ設定アウトプット
type BriscolaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a BriscolaConfig from the nested web config, applying bounds checking.
func (c *BriscolaWebConfig) ToConfig() domain.BriscolaConfig {
	cfg := domain.DefaultBriscolaConfig()
	cfg.CpuDifficulty = domain.BriscolaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.BriscolaCpuDifficultyNormal), int(domain.BriscolaCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a BriscolaConfig from the web input.
func (p BriscolaWebInput) ToConfig() domain.BriscolaConfig {
	return configOrDefault(p.Config, (*BriscolaWebConfig).ToConfig, domain.DefaultBriscolaConfig())
}

// BriscolaWebController ブリスコラWebコントローラークラス
type BriscolaWebController = GameWebController[usecase.BriscolaInteractorIF, BriscolaWebInput, *BriscolaWebOutput]

// NewBriscolaWebController and NewBriscolaWebControllerWithProvider are
// the standard and provider-backed constructors for BriscolaWebController.
var NewBriscolaWebController, NewBriscolaWebControllerWithProvider = webControllerPair[usecase.BriscolaInteractorIF, BriscolaWebInput, *BriscolaWebOutput](
	newBriscolaDefaultOutput, briscolaDispatch,
)

func newBriscolaDefaultOutput(msg string) *BriscolaWebOutput {
	return &BriscolaWebOutput{
		Players:       make([]*BriscolaWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func briscolaDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BriscolaInteractorIF, param BriscolaWebInput, newDefault func(string) *BriscolaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextTrick())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
