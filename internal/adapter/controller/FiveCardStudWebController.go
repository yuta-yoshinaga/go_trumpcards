//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FiveCardStudWebInput ファイブカードスタッドWebインプット
type FiveCardStudWebInput struct {
	BaseWebInput
	PokerCommonInput
	Amount         int             `json:"amount,omitempty"`
	HumanPlayMs    int             `json:"humanPlayMs,omitempty"`
	Ante           *int            `json:"ante,omitempty"`
	BringIn        *int            `json:"bringIn,omitempty"`
	SmallBet       *int            `json:"smallBet,omitempty"`
	BigBet         *int            `json:"bigBet,omitempty"`
	AnteLevelHands *int            `json:"anteLevelHands,omitempty"`
	AnteMultiplier *int            `json:"anteMultiplier,omitempty"`
	CpuMetaAI      bool            `json:"cpuMetaAI,omitempty"`
	Profile        json.RawMessage `json:"profile,omitempty"`
}

// FiveCardStudWebOutputPlayer ファイブカードスタッドWebアウトプットプレイヤー
type FiveCardStudWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	HoleCards     []*WebOutputCard `json:"holeCards"`
	DoorCards     []*WebOutputCard `json:"doorCards"`
	Chips         int              `json:"chips"`
	CurrentBet    int              `json:"currentBet"`
	Folded        bool             `json:"folded"`
	AllIn         bool             `json:"allIn"`
	HandRank      int              `json:"handRank"`
	HandName      string           `json:"handName"`
	BestHand      []*WebOutputCard `json:"bestHand"`
	PlayStyleName string           `json:"playStyleName"`
	TotalHands    int              `json:"totalHands"`
	VPIP          int              `json:"vpip"`
	PFR           int              `json:"pfr"`
	ThreeBet      int              `json:"threeBet"`
	AF            string           `json:"af"`
}

// FiveCardStudWebOutputCpuAction ファイブカードスタッドCPU行動記録
type FiveCardStudWebOutputCpuAction struct {
	PlayerIdx int `json:"playerIdx"`
	Action    int `json:"action"`
	Amount    int `json:"amount"`
}

// FiveCardStudWebOutputResult ファイブカードスタッドショーダウン結果
type FiveCardStudWebOutputResult struct {
	PlayerIdx int              `json:"playerIdx"`
	HandRank  int              `json:"handRank"`
	HandName  string           `json:"handName"`
	Kickers   string           `json:"kickers"`
	BestHand  []*WebOutputCard `json:"bestHand"`
	WonAmount int              `json:"wonAmount"`
	Mucked    bool             `json:"mucked"`
}

// FiveCardStudWebOutputSidePot ファイブカードスタッドサイドポット
type FiveCardStudWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// FiveCardStudWebOutputMetaAI メタAI情報
type FiveCardStudWebOutputMetaAI struct {
	Enabled        bool    `json:"enabled"`
	GamesPlayed    int     `json:"gamesPlayed"`
	BluffRate      float64 `json:"bluffRate"`
	FoldRate       float64 `json:"foldRate"`
	HesitationMean float64 `json:"hesitationMean"`
}

// FiveCardStudWebOutput ファイブカードスタッドWebアウトプット
type FiveCardStudWebOutput struct {
	Players          []*FiveCardStudWebOutputPlayer    `json:"players"`
	CommunityCard    *WebOutputCard                    `json:"communityCard"`
	Pot              int                               `json:"pot"`
	SidePots         []*FiveCardStudWebOutputSidePot   `json:"sidePots"`
	DealerIdx        int                               `json:"dealerIdx"`
	CurrentTurn      int                               `json:"currentTurn"`
	Phase            int                               `json:"phase"`
	GameEndFlag      bool                              `json:"gameEndFlag"`
	LastBet          int                               `json:"lastBet"`
	MinRaise         int                               `json:"minRaise"`
	BettingLimit     int                               `json:"bettingLimit"`
	RaiseCount       int                               `json:"raiseCount"`
	MaxBetAmount     int                               `json:"maxBetAmount"`
	RoundResults     []*FiveCardStudWebOutputResult    `json:"roundResults"`
	CpuActions       []*FiveCardStudWebOutputCpuAction `json:"cpuActions"`
	HandCount        int                               `json:"handCount"`
	Ante             int                               `json:"ante"`
	BringIn          int                               `json:"bringIn"`
	SmallBet         int                               `json:"smallBet"`
	BigBet           int                               `json:"bigBet"`
	TournamentMode   bool                              `json:"tournamentMode"`
	AnteLevelHands   int                               `json:"anteLevelHands"`
	AnteMultiplier   int                               `json:"anteMultiplier"`
	TableSize        int                               `json:"tableSize"`
	BringInPlayerIdx int                               `json:"bringInPlayerIdx"`
	RebuyAvailable   bool                              `json:"rebuyAvailable"`
	AddonAvailable   bool                              `json:"addonAvailable"`
	RebuyCounts      []int                             `json:"rebuyCounts"`
	AddonUsed        []bool                            `json:"addonUsed"`
	RebuyEnabled     bool                              `json:"rebuyEnabled"`
	AddonEnabled     bool                              `json:"addonEnabled"`
	RebuyMaxCount    int                               `json:"rebuyMaxCount"`
	RebuyChips       int                               `json:"rebuyChips"`
	AddonChips       int                               `json:"addonChips"`
	RebuyPeriodHands int                               `json:"rebuyPeriodHands"`
	AddonAfterHand   int                               `json:"addonAfterHand"`
	RebuyPhaseType   int                               `json:"rebuyPhaseType"`
	MuckAvailable    bool                              `json:"muckAvailable"`
	MetaAI           *FiveCardStudWebOutputMetaAI      `json:"metaAI,omitempty"`
	Profile          *domain.BettingHumanProfileData   `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a FiveCardStudConfig from the web input.
// Returns an error if the table size is invalid.
func (p FiveCardStudWebInput) ToConfig() (domain.FiveCardStudConfig, error) {
	cfg := domain.DefaultFiveCardStudConfig()
	applyIntIfGte(&cfg.Ante, p.Ante, 1)
	applyIntIfGte(&cfg.BringIn, p.BringIn, 1)
	applyIntIfGte(&cfg.SmallBet, p.SmallBet, 1)
	applyIntIfGte(&cfg.BigBet, p.BigBet, 1)
	applyBool(&cfg.TournamentMode, p.TournamentMode)
	applyIntIfGte(&cfg.AnteLevelHands, p.AnteLevelHands, 1)
	applyIntIfGte(&cfg.AnteMultiplier, p.AnteMultiplier, 101)
	applyBettingLimit(&cfg.BettingLimit, p.BettingLimit)
	if err := applyTableSize(&cfg.TableSize, p.TableSize, domain.IsValidFiveCardStudTableSize, "param error: tableSize must be 2-6"); err != nil {
		return domain.FiveCardStudConfig{}, err
	}
	applyRebuyConfig(&cfg.RebuyEnabled, &cfg.RebuyMaxCount, &cfg.RebuyChips, &cfg.RebuyPeriodHands,
		p.RebuyEnabled, p.RebuyMaxCount, p.RebuyChips, p.RebuyPeriodHands)
	applyAddonConfig(&cfg.AddonEnabled, &cfg.AddonChips, &cfg.AddonAfterHand,
		p.AddonEnabled, p.AddonChips, p.AddonAfterHand)
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg, nil
}

// FiveCardStudWebController ファイブカードスタッドWebコントローラークラス
type FiveCardStudWebController = GameWebController[usecase.FiveCardStudInteractorIF, FiveCardStudWebInput, *FiveCardStudWebOutput]

// NewFiveCardStudWebController and NewFiveCardStudWebControllerWithProvider are
// the standard and provider-backed constructors for FiveCardStudWebController.
var NewFiveCardStudWebController, NewFiveCardStudWebControllerWithProvider = webControllerPair[usecase.FiveCardStudInteractorIF, FiveCardStudWebInput, *FiveCardStudWebOutput](
	newFiveCardStudDefaultOutput, fiveCardStudDispatch,
)

func newFiveCardStudDefaultOutput(msg string) *FiveCardStudWebOutput {
	return &FiveCardStudWebOutput{
		Players:       make([]*FiveCardStudWebOutputPlayer, 0),
		SidePots:      make([]*FiveCardStudWebOutputSidePot, 0),
		RoundResults:  make([]*FiveCardStudWebOutputResult, 0),
		CpuActions:    make([]*FiveCardStudWebOutputCpuAction, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func fiveCardStudDispatch(bc *baseController, w http.ResponseWriter, sgi usecase.FiveCardStudInteractorIF, param FiveCardStudWebInput, newDefault func(string) *FiveCardStudWebOutput) bool {
	if dispatchPokerAction(bc, w, sgi, param.Command, param.Amount, param.HumanPlayMs) {
		return true
	}
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, sgi.ResetWithConfig(cfg, param.Profile))
	default:
		return dispatchLog(param.Command, bc, w, sgi.ActionLog)
	}
	return true
}
