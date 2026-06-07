//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HoldemWebInput テキサスホールデムWebインプット
type HoldemWebInput struct {
	BaseWebInput
	PokerCommonInput
	PokerBlindsInput
	Amount      int             `json:"amount,omitempty"`
	HumanPlayMs int             `json:"humanPlayMs,omitempty"`
	CpuMetaAI   bool            `json:"cpuMetaAI,omitempty"`
	Profile     json.RawMessage `json:"profile,omitempty"`
}

// HoldemWebOutputPlayer テキサスホールデムWebアウトプットプレイヤー
//
// Hi-Lo (Omaha 8 or Better) では、ショーダウン時に LowBestHand /
// LowQualifies が populated される。それ以外のゲームでは omitempty に
// より JSON 出力に含まれない。
type HoldemWebOutputPlayer struct {
	ID            int              `json:"id"`
	IsHuman       bool             `json:"isHuman"`
	Cards         []*WebOutputCard `json:"cards"`
	Chips         int              `json:"chips"`
	CurrentBet    int              `json:"currentBet"`
	Folded        bool             `json:"folded"`
	AllIn         bool             `json:"allIn"`
	HandRank      int              `json:"handRank"`
	HandName      string           `json:"handName"`
	BestHand      []*WebOutputCard `json:"bestHand"`
	LowBestHand   []*WebOutputCard `json:"lowBestHand,omitempty"`
	LowQualifies  bool             `json:"lowQualifies,omitempty"`
	PlayStyleName string           `json:"playStyleName"`
	TotalHands    int              `json:"totalHands"`
	VPIP          int              `json:"vpip"`
	PFR           int              `json:"pfr"`
	ThreeBet      int              `json:"threeBet"`
	AF            string           `json:"af"`
}

// HoldemWebOutputCpuAction テキサスホールデムCPU行動記録
type HoldemWebOutputCpuAction struct {
	PlayerIdx int `json:"playerIdx"`
	Action    int `json:"action"`
	Amount    int `json:"amount"`
}

// HoldemWebOutputResult テキサスホールデムショーダウン結果
//
// Hi-Lo (Omaha 8 or Better) ではポット分割により Hi/Lo 内訳と
// LowBestHand / LowKickers / LowQualifies が populated される。
// Hi-Lo 以外のゲームでは omitempty で JSON 出力に含まれない。
type HoldemWebOutputResult struct {
	PlayerIdx    int              `json:"playerIdx"`
	HandRank     int              `json:"handRank"`
	HandName     string           `json:"handName"`
	Kickers      string           `json:"kickers"`
	BestHand     []*WebOutputCard `json:"bestHand"`
	WonAmount    int              `json:"wonAmount"`
	Mucked       bool             `json:"mucked"`
	LowBestHand  []*WebOutputCard `json:"lowBestHand,omitempty"`
	LowKickers   string           `json:"lowKickers,omitempty"`
	LowQualifies bool             `json:"lowQualifies,omitempty"`
	HiWonAmount  int              `json:"hiWonAmount,omitempty"`
	LowWonAmount int              `json:"lowWonAmount,omitempty"`
}

// HoldemWebOutputSidePot テキサスホールデムサイドポット
type HoldemWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// HoldemWebOutputEquity テキサスホールデムエクイティ情報
type HoldemWebOutputEquity struct {
	WinProbability float64                    `json:"winProbability"`
	HandOdds       []*HoldemWebOutputHandOdds `json:"handOdds"`
}

// HoldemWebOutputHandOdds テキサスホールデムハンドオッズ
type HoldemWebOutputHandOdds struct {
	HandRank    int     `json:"handRank"`
	HandName    string  `json:"handName"`
	Probability float64 `json:"probability"`
}

// HoldemWebOutputMetaAI メタAI情報
type HoldemWebOutputMetaAI struct {
	Enabled        bool    `json:"enabled"`
	GamesPlayed    int     `json:"gamesPlayed"`
	BluffRate      float64 `json:"bluffRate"`
	FoldRate       float64 `json:"foldRate"`
	HesitationMean float64 `json:"hesitationMean"`
}

// HoldemWebOutput テキサスホールデムWebアウトプット
type HoldemWebOutput struct {
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
	IsHiLo           bool                            `json:"isHiLo,omitempty"`
	Equity           *HoldemWebOutputEquity          `json:"equity,omitempty"`
	PotOdds          *float64                        `json:"potOdds,omitempty"`
	MetaAI           *HoldemWebOutputMetaAI          `json:"metaAI,omitempty"`
	Profile          *domain.BettingHumanProfileData `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a HoldemConfig from the web input.
// Returns an error if the blind configuration is invalid or tableSize is invalid.
func (p HoldemWebInput) ToConfig() (domain.HoldemConfig, error) {
	cfg := domain.DefaultHoldemConfig()
	if err := validateAndApplyBlinds(&cfg.SmallBlind, &cfg.BigBlind, p.SmallBlind, p.BigBlind, cfg.BigBlind); err != nil {
		return domain.HoldemConfig{}, err
	}
	applyBool(&cfg.TournamentMode, p.TournamentMode)
	applyIntIfGte(&cfg.BlindLevelHands, p.BlindLevelHands, 1)
	applyIntIfGte(&cfg.BlindMultiplier, p.BlindMultiplier, 101)
	applyBettingLimit(&cfg.BettingLimit, p.BettingLimit)
	if err := applyTableSize(&cfg.TableSize, p.TableSize, domain.IsValidHoldemTableSize, "param error: tableSize must be 4, 6, or 9"); err != nil {
		return domain.HoldemConfig{}, err
	}
	applyRebuyConfig(&cfg.RebuyEnabled, &cfg.RebuyMaxCount, &cfg.RebuyChips, &cfg.RebuyPeriodHands,
		p.RebuyEnabled, p.RebuyMaxCount, p.RebuyChips, p.RebuyPeriodHands)
	applyAddonConfig(&cfg.AddonEnabled, &cfg.AddonChips, &cfg.AddonAfterHand,
		p.AddonEnabled, p.AddonChips, p.AddonAfterHand)
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg, nil
}

// HoldemWebController テキサスホールデムWebコントローラークラス
type HoldemWebController = GameWebController[usecase.HoldemInteractorIF, HoldemWebInput, *HoldemWebOutput]

// NewHoldemWebController and NewHoldemWebControllerWithProvider are
// the standard and provider-backed constructors for HoldemWebController.
var NewHoldemWebController, NewHoldemWebControllerWithProvider = webControllerPair[usecase.HoldemInteractorIF, HoldemWebInput, *HoldemWebOutput](
	newHoldemDefaultOutput, holdemDispatch,
)

func newHoldemDefaultOutput(msg string) *HoldemWebOutput {
	return &HoldemWebOutput{
		Players:        make([]*HoldemWebOutputPlayer, 0),
		CommunityCards: make([]*WebOutputCard, 0),
		SidePots:       make([]*HoldemWebOutputSidePot, 0),
		RoundResults:   make([]*HoldemWebOutputResult, 0),
		CpuActions:     make([]*HoldemWebOutputCpuAction, 0),
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func holdemDispatch(bc *baseController, w http.ResponseWriter, hgi usecase.HoldemInteractorIF, param HoldemWebInput, newDefault func(string) *HoldemWebOutput) bool {
	if dispatchPokerAction(bc, w, hgi, param.Command, param.Amount, param.HumanPlayMs) {
		return true
	}
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, hgi.ResetWithConfig(cfg, param.Profile))
	default:
		return dispatchLog(param.Command, bc, w, hgi.ActionLog)
	}
	return true
}
