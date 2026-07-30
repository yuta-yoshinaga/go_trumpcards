//go:build !js || !wasm || casino

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevenCardStudWebInput セブンカードスタッドWebインプット
type SevenCardStudWebInput struct {
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
	// LowQualifies は 8-or-better のローが成立したか (Hi-Lo のみ)。
	LowQualifies bool `json:"lowQualifies,omitempty"`
	// LowBestHand はローのベスト5枚 (Hi-Lo のみ)。
	LowBestHand []*WebOutputCard `json:"lowBestHand,omitempty"`
	// WonLow はローとして獲得したチップ (Hi-Lo のみ)。WonAmount はハイとローの合計。
	WonLow int `json:"wonLow,omitempty"`
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
	Players       []*SevenCardStudWebOutputPlayer  `json:"players"`
	CommunityCard *WebOutputCard                   `json:"communityCard"`
	Pot           int                              `json:"pot"`
	SidePots      []*SevenCardStudWebOutputSidePot `json:"sidePots"`
	DealerIdx     int                              `json:"dealerIdx"`
	CurrentTurn   int                              `json:"currentTurn"`
	Phase         int                              `json:"phase"`
	// IsHiLo は 8-or-better のスプリットかどうか。ページがルート名から推測
	// しなくて済むように送る。
	IsHiLo           bool                               `json:"isHiLo,omitempty"`
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
	applyIntIfGte(&cfg.Ante, p.Ante, 1)
	applyIntIfGte(&cfg.BringIn, p.BringIn, 1)
	applyIntIfGte(&cfg.SmallBet, p.SmallBet, 1)
	applyIntIfGte(&cfg.BigBet, p.BigBet, 1)
	applyBool(&cfg.TournamentMode, p.TournamentMode)
	applyIntIfGte(&cfg.AnteLevelHands, p.AnteLevelHands, 1)
	applyIntIfGte(&cfg.AnteMultiplier, p.AnteMultiplier, 101)
	applyBettingLimit(&cfg.BettingLimit, p.BettingLimit)
	if err := applyTableSize(&cfg.TableSize, p.TableSize, domain.IsValidSevenCardStudTableSize, "param error: tableSize must be 2-7"); err != nil {
		return domain.SevenCardStudConfig{}, err
	}
	applyRebuyConfig(&cfg.RebuyEnabled, &cfg.RebuyMaxCount, &cfg.RebuyChips, &cfg.RebuyPeriodHands,
		p.RebuyEnabled, p.RebuyMaxCount, p.RebuyChips, p.RebuyPeriodHands)
	applyAddonConfig(&cfg.AddonEnabled, &cfg.AddonChips, &cfg.AddonAfterHand,
		p.AddonEnabled, p.AddonChips, p.AddonAfterHand)
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
