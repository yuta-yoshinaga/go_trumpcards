//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KoenigrufenWebInput ケーニッヒルーフェン (Königrufen) のWebインプット
type KoenigrufenWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices 場札交換で捨てるカードのインデックス (6 枚)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Bid 入札 ("rufer")
	Bid *string `json:"bid,omitempty"`
	// CallSuit 呼ぶキングのスート (1..4)
	CallSuit *int `json:"callSuit,omitempty"`
	// Config ゲーム設定
	Config *KoenigrufenWebConfig `json:"config,omitempty"`
}

// KoenigrufenWebConfig ケーニッヒルーフェンのWeb設定
type KoenigrufenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// KoenigrufenWebOutputPlayer ケーニッヒルーフェンのWebアウトプットプレイヤー。
// IsPartner は partnerRevealed=true のときのみ真になり得る (秘密のパートナーを漏らさない)。
type KoenigrufenWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	CardPoints int              `json:"cardPoints"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
	IsPartner  bool             `json:"isPartner"`
}

// KoenigrufenWebOutputHint ヒント出力
type KoenigrufenWebOutputHint struct {
	Bid         *int   `json:"bid,omitempty"`
	CallSuit    *int   `json:"callSuit,omitempty"`
	CardIndices []int  `json:"cardIndices"`
	Reason      string `json:"reason"`
}

// KoenigrufenWebOutput ケーニッヒルーフェンのWebアウトプット。
// PartnerIdx は partnerRevealed=true になるまで常に -1 を出力し、秘密のパートナーを漏らさない。
type KoenigrufenWebOutput struct {
	Players          []*KoenigrufenWebOutputPlayer    `json:"players"`
	Phase            int                              `json:"phase"`
	RoundNumber      int                              `json:"roundNumber"`
	TrickNumber      int                              `json:"trickNumber"`
	CurrentPlayerIdx int                              `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                              `json:"leadPlayerIdx"`
	DealerIdx        int                              `json:"dealerIdx"`
	BidPlayerIdx     int                              `json:"bidPlayerIdx"`
	HighestBid       int                              `json:"highestBid"`
	HighestBidder    int                              `json:"highestBidder"`
	DeclarerIdx      int                              `json:"declarerIdx"`
	Contract         int                              `json:"contract"`
	CalledKing       int                              `json:"calledKing"`
	PartnerIdx       int                              `json:"partnerIdx"`
	PartnerRevealed  bool                             `json:"partnerRevealed"`
	TalonCount       int                              `json:"talonCount"`
	Talon            []*WebOutputCard                 `json:"talon"`
	StashOwner       int                              `json:"stashOwner"`
	CurrentTrick     []*WebOutputTrickCard            `json:"currentTrick"`
	PlayerScores     [domain.KoenigrufenPlayerCnt]int `json:"playerScores"`
	LastTrickWinner  int                              `json:"lastTrickWinner"`
	Outcome          int                              `json:"outcome"`
	Result           int                              `json:"result"`
	PlayableIndices  []int                            `json:"playableIndices"`
	GameEndFlag      bool                             `json:"gameEndFlag"`
	WinnerPlayer     int                              `json:"winnerPlayer"`
	IsHumanTurn      bool                             `json:"isHumanTurn"`
	IsHumanBidTurn   bool                             `json:"isHumanBidTurn"`
	IsHumanCall      bool                             `json:"isHumanCall"`
	IsHumanDiscard   bool                             `json:"isHumanDiscard"`
	Hint             *KoenigrufenWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config KoenigrufenWebOutputConfig `json:"config"`
}

// KoenigrufenWebOutputConfig ケーニッヒルーフェンの設定アウトプット
type KoenigrufenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig builds a KoenigrufenConfig from the nested web config, applying bounds checking.
func (c *KoenigrufenWebConfig) ToConfig() domain.KoenigrufenConfig {
	cfg := domain.DefaultKoenigrufenConfig()
	cfg.CpuDifficulty = domain.KoenigrufenCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.KoenigrufenCpuDifficultyEasy), int(domain.KoenigrufenCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals, 1, 1000)
	return cfg
}

// ToConfig builds a KoenigrufenConfig from the web input.
func (p KoenigrufenWebInput) ToConfig() domain.KoenigrufenConfig {
	return configOrDefault(p.Config, (*KoenigrufenWebConfig).ToConfig, domain.DefaultKoenigrufenConfig())
}

// koenigrufenParseBid 入札文字列を KoenigrufenBid に変換する (Pass=不正/パス)。
func koenigrufenParseBid(s string) domain.KoenigrufenBid {
	switch s {
	case "rufer", "r":
		return domain.KoenigrufenBidRufer
	default:
		return domain.KoenigrufenBidPass
	}
}

// KoenigrufenWebController ケーニッヒルーフェンのWebコントローラークラス
type KoenigrufenWebController = GameWebController[usecase.KoenigrufenInteractorIF, KoenigrufenWebInput, *KoenigrufenWebOutput]

// NewKoenigrufenWebController and NewKoenigrufenWebControllerWithProvider are
// the standard and provider-backed constructors for KoenigrufenWebController.
var NewKoenigrufenWebController, NewKoenigrufenWebControllerWithProvider = webControllerPair[usecase.KoenigrufenInteractorIF, KoenigrufenWebInput, *KoenigrufenWebOutput](
	newKoenigrufenDefaultOutput, koenigrufenDispatch,
)

func newKoenigrufenDefaultOutput(msg string) *KoenigrufenWebOutput {
	return &KoenigrufenWebOutput{
		Players:         make([]*KoenigrufenWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		Talon:           make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		HighestBidder:   -1,
		CalledKing:      -1,
		PartnerIdx:      -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func koenigrufenDispatch(bc *baseController, w http.ResponseWriter, di usecase.KoenigrufenInteractorIF, param KoenigrufenWebInput, newDefault func(string) *KoenigrufenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(koenigrufenParseBid(*param.Bid)))
	case "pass":
		bc.writePresenterResponse(w, di.Pass())
	case "ck", "callking":
		if !requireParam(bc, w, newDefault, param.CallSuit == nil, "param error: callSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.CallKing(*param.CallSuit))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndices == nil, "param error: cardIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(param.CardIndices))
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
