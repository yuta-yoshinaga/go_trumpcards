//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PiquetWebInput Piquet Webインプット
type PiquetWebInput struct {
	BaseWebInput
	DiscardIndices []int            `json:"discardIndices,omitempty"`
	CardIndex      *int             `json:"cardIndex,omitempty"`
	Config         *PiquetWebConfig `json:"config,omitempty"`
}

// PiquetWebConfig Piquet Web設定
type PiquetWebConfig struct {
	CpuDifficulty  *int `json:"cpuDifficulty,omitempty"`
	DealsPerPartie *int `json:"dealsPerPartie,omitempty"`
}

// ToConfig builds a PiquetConfig from the nested web config, applying bounds checking.
func (c *PiquetWebConfig) ToConfig() domain.PiquetConfig {
	cfg := domain.DefaultPiquetConfig()
	cfg.CpuDifficulty = domain.PiquetCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.PiquetCpuDifficultyEasy),
		int(domain.PiquetCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.DealsPerPartie, c.DealsPerPartie, 1, 100)
	return cfg
}

// ToConfig builds a PiquetConfig from the web input.
func (p PiquetWebInput) ToConfig() domain.PiquetConfig {
	return configOrDefault(p.Config, (*PiquetWebConfig).ToConfig, domain.DefaultPiquetConfig())
}

// PiquetWebOutputPlayer プレイヤー状態
type PiquetWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	DeclScore  int              `json:"declScore"`
	TrickScore int              `json:"trickScore"`
	BonusScore int              `json:"bonusScore"`
	RoundScore int              `json:"roundScore"`
	MatchScore int              `json:"matchScore"`
}

// PiquetWebOutputClaim 宣言の中身
type PiquetWebOutputClaim struct {
	Length   int              `json:"length"`
	TopRank  int              `json:"topRank"`
	PipTotal int              `json:"pipTotal"`
	Suit     int              `json:"suit"`
	Cards    []*WebOutputCard `json:"cards"`
}

// PiquetWebOutputDeclaration 宣言結果
type PiquetWebOutputDeclaration struct {
	Kind         int                     `json:"kind"`
	ElderClaim   *PiquetWebOutputClaim   `json:"elderClaim,omitempty"`
	YoungerClaim *PiquetWebOutputClaim   `json:"youngerClaim,omitempty"`
	Winner       int                     `json:"winner"`
	Score        int                     `json:"score"`
	ScoredBy     int                     `json:"scoredBy"`
	Sets         []*PiquetWebOutputClaim `json:"sets,omitempty"`
}

// PiquetWebOutputHint ヒント
type PiquetWebOutputHint struct {
	CardIndex      *int   `json:"cardIndex,omitempty"`
	DiscardIndices []int  `json:"discardIndices,omitempty"`
	Reason         string `json:"reason"`
}

// PiquetWebOutputConfig 設定アウトプット
type PiquetWebOutputConfig struct {
	CpuDifficulty  int `json:"cpuDifficulty"`
	DealsPerPartie int `json:"dealsPerPartie"`
}

// PiquetWebOutput ゲーム状態
type PiquetWebOutput struct {
	Players              []*PiquetWebOutputPlayer      `json:"players"`
	Phase                int                           `json:"phase"`
	DealNumber           int                           `json:"dealNumber"`
	DealsPerPartie       int                           `json:"dealsPerPartie"`
	ElderIdx             int                           `json:"elderIdx"`
	YoungerIdx           int                           `json:"youngerIdx"`
	CurrentPlayerIdx     int                           `json:"currentPlayerIdx"`
	LeadPlayerIdx        int                           `json:"leadPlayerIdx"`
	TrickNumber          int                           `json:"trickNumber"`
	TricksWon            [2]int                        `json:"tricksWon"`
	ExchangeTurn         int                           `json:"exchangeTurn"`
	ElderExchangedCnt    int                           `json:"elderExchangedCnt"`
	YoungerExchangedCnt  int                           `json:"youngerExchangedCnt"`
	ElderTalon           []*WebOutputCard              `json:"elderTalon"`
	YoungerTalon         []*WebOutputCard              `json:"youngerTalon"`
	ElderRevealedTalon   []*WebOutputCard              `json:"elderRevealedTalon"`
	YoungerRevealedTalon []*WebOutputCard              `json:"youngerRevealedTalon"`
	CarteBlanche         [2]bool                       `json:"carteBlanche"`
	DeclStage            int                           `json:"declStage"`
	DeclResults          []*PiquetWebOutputDeclaration `json:"declResults"`
	CurrentTrick         []*WebOutputTrickCard         `json:"currentTrick"`
	LegalPlayIndices     []int                         `json:"legalPlayIndices,omitempty"`
	GameEndFlag          bool                          `json:"gameEndFlag"`
	WinnerIdx            int                           `json:"winnerIdx"`
	Hint                 *PiquetWebOutputHint          `json:"hint,omitempty"`
	WebOutputBase
	Config PiquetWebOutputConfig `json:"config"`
}

// PiquetWebController Piquet Webコントローラークラス
type PiquetWebController = GameWebController[usecase.PiquetInteractorIF, PiquetWebInput, *PiquetWebOutput]

// NewPiquetWebController + NewPiquetWebControllerWithProvider
var NewPiquetWebController, NewPiquetWebControllerWithProvider = webControllerPair[usecase.PiquetInteractorIF, PiquetWebInput, *PiquetWebOutput](
	newPiquetDefaultOutput, piquetDispatch,
)

func newPiquetDefaultOutput(msg string) *PiquetWebOutput {
	return &PiquetWebOutput{
		Players:       make([]*PiquetWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		DeclResults:   make([]*PiquetWebOutputDeclaration, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func piquetDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PiquetInteractorIF, param PiquetWebInput, newDefault func(string) *PiquetWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig()))
	case "e", "elder":
		if !requireParam(bc, w, newDefault, param.DiscardIndices == nil, "param error: discardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.ExchangeElder(param.DiscardIndices))
	case "y", "younger":
		if !requireParam(bc, w, newDefault, param.DiscardIndices == nil, "param error: discardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.ExchangeYounger(param.DiscardIndices))
	case "d", "declare":
		bc.writePresenterResponse(w, pi.ResolveDeclaration())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Play(*param.CardIndex))
	case "nd", "nextdeal":
		bc.writePresenterResponse(w, pi.NextDeal())
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}
