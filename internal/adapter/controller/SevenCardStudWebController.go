package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevenCardStudWebInput セブンカードスタッドWebインプット
type SevenCardStudWebInput struct {
	BaseWebInput
	Amount           int             `json:"amount,omitempty"`
	HumanPlayMs      int             `json:"humanPlayMs,omitempty"`
	Ante             *int            `json:"ante,omitempty"`
	BringIn          *int            `json:"bringIn,omitempty"`
	SmallBet         *int            `json:"smallBet,omitempty"`
	BigBet           *int            `json:"bigBet,omitempty"`
	TournamentMode   *bool           `json:"tournamentMode,omitempty"`
	AnteLevelHands   *int            `json:"anteLevelHands,omitempty"`
	AnteMultiplier   *int            `json:"anteMultiplier,omitempty"`
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

// SevenCardStudWebOutputPlayer セブンカードスタッドWebアウトプットプレイヤー
type SevenCardStudWebOutputPlayer struct {
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

// SevenCardStudWebOutputCpuAction セブンカードスタッドCPU行動記録
type SevenCardStudWebOutputCpuAction struct {
	PlayerIdx int `json:"playerIdx"`
	Action    int `json:"action"`
	Amount    int `json:"amount"`
}

// SevenCardStudWebOutputResult セブンカードスタッドショーダウン結果
type SevenCardStudWebOutputResult struct {
	PlayerIdx int              `json:"playerIdx"`
	HandRank  int              `json:"handRank"`
	HandName  string           `json:"handName"`
	Kickers   string           `json:"kickers"`
	BestHand  []*WebOutputCard `json:"bestHand"`
	WonAmount int              `json:"wonAmount"`
	Mucked    bool             `json:"mucked"`
}

// SevenCardStudWebOutputSidePot セブンカードスタッドサイドポット
type SevenCardStudWebOutputSidePot struct {
	Amount          int   `json:"amount"`
	EligiblePlayers []int `json:"eligiblePlayers"`
}

// SevenCardStudWebOutputMetaAI メタAI情報
type SevenCardStudWebOutputMetaAI struct {
	Enabled        bool    `json:"enabled"`
	GamesPlayed    int     `json:"gamesPlayed"`
	BluffRate      float64 `json:"bluffRate"`
	FoldRate       float64 `json:"foldRate"`
	HesitationMean float64 `json:"hesitationMean"`
}

// SevenCardStudWebOutput セブンカードスタッドWebアウトプット
type SevenCardStudWebOutput struct {
	Players          []*SevenCardStudWebOutputPlayer    `json:"players"`
	CommunityCard    *WebOutputCard                     `json:"communityCard"`
	Pot              int                                `json:"pot"`
	SidePots         []*SevenCardStudWebOutputSidePot   `json:"sidePots"`
	DealerIdx        int                                `json:"dealerIdx"`
	CurrentTurn      int                                `json:"currentTurn"`
	Phase            int                                `json:"phase"`
	GameEndFlag      bool                               `json:"gameEndFlag"`
	LastBet          int                                `json:"lastBet"`
	MinRaise         int                                `json:"minRaise"`
	BettingLimit     int                                `json:"bettingLimit"`
	RaiseCount       int                                `json:"raiseCount"`
	MaxBetAmount     int                                `json:"maxBetAmount"`
	RoundResults     []*SevenCardStudWebOutputResult    `json:"roundResults"`
	CpuActions       []*SevenCardStudWebOutputCpuAction `json:"cpuActions"`
	HandCount        int                                `json:"handCount"`
	Ante             int                                `json:"ante"`
	BringIn          int                                `json:"bringIn"`
	SmallBet         int                                `json:"smallBet"`
	BigBet           int                                `json:"bigBet"`
	TournamentMode   bool                               `json:"tournamentMode"`
	AnteLevelHands   int                                `json:"anteLevelHands"`
	AnteMultiplier   int                                `json:"anteMultiplier"`
	TableSize        int                                `json:"tableSize"`
	BringInPlayerIdx int                                `json:"bringInPlayerIdx"`
	RebuyAvailable   bool                               `json:"rebuyAvailable"`
	AddonAvailable   bool                               `json:"addonAvailable"`
	RebuyCounts      []int                              `json:"rebuyCounts"`
	AddonUsed        []bool                             `json:"addonUsed"`
	RebuyEnabled     bool                               `json:"rebuyEnabled"`
	AddonEnabled     bool                               `json:"addonEnabled"`
	RebuyMaxCount    int                                `json:"rebuyMaxCount"`
	RebuyChips       int                                `json:"rebuyChips"`
	AddonChips       int                                `json:"addonChips"`
	RebuyPeriodHands int                                `json:"rebuyPeriodHands"`
	AddonAfterHand   int                                `json:"addonAfterHand"`
	RebuyPhaseType   int                                `json:"rebuyPhaseType"`
	MuckAvailable    bool                               `json:"muckAvailable"`
	MetaAI           *SevenCardStudWebOutputMetaAI      `json:"metaAI,omitempty"`
	Profile          *domain.BettingHumanProfileData    `json:"profile,omitempty"`
	WebOutputBase
}

// ToConfig builds a SevenCardStudConfig from the web input.
// Returns an error if the table size is invalid.
func (p SevenCardStudWebInput) ToConfig() (domain.SevenCardStudConfig, error) {
	cfg := domain.DefaultSevenCardStudConfig()
	if p.Ante != nil && *p.Ante >= 1 {
		cfg.Ante = *p.Ante
	}
	if p.BringIn != nil && *p.BringIn >= 1 {
		cfg.BringIn = *p.BringIn
	}
	if p.SmallBet != nil && *p.SmallBet >= 1 {
		cfg.SmallBet = *p.SmallBet
	}
	if p.BigBet != nil && *p.BigBet >= 1 {
		cfg.BigBet = *p.BigBet
	}
	if p.TournamentMode != nil {
		cfg.TournamentMode = *p.TournamentMode
	}
	if p.AnteLevelHands != nil && *p.AnteLevelHands >= 1 {
		cfg.AnteLevelHands = *p.AnteLevelHands
	}
	if p.AnteMultiplier != nil && *p.AnteMultiplier >= 101 {
		cfg.AnteMultiplier = *p.AnteMultiplier
	}
	if p.BettingLimit != nil {
		bl := *p.BettingLimit
		if bl < 0 {
			bl = 0
		} else if bl > 2 {
			bl = 2
		}
		cfg.BettingLimit = domain.BettingLimitType(bl)
	}
	if p.TableSize != nil {
		ts := *p.TableSize
		if !domain.IsValidSevenCardStudTableSize(ts) {
			return domain.SevenCardStudConfig{}, errors.New("param error: tableSize must be 2-7")
		}
		cfg.TableSize = ts
	}
	if p.RebuyEnabled != nil {
		cfg.RebuyEnabled = *p.RebuyEnabled
	}
	if p.RebuyMaxCount != nil && *p.RebuyMaxCount >= 1 {
		cfg.RebuyMaxCount = *p.RebuyMaxCount
	}
	if p.RebuyChips != nil && *p.RebuyChips >= 1 {
		cfg.RebuyChips = *p.RebuyChips
	}
	if p.RebuyPeriodHands != nil && *p.RebuyPeriodHands >= 1 {
		cfg.RebuyPeriodHands = *p.RebuyPeriodHands
	}
	if p.AddonEnabled != nil {
		cfg.AddonEnabled = *p.AddonEnabled
	}
	if p.AddonChips != nil && *p.AddonChips >= 1 {
		cfg.AddonChips = *p.AddonChips
	}
	if p.AddonAfterHand != nil && *p.AddonAfterHand >= 1 {
		cfg.AddonAfterHand = *p.AddonAfterHand
	}
	cfg.CpuMetaAI = p.CpuMetaAI
	return cfg, nil
}

// SevenCardStudWebController セブンカードスタッドWebコントローラークラス
type SevenCardStudWebController = GameWebController[usecase.SevenCardStudInteractorIF, SevenCardStudWebInput, *SevenCardStudWebOutput]

// NewSevenCardStudWebController and NewSevenCardStudWebControllerWithProvider are
// the standard and provider-backed constructors for SevenCardStudWebController.
var NewSevenCardStudWebController, NewSevenCardStudWebControllerWithProvider = webControllerPair[usecase.SevenCardStudInteractorIF, SevenCardStudWebInput, *SevenCardStudWebOutput](
	newSevenCardStudDefaultOutput, sevenCardStudDispatch,
)

func newSevenCardStudDefaultOutput(msg string) *SevenCardStudWebOutput {
	return &SevenCardStudWebOutput{
		Players:       make([]*SevenCardStudWebOutputPlayer, 0),
		SidePots:      make([]*SevenCardStudWebOutputSidePot, 0),
		RoundResults:  make([]*SevenCardStudWebOutputResult, 0),
		CpuActions:    make([]*SevenCardStudWebOutputCpuAction, 0),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sevenCardStudDispatch(bc *baseController, w http.ResponseWriter, sgi usecase.SevenCardStudInteractorIF, param SevenCardStudWebInput, newDefault func(string) *SevenCardStudWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg, err := param.ToConfig()
		if err != nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault(err.Error()))
			return true
		}
		bc.writePresenterResponse(w, sgi.ResetWithConfig(cfg, param.Profile))
	case "f", "fold":
		bc.writePresenterResponse(w, sgi.Action(domain.SevenCardStudActionFold, 0, param.HumanPlayMs))
	case "ck", "check":
		bc.writePresenterResponse(w, sgi.Action(domain.SevenCardStudActionCheck, 0, param.HumanPlayMs))
	case "c", "call":
		bc.writePresenterResponse(w, sgi.Action(domain.SevenCardStudActionCall, 0, param.HumanPlayMs))
	case "b", "bet":
		bc.writePresenterResponse(w, sgi.Action(domain.SevenCardStudActionBet, param.Amount, param.HumanPlayMs))
	case "ra", "raise":
		bc.writePresenterResponse(w, sgi.Action(domain.SevenCardStudActionRaise, param.Amount, param.HumanPlayMs))
	case "a", "allin":
		bc.writePresenterResponse(w, sgi.Action(domain.SevenCardStudActionAllIn, 0, param.HumanPlayMs))
	case "rb", "rebuy":
		bc.writePresenterResponse(w, sgi.Rebuy())
	case "sr", "skiprebuy":
		bc.writePresenterResponse(w, sgi.SkipRebuy())
	case "ad", "addon":
		bc.writePresenterResponse(w, sgi.Addon())
	case "sa", "skipaddon":
		bc.writePresenterResponse(w, sgi.SkipAddon())
	case "m", "muck":
		bc.writePresenterResponse(w, sgi.Muck())
	case "sh", "show":
		bc.writePresenterResponse(w, sgi.ShowHand())
	default:
		return dispatchLog(param.Command, bc, w, sgi.ActionLog)
	}
	return true
}
