//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TarocchiniWebInput タロッキーニのWebインプット
type TarocchiniWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices スカルトで捨てるカードのインデックス (2 枚)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Config ゲーム設定
	Config *TarocchiniWebConfig `json:"config,omitempty"`
}

// TarocchiniWebConfig タロッキーニのWeb設定
type TarocchiniWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// TarocchiniWebOutputPlayer タロッキーニのWebアウトプットプレイヤー
type TarocchiniWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Team       int              `json:"team"`
	IsDealer   bool             `json:"isDealer"`
}

// TarocchiniWebOutput タロッキーニのWebアウトプット
type TarocchiniWebOutput struct {
	Players          []*TarocchiniWebOutputPlayer    `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	DealerIdx        int                             `json:"dealerIdx"`
	ScartoCount      int                             `json:"scartoCount"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	TeamScores       [2]int                          `json:"teamScores"`
	RoundTricks      [domain.TarocchiniPlayerCnt]int `json:"roundTricks"`
	LastTrickWinner  int                             `json:"lastTrickWinner"`
	PlayableIndices  []int                           `json:"playableIndices"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerTeam       int                             `json:"winnerTeam"`
	IsHumanTurn      bool                            `json:"isHumanTurn"`
	IsHumanScarto    bool                            `json:"isHumanScarto"`
	Hint             *WebOutputCardHint              `json:"hint,omitempty"`
	WebOutputBase
	Config TarocchiniWebOutputConfig `json:"config"`
}

// TarocchiniWebOutputConfig タロッキーニの設定アウトプット
type TarocchiniWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds a TarocchiniConfig from the nested web config, applying bounds checking.
func (c *TarocchiniWebConfig) ToConfig() domain.TarocchiniConfig {
	cfg := domain.DefaultTarocchiniConfig()
	cfg.CpuDifficulty = domain.TarocchiniCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.TarocchiniCpuDifficultyEasy), int(domain.TarocchiniCpuDifficultyHard), int(cfg.CpuDifficulty)))
	// **下限は TarocchiniPlayerCnt。**1 を許すと境界検査は通るが Validate が落とすので、
	// リセットが黙って無視される。倍数条件は Validate 側が見る。
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.TarocchiniPlayerCnt, 1000)
	return cfg
}

// ToConfig builds a TarocchiniConfig from the web input.
func (p TarocchiniWebInput) ToConfig() domain.TarocchiniConfig {
	return configOrDefault(p.Config, (*TarocchiniWebConfig).ToConfig, domain.DefaultTarocchiniConfig())
}

// TarocchiniWebController タロッキーニのWebコントローラークラス
type TarocchiniWebController = GameWebController[usecase.TarocchiniInteractorIF, TarocchiniWebInput, *TarocchiniWebOutput]

// NewTarocchiniWebController and NewTarocchiniWebControllerWithProvider are the standard
// and provider-backed constructors for TarocchiniWebController.
var NewTarocchiniWebController, NewTarocchiniWebControllerWithProvider = webControllerPair[usecase.TarocchiniInteractorIF, TarocchiniWebInput, *TarocchiniWebOutput](
	newTarocchiniDefaultOutput, tarocchiniDispatch,
)

func newTarocchiniDefaultOutput(msg string) *TarocchiniWebOutput {
	return &TarocchiniWebOutput{
		Players:         make([]*TarocchiniWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func tarocchiniDispatch(bc *baseController, w http.ResponseWriter, di usecase.TarocchiniInteractorIF, param TarocchiniWebInput, newDefault func(string) *TarocchiniWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "s", "scarto", "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(param.CardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
