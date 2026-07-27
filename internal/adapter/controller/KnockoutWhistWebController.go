//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KnockoutWhistWebInput ノックアウト・ホイストのWebインプット
type KnockoutWhistWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// TrumpSuit 選択する切り札スート 1-4 (selecttrump コマンド)
	TrumpSuit *int `json:"trumpSuit,omitempty"`
	// Config ゲーム設定
	Config *KnockoutWhistWebConfig `json:"config,omitempty"`
}

// KnockoutWhistWebConfig ノックアウト・ホイストのWeb設定
type KnockoutWhistWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// KnockoutWhistWebOutputPlayer ノックアウト・ホイストのWebアウトプットプレイヤー
type KnockoutWhistWebOutputPlayer struct {
	ID          int              `json:"id"`
	IsHuman     bool             `json:"isHuman"`
	CardCount   int              `json:"cardCount"`
	Cards       []*WebOutputCard `json:"cards"`
	TrickCount  int              `json:"trickCount"`
	Eliminated  bool             `json:"eliminated"`
	Dogbones    int              `json:"dogbones"`
	RoundTricks int              `json:"roundTricks"`
}

// KnockoutWhistWebOutput ノックアウト・ホイストのWebアウトプット
type KnockoutWhistWebOutput struct {
	Players          []*KnockoutWhistWebOutputPlayer `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	HandSize         int                             `json:"handSize"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	DealerIdx        int                             `json:"dealerIdx"`
	TrumpSuit        int                             `json:"trumpSuit"`
	RoundWinnerIdx   int                             `json:"roundWinnerIdx"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	ActiveCount      int                             `json:"activeCount"`
	PlayableIndices  []int                           `json:"playableIndices"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerPlayer     int                             `json:"winnerPlayer"`
	IsHumanTurn      bool                            `json:"isHumanTurn"`
	Hint             *WebOutputCardHint              `json:"hint,omitempty"`
	WebOutputBase
	Config KnockoutWhistWebOutputConfig `json:"config"`
}

// KnockoutWhistWebOutputConfig ノックアウト・ホイストの設定アウトプット
type KnockoutWhistWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a KnockoutWhistConfig from the nested web config, applying bounds checking.
func (c *KnockoutWhistWebConfig) ToConfig() domain.KnockoutWhistConfig {
	cfg := domain.DefaultKnockoutWhistConfig()
	cfg.CpuDifficulty = domain.KnockoutWhistCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.KnockoutWhistCpuDifficultyEasy), int(domain.KnockoutWhistCpuDifficultyHard), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a KnockoutWhistConfig from the web input.
func (p KnockoutWhistWebInput) ToConfig() domain.KnockoutWhistConfig {
	return configOrDefault(p.Config, (*KnockoutWhistWebConfig).ToConfig, domain.DefaultKnockoutWhistConfig())
}

// KnockoutWhistWebController ノックアウト・ホイストのWebコントローラークラス
type KnockoutWhistWebController = GameWebController[usecase.KnockoutWhistInteractorIF, KnockoutWhistWebInput, *KnockoutWhistWebOutput]

// NewKnockoutWhistWebController and NewKnockoutWhistWebControllerWithProvider are
// the standard and provider-backed constructors for KnockoutWhistWebController.
var NewKnockoutWhistWebController, NewKnockoutWhistWebControllerWithProvider = webControllerPair[usecase.KnockoutWhistInteractorIF, KnockoutWhistWebInput, *KnockoutWhistWebOutput](
	newKnockoutWhistDefaultOutput, knockoutWhistDispatch,
)

func newKnockoutWhistDefaultOutput(msg string) *KnockoutWhistWebOutput {
	return &KnockoutWhistWebOutput{
		Players:         make([]*KnockoutWhistWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		RoundWinnerIdx:  -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func knockoutWhistDispatch(bc *baseController, w http.ResponseWriter, di usecase.KnockoutWhistInteractorIF, param KnockoutWhistWebInput, newDefault func(string) *KnockoutWhistWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	case "st", "selecttrump":
		if !requireParam(bc, w, newDefault, param.TrumpSuit == nil, "param error: trumpSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.SelectTrump(*param.TrumpSuit))
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
