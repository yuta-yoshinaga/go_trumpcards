//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BridgeWebInput ブリッジWebインプット
type BridgeWebInput struct {
	BaseWebInput
	BidType   *int             `json:"bidType,omitempty"`
	BidLevel  *int             `json:"bidLevel,omitempty"`
	BidSuit   *int             `json:"bidSuit,omitempty"`
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *BridgeWebConfig `json:"config,omitempty"`
}

// BridgeWebConfig ブリッジWeb設定
type BridgeWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// BridgeWebOutputPlayer ブリッジWebアウトプットプレイヤー
type BridgeWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// BridgeWebOutputBidEntry ビッド履歴エントリ
type BridgeWebOutputBidEntry struct {
	PlayerIdx int `json:"playerIdx"`
	BidType   int `json:"bidType"`
	Level     int `json:"level"`
	Suit      int `json:"suit"`
}

// BridgeWebOutputHint ヒント出力
type BridgeWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	BidType   *int   `json:"bidType,omitempty"`
	BidLevel  *int   `json:"bidLevel,omitempty"`
	BidSuit   *int   `json:"bidSuit,omitempty"`
	Reason    string `json:"reason"`
}

// BridgeWebOutput ブリッジWebアウトプット
type BridgeWebOutput struct {
	Players          []*BridgeWebOutputPlayer   `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	BidPlayerIdx     int                        `json:"bidPlayerIdx"`
	DealerIdx        int                        `json:"dealerIdx"`
	TrumpSuit        int                        `json:"trumpSuit"`
	ContractLevel    int                        `json:"contractLevel"`
	ContractSuit     int                        `json:"contractSuit"`
	Doubled          int                        `json:"doubled"`
	DeclarerIdx      int                        `json:"declarerIdx"`
	DummyIdx         int                        `json:"dummyIdx"`
	BidHistory       []*BridgeWebOutputBidEntry `json:"bidHistory"`
	Vulnerability    [2]bool                    `json:"vulnerability"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	TeamScores       [2]int                     `json:"teamScores"`
	GamesWon         [2]int                     `json:"gamesWon"`
	BelowLine        [2]int                     `json:"belowLine"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerTeam       int                        `json:"winnerTeam"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	OpeningLeadDone  bool                       `json:"openingLeadDone"`
	DummyHand        []*WebOutputCard           `json:"dummyHand"`
	Hint             *BridgeWebOutputHint       `json:"hint,omitempty"`
	WebOutputBase
	Config BridgeWebOutputConfig `json:"config"`
}

// BridgeWebOutputConfig ブリッジ設定アウトプット
type BridgeWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a BridgeConfig from the nested web config, applying bounds checking.
func (c *BridgeWebConfig) ToConfig() domain.BridgeConfig {
	cfg := domain.DefaultBridgeConfig()
	cfg.CpuDifficulty = domain.BridgeCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.BridgeCpuDifficultyEasy), int(domain.BridgeCpuDifficultyHard), int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a BridgeConfig from the web input.
func (p BridgeWebInput) ToConfig() domain.BridgeConfig {
	return configOrDefault(p.Config, (*BridgeWebConfig).ToConfig, domain.DefaultBridgeConfig())
}

// BridgeWebController ブリッジWebコントローラークラス
type BridgeWebController = GameWebController[usecase.BridgeInteractorIF, BridgeWebInput, *BridgeWebOutput]

// NewBridgeWebController and NewBridgeWebControllerWithProvider are
// the standard and provider-backed constructors for BridgeWebController.
var NewBridgeWebController, NewBridgeWebControllerWithProvider = webControllerPair[usecase.BridgeInteractorIF, BridgeWebInput, *BridgeWebOutput](
	newBridgeDefaultOutput, bridgeDispatch,
)

func newBridgeDefaultOutput(msg string) *BridgeWebOutput {
	return &BridgeWebOutput{
		Players:       make([]*BridgeWebOutputPlayer, 0),
		BidHistory:    make([]*BridgeWebOutputBidEntry, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		DummyHand:     make([]*WebOutputCard, 0),
		WinnerTeam:    -1,
		DeclarerIdx:   -1,
		DummyIdx:      -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func bridgeDispatch(bc *baseController, w http.ResponseWriter, bi usecase.BridgeInteractorIF, param BridgeWebInput, newDefault func(string) *BridgeWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		bidType := 0
		if param.BidType != nil {
			bidType = *param.BidType
		}
		bidLevel := 0
		if param.BidLevel != nil {
			bidLevel = *param.BidLevel
		}
		bidSuit := 0
		if param.BidSuit != nil {
			bidSuit = *param.BidSuit
		}
		bc.writePresenterResponse(w, bi.Bid(bidType, bidLevel, bidSuit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, bi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, bi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, bi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, bi.Hint, bi.ActionLog)
	}
	return true
}
