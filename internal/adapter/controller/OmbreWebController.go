//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OmbreWebInput オンブル (Ombre) のWebインプット
type OmbreWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Bid ビッド宣言 (0=pass, 1=entrar, 2=solo)
	Bid *int `json:"bid,omitempty"`
	// TrumpSuit ビッド勝利時に選ぶ切り札スート (1=spade, 2=club, 3=heart, 4=diamond)
	TrumpSuit *int `json:"trumpSuit,omitempty"`
	// Config ゲーム設定
	Config *OmbreWebConfig `json:"config,omitempty"`
}

// OmbreWebConfig オンブルのWeb設定
type OmbreWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// OmbreWebOutputPlayer オンブルのWebアウトプットプレイヤー
type OmbreWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	IsOmbre    bool             `json:"isOmbre"`
}

// OmbreWebOutput オンブルのWebアウトプット
type OmbreWebOutput struct {
	Players          []*OmbreWebOutputPlayer    `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	CurrentBidderIdx int                        `json:"currentBidderIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	ForehandIdx      int                        `json:"forehandIdx"`
	OmbreIdx         int                        `json:"ombreIdx"`
	WinningBid       int                        `json:"winningBid"`
	TrumpSuit        int                        `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	PlayerScores     [domain.OmbrePlayerCnt]int `json:"playerScores"`
	LastTrickWinner  int                        `json:"lastTrickWinner"`
	Outcome          int                        `json:"outcome"`
	Result           int                        `json:"result"`
	PlayableIndices  []int                      `json:"playableIndices"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerPlayer     int                        `json:"winnerPlayer"`
	IsHumanTurn      bool                       `json:"isHumanTurn"`
	IsHumanBidTurn   bool                       `json:"isHumanBidTurn"`
	Hint             *WebOutputCardHint         `json:"hint,omitempty"`
	WebOutputBase
	Config OmbreWebOutputConfig `json:"config"`
}

// OmbreWebOutputConfig オンブルの設定アウトプット
type OmbreWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds an OmbreConfig from the nested web config, applying bounds checking.
func (c *OmbreWebConfig) ToConfig() domain.OmbreConfig {
	cfg := domain.DefaultOmbreConfig()
	cfg.CpuDifficulty = domain.OmbreCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.OmbreCpuDifficultyEasy), int(domain.OmbreCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 1000)
	return cfg
}

// ToConfig builds an OmbreConfig from the web input.
func (p OmbreWebInput) ToConfig() domain.OmbreConfig {
	return configOrDefault(p.Config, (*OmbreWebConfig).ToConfig, domain.DefaultOmbreConfig())
}

// OmbreWebController オンブルのWebコントローラークラス
type OmbreWebController = GameWebController[usecase.OmbreInteractorIF, OmbreWebInput, *OmbreWebOutput]

// NewOmbreWebController and NewOmbreWebControllerWithProvider are
// the standard and provider-backed constructors for OmbreWebController.
var NewOmbreWebController, NewOmbreWebControllerWithProvider = webControllerPair[usecase.OmbreInteractorIF, OmbreWebInput, *OmbreWebOutput](
	newOmbreDefaultOutput, ombreDispatch,
)

func newOmbreDefaultOutput(msg string) *OmbreWebOutput {
	return &OmbreWebOutput{
		Players:         make([]*OmbreWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		OmbreIdx:        -1,
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func ombreDispatch(bc *baseController, w http.ResponseWriter, di usecase.OmbreInteractorIF, param OmbreWebInput, newDefault func(string) *OmbreWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		trump := -1
		if param.TrumpSuit != nil {
			trump = *param.TrumpSuit
		}
		bc.writePresenterResponse(w, di.Bid(domain.OmbreBid(*param.Bid), trump))
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
