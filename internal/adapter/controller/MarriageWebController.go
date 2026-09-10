//go:build !js || !wasm || extra5

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MarriageWebInput マリッジ Web インプット
type MarriageWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *MarriageWebConfig `json:"config,omitempty"`
}

// MarriageWebConfig マリッジ Web 設定
type MarriageWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// MarriageWebOutputPlayer プレイヤーのアウトプット
type MarriageWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	Deadwood        int              `json:"deadwood"`
	Maal            int              `json:"maal"`
	HasPureSequence bool             `json:"hasPureSequence"`
}

// MarriageWebOutput マリッジ Web アウトプット
type MarriageWebOutput struct {
	Players          []*MarriageWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TargetRounds     int                        `json:"targetRounds"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	DiscardTop       *WebOutputCard             `json:"discardTop"`
	DrawPileCount    int                        `json:"drawPileCount"`
	WildJoker        *WebOutputCard             `json:"wildJoker"`
	WildRank         int                        `json:"wildRank"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	DeclarerIdx      int                        `json:"declarerIdx"`
	DeclarationValid bool                       `json:"declarationValid"`
	WebOutputBase
	Config MarriageWebOutputConfig `json:"config"`
}

// MarriageWebOutputConfig 設定アウトプット
type MarriageWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds an MarriageConfig from the nested web config, applying bounds checking.
func (c *MarriageWebConfig) ToConfig() domain.MarriageConfig {
	cfg := domain.DefaultMarriageConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.MarriagePlayerCountMin, domain.MarriagePlayerCountMax)
	cfg.CpuDifficulty = domain.MarriageCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.MarriageCpuDifficultyEasy),
		int(domain.MarriageCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 100)
	return cfg
}

// ToConfig builds an MarriageConfig from the web input.
func (p MarriageWebInput) ToConfig() domain.MarriageConfig {
	return configOrDefault(p.Config, (*MarriageWebConfig).ToConfig, domain.DefaultMarriageConfig())
}

// MarriageWebController マリッジ Web コントローラー
type MarriageWebController = GameWebController[usecase.MarriageInteractorIF, MarriageWebInput, *MarriageWebOutput]

// NewMarriageWebController / NewMarriageWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewMarriageWebController, NewMarriageWebControllerWithProvider = webControllerPair[usecase.MarriageInteractorIF, MarriageWebInput, *MarriageWebOutput](
	newMarriageDefaultOutput, marriageDispatch,
)

func newMarriageDefaultOutput(msg string) *MarriageWebOutput {
	return &MarriageWebOutput{
		Players:       make([]*MarriageWebOutputPlayer, 0),
		WinnerIdx:     -1,
		DeclarerIdx:   -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func marriageDispatch(bc *baseController, w http.ResponseWriter, ci usecase.MarriageInteractorIF, param MarriageWebInput, newDefault func(string) *MarriageWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "ds", "drawstock":
		bc.writePresenterResponse(w, ci.DrawFromStock())
	case "dd", "drawdiscard":
		bc.writePresenterResponse(w, ci.DrawFromDiscard())
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Discard(*param.CardIndex))
	case "de", "declare":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Declare(*param.CardIndex))
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ci.ActionLog)
	}
	return true
}
