//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SoloWhistWebInput ソロ・ホイストのWebインプット
type SoloWhistWebInput struct {
	BaseWebInput
	// Bid 入札種別 (0=Pass,1=Solo,2=Misère,3=Abundance)
	Bid *int `json:"bid,omitempty"`
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *SoloWhistWebConfig `json:"config,omitempty"`
}

// SoloWhistWebConfig ソロ・ホイストのWeb設定
type SoloWhistWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// SoloWhistWebOutputPlayer ソロ・ホイストのWebアウトプットプレイヤー
type SoloWhistWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// SoloWhistWebOutput ソロ・ホイストのWebアウトプット
type SoloWhistWebOutput struct {
	Players          []*SoloWhistWebOutputPlayer    `json:"players"`
	Phase            int                            `json:"phase"`
	RoundNumber      int                            `json:"roundNumber"`
	TrickNumber      int                            `json:"trickNumber"`
	CurrentPlayerIdx int                            `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                            `json:"leadPlayerIdx"`
	DealerIdx        int                            `json:"dealerIdx"`
	DeclarerIdx      int                            `json:"declarerIdx"`
	Contract         int                            `json:"contract"`
	TrumpSuit        int                            `json:"trumpSuit"`
	Bids             [domain.SoloWhistPlayerCnt]int `json:"bids"`
	CurrentTrick     []*WebOutputTrickCard          `json:"currentTrick"`
	PlayerScores     [domain.SoloWhistPlayerCnt]int `json:"playerScores"`
	RoundTricks      [domain.SoloWhistPlayerCnt]int `json:"roundTricks"`
	PlayableIndices  []int                          `json:"playableIndices"`
	GameEndFlag      bool                           `json:"gameEndFlag"`
	WinnerPlayer     int                            `json:"winnerPlayer"`
	IsHumanTurn      bool                           `json:"isHumanTurn"`
	IsHumanBidTurn   bool                           `json:"isHumanBidTurn"`
	Hint             *WebOutputCardHint             `json:"hint,omitempty"`
	WebOutputBase
	Config SoloWhistWebOutputConfig `json:"config"`
}

// SoloWhistWebOutputConfig ソロ・ホイストの設定アウトプット
type SoloWhistWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a SoloWhistConfig from the nested web config, applying bounds checking.
func (c *SoloWhistWebConfig) ToConfig() domain.SoloWhistConfig {
	cfg := domain.DefaultSoloWhistConfig()
	cfg.CpuDifficulty = domain.SoloWhistCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SoloWhistCpuDifficultyEasy), int(domain.SoloWhistCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a SoloWhistConfig from the web input.
func (p SoloWhistWebInput) ToConfig() domain.SoloWhistConfig {
	return configOrDefault(p.Config, (*SoloWhistWebConfig).ToConfig, domain.DefaultSoloWhistConfig())
}

// SoloWhistWebController ソロ・ホイストのWebコントローラークラス
type SoloWhistWebController = GameWebController[usecase.SoloWhistInteractorIF, SoloWhistWebInput, *SoloWhistWebOutput]

// NewSoloWhistWebController and NewSoloWhistWebControllerWithProvider are
// the standard and provider-backed constructors for SoloWhistWebController.
var NewSoloWhistWebController, NewSoloWhistWebControllerWithProvider = webControllerPair[usecase.SoloWhistInteractorIF, SoloWhistWebInput, *SoloWhistWebOutput](
	newSoloWhistDefaultOutput, soloWhistDispatch,
)

func newSoloWhistDefaultOutput(msg string) *SoloWhistWebOutput {
	return &SoloWhistWebOutput{
		Players:         make([]*SoloWhistWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func soloWhistDispatch(bc *baseController, w http.ResponseWriter, di usecase.SoloWhistInteractorIF, param SoloWhistWebInput, newDefault func(string) *SoloWhistWebOutput) bool {
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
