//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// IndianRummyWebInput インドラミー Web インプット
type IndianRummyWebInput struct {
	BaseWebInput
	CardIndex *int                  `json:"cardIndex,omitempty"`
	Config    *IndianRummyWebConfig `json:"config,omitempty"`
}

// IndianRummyWebConfig インドラミー Web 設定
type IndianRummyWebConfig struct {
	PlayerCount   *int `json:"playerCount,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// IndianRummyWebOutputPlayer プレイヤーのアウトプット
type IndianRummyWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	Deadwood        int              `json:"deadwood"`
	HasPureSequence bool             `json:"hasPureSequence"`
}

// IndianRummyWebOutput インドラミー Web アウトプット
type IndianRummyWebOutput struct {
	Players []*IndianRummyWebOutputPlayer `json:"players"`
	// HumanDeadwood は人間の手札のデッドウッド採点値、HumanHasPureSequence は
	// 必須のピュアシーケンスを満たしているか。CUI は毎ターン出しているのに Web は
	// 狭い条件でしか出していなかった (#4824)。
	HumanDeadwood        int            `json:"humanDeadwood"`
	HumanHasPureSequence bool           `json:"humanHasPureSequence"`
	Phase                int            `json:"phase"`
	RoundNumber          int            `json:"roundNumber"`
	TargetRounds         int            `json:"targetRounds"`
	CurrentPlayerIdx     int            `json:"currentPlayerIdx"`
	DealerIdx            int            `json:"dealerIdx"`
	DiscardTop           *WebOutputCard `json:"discardTop"`
	DrawPileCount        int            `json:"drawPileCount"`
	WildJoker            *WebOutputCard `json:"wildJoker"`
	WildRank             int            `json:"wildRank"`
	GameEndFlag          bool           `json:"gameEndFlag"`
	WinnerIdx            int            `json:"winnerIdx"`
	DeclarerIdx          int            `json:"declarerIdx"`
	DeclarationValid     bool           `json:"declarationValid"`
	WebOutputBase
	Config IndianRummyWebOutputConfig `json:"config"`
}

// IndianRummyWebOutputConfig 設定アウトプット
type IndianRummyWebOutputConfig struct {
	PlayerCount   int `json:"playerCount"`
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds an IndianRummyConfig from the nested web config, applying bounds checking.
func (c *IndianRummyWebConfig) ToConfig() domain.IndianRummyConfig {
	cfg := domain.DefaultIndianRummyConfig()
	webutil.ApplyBoundedInt(&cfg.PlayerCount, c.PlayerCount, domain.IndianRummyPlayerCountMin, domain.IndianRummyPlayerCountMax)
	cfg.CpuDifficulty = domain.IndianRummyCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.IndianRummyCpuDifficultyEasy),
		int(domain.IndianRummyCpuDifficultyHard),
		int(cfg.CpuDifficulty),
	))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 100)
	return cfg
}

// ToConfig builds an IndianRummyConfig from the web input.
func (p IndianRummyWebInput) ToConfig() domain.IndianRummyConfig {
	return configOrDefault(p.Config, (*IndianRummyWebConfig).ToConfig, domain.DefaultIndianRummyConfig())
}

// IndianRummyWebController インドラミー Web コントローラー
type IndianRummyWebController = GameWebController[usecase.IndianRummyInteractorIF, IndianRummyWebInput, *IndianRummyWebOutput]

// NewIndianRummyWebController / NewIndianRummyWebControllerWithProvider: 標準／provider 背後の 2 種類のコンストラクタ
var NewIndianRummyWebController, NewIndianRummyWebControllerWithProvider = webControllerPair[usecase.IndianRummyInteractorIF, IndianRummyWebInput, *IndianRummyWebOutput](
	newIndianRummyDefaultOutput, indianRummyDispatch,
)

func newIndianRummyDefaultOutput(msg string) *IndianRummyWebOutput {
	return &IndianRummyWebOutput{
		Players:       make([]*IndianRummyWebOutputPlayer, 0),
		WinnerIdx:     -1,
		DeclarerIdx:   -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func indianRummyDispatch(bc *baseController, w http.ResponseWriter, ci usecase.IndianRummyInteractorIF, param IndianRummyWebInput, newDefault func(string) *IndianRummyWebOutput) bool {
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
