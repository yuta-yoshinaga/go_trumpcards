package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PineappleWebInput パイナ���プルポーカーWebインプット
type PineappleWebInput struct {
	BaseWebInput
	Amount           int             `json:"amount,omitempty"`
	HumanPlayMs      int             `json:"humanPlayMs,omitempty"`
	CardIdx          *int            `json:"cardIdx,omitempty"`
	SmallBlind       *int            `json:"smallBlind,omitempty"`
	BigBlind         *int            `json:"bigBlind,omitempty"`
	TournamentMode   *bool           `json:"tournamentMode,omitempty"`
	BlindLevelHands  *int            `json:"blindLevelHands,omitempty"`
	BlindMultiplier  *int            `json:"blindMultiplier,omitempty"`
	BettingLimit     *int            `json:"bettingLimit,omitempty"`
	TableSize        *int            `json:"tableSize,omitempty"`
	RebuyEnabled     *bool           `json:"rebuyEnabled,omitempty"`
	RebuyMaxCount    *int            `json:"rebuyMaxCount,omitempty"`
	RebuyChips       *int            `json:"rebuyChips,omitempty"`
	RebuyPeriodHands *int            `json:"rebuyPeriodHands,omitempty"`
	AddonEnabled     *bool           `json:"addonEnabled,omitempty"`
	AddonChips       *int            `json:"addonChips,omitempty"`
	AddonAfterHand   *int            `json:"addonAfterHand,omitempty"`
	CpuMetaAI        bool            `json:"cpuMetaAI,omitempty"`
	Profile          json.RawMessage `json:"profile,omitempty"`
}

// PineappleWebOutput パイナップルポーカーWebアウトプット
type PineappleWebOutput struct {
	Players          []*HoldemWebOutputPlayer        `json:"players"`
	CommunityCards   []*WebOutputCard                `json:"communityCards"`
	Pot              int                             `json:"pot"`
	SidePots         []*HoldemWebOutputSidePot       `json:"sidePots"`
	DealerIdx        int                             `json:"dealerIdx"`
	CurrentTurn      int                             `json:"currentTurn"`
	Phase            int                             `json:"phase"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	LastBet          int                             `json:"lastBet"`
	MinRaise         int                             `json:"minRaise"`
	BettingLimit     int                             `json:"bettingLimit"`
	RaiseCount       int                             `json:"raiseCount"`
	MaxBetAmount     int                             `json:"maxBetAmount"`
	RoundResults     []*HoldemWebOutputResult        `json:"roundResults"`
	CpuActions       []*HoldemWebOutputCpuAction     `json:"cpuActions"`
	HandCount        int                             `json:"handCount"`
	SmallBlind       int                             `json:"smallBlind"`
	BigBlind         int                             `json:"bigBlind"`
	TournamentMode   bool                            `json:"tournamentMode"`
	BlindLevelHands  int                             `json:"blindLevelHands"`
	BlindMultiplier  int                             `json:"blindMultiplier"`
	TableSize        int                             `json:"tableSize"`
	RebuyAvailable   bool                            `json:"rebuyAvailable"`
	AddonAvailable   bool                            `json:"addonAvailable"`
	RebuyCounts      []int                           `json:"rebuyCounts"`
	AddonUsed        []bool                          `json:"addonUsed"`
	RebuyEnabled     bool                            `json:"rebuyEnabled"`
	AddonEnabled     bool                            `json:"addonEnabled"`
	RebuyMaxCount    int                             `json:"rebuyMaxCount"`
	RebuyChips       int                             `json:"rebuyChips"`
	AddonChips       int                             `json:"addonChips"`
	RebuyPeriodHands int                             `json:"rebuyPeriodHands"`
	AddonAfterHand   int                             `json:"addonAfterHand"`
	RebuyPhaseType   int                             `json:"rebuyPhaseType"`
	MuckAvailable    bool                            `json:"muckAvailable"`
	IsDiscardPhase   bool                            `json:"isDiscardPhase"`
	DiscardDone      []bool                          `json:"discardDone"`
	Equity           *HoldemWebOutputEquity          `json:"equity,omitempty"`
	PotOdds          *float64                        `json:"potOdds,omitempty"`
	MetaAI           *HoldemWebOutputMetaAI          `json:"metaAI,omitempty"`
	Profile          *domain.BettingHumanProfileData `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a PineappleConfig from the web input.
func (p PineappleWebInput) ToConfig() (domain.PineappleConfig, error) {
	cfg := domain.DefaultPineappleConfig()
	if err := validateAndApplyBlinds(&cfg.SmallBlind, &cfg.BigBlind, p.SmallBlind, p.BigBlind, cfg.BigBlind); err != nil {
		return domain.PineappleConfig{}, err
	}
	applyBool(&cfg.TournamentMode, p.TournamentMode)
	applyIntIfGte(&cfg.BlindLevelHands, p.BlindLevelHands, 1)
	applyIntIfGte(&cfg.BlindMultiplier, p.BlindMultiplier, 101)
	applyBettingLimit(&cfg.BettingLimit, p.BettingLimit)
	if p.TableSize != nil {
		ts := *p.TableSize
		if !domain.IsValidHoldemTableSize(ts) {
			return domain.PineappleConfig{}, errors.New("param error: tableSize must be 4, 6, or 9")
		}
		cfg.TableSize = ts
	}
	applyRebuyConfig(&cfg.RebuyEnabled, &cfg.RebuyMaxCount, &cfg.RebuyChips, &cfg.RebuyPeriodHands,
		p.RebuyEnabled, p.RebuyMaxCount, p.RebuyChips, p.RebuyPeriodHands)
	applyAddonConfig(&cfg.AddonEnabled, &cfg.AddonChips, &cfg.AddonAfterHand,
		p.AddonEnabled, p.AddonChips, p.AddonAfterHand)
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg, nil
}

// PineappleWebController パイナップルポーカーWebコントローラークラス
type PineappleWebController = GameWebController[usecase.PineappleInteractorIF, PineappleWebInput, *PineappleWebOutput]

// NewPineappleWebController and NewPineappleWebControllerWithProvider are
// the standard and provider-backed constructors for PineappleWebController.
var NewPineappleWebController, NewPineappleWebControllerWithProvider = webControllerPair[usecase.PineappleInteractorIF, PineappleWebInput, *PineappleWebOutput](
	newPineappleDefaultOutput, pineappleDispatch,
)

func newPineappleDefaultOutput(msg string) *PineappleWebOutput {
	return &PineappleWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func pineappleDispatch(bc *baseController, w http.ResponseWriter, pgi usecase.PineappleInteractorIF, param PineappleWebInput, newDefault func(string) *PineappleWebOutput) bool {
	if dispatchPokerAction(bc, w, pgi, param.Command, param.Amount, param.HumanPlayMs) {
		return true
	}
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, pgi.ResetWithConfig(cfg, param.Profile))
	case "d", "discard":
		if param.CardIdx == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: cardIdx is required for discard"))
			return true
		}
		bc.writePresenterResponse(w, pgi.Discard(*param.CardIdx))
	default:
		return dispatchLog(param.Command, bc, w, pgi.ActionLog)
	}
	return true
}
