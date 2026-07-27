//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NapWebInput ナップのWebインプット
type NapWebInput struct {
	BaseWebInput
	// Bid 入札種別 (0=Pass,2=Two,3=Three,4=Four,5=Nap)
	Bid *int `json:"bid,omitempty"`
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *NapWebConfig `json:"config,omitempty"`
}

// NapWebConfig ナップのWeb設定
type NapWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// NapWebOutputPlayer ナップのWebアウトプットプレイヤー
type NapWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// NapWebOutput ナップのWebアウトプット
type NapWebOutput struct {
	Players          []*NapWebOutputPlayer    `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	DeclarerIdx      int                      `json:"declarerIdx"`
	Contract         int                      `json:"contract"`
	TrumpSuit        int                      `json:"trumpSuit"`
	Bids             [domain.NapPlayerCnt]int `json:"bids"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	PlayerScores     [domain.NapPlayerCnt]int `json:"playerScores"`
	RoundTricks      [domain.NapPlayerCnt]int `json:"roundTricks"`
	PlayableIndices  []int                    `json:"playableIndices"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerPlayer     int                      `json:"winnerPlayer"`
	IsHumanTurn      bool                     `json:"isHumanTurn"`
	IsHumanBidTurn   bool                     `json:"isHumanBidTurn"`
	Hint             *WebOutputCardHint       `json:"hint,omitempty"`
	WebOutputBase
	Config NapWebOutputConfig `json:"config"`
}

// NapWebOutputConfig ナップの設定アウトプット
type NapWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a NapConfig from the nested web config, applying bounds checking.
func (c *NapWebConfig) ToConfig() domain.NapConfig {
	cfg := domain.DefaultNapConfig()
	cfg.CpuDifficulty = domain.NapCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.NapCpuDifficultyEasy), int(domain.NapCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a NapConfig from the web input.
func (p NapWebInput) ToConfig() domain.NapConfig {
	return configOrDefault(p.Config, (*NapWebConfig).ToConfig, domain.DefaultNapConfig())
}

// NapWebController ナップのWebコントローラークラス
type NapWebController = GameWebController[usecase.NapInteractorIF, NapWebInput, *NapWebOutput]

// NewNapWebController and NewNapWebControllerWithProvider are
// the standard and provider-backed constructors for NapWebController.
var NewNapWebController, NewNapWebControllerWithProvider = webControllerPair[usecase.NapInteractorIF, NapWebInput, *NapWebOutput](
	newNapDefaultOutput, napDispatch,
)

func newNapDefaultOutput(msg string) *NapWebOutput {
	return &NapWebOutput{
		Players:         make([]*NapWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func napDispatch(bc *baseController, w http.ResponseWriter, di usecase.NapInteractorIF, param NapWebInput, newDefault func(string) *NapWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(*param.Bid))
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
