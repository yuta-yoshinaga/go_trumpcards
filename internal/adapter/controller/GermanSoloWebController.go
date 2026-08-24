//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GermanSoloWebInput ジャーマン・ソロ (GermanSolo) のWebインプット
type GermanSoloWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Bid ビッド宣言 (0=pass, 2=frage, 3=solo, 4=tout。1=Mussfrage は卓が押し付ける契約なので宣言できない)
	Bid *int `json:"bid,omitempty"`
	// TrumpSuit ビッド勝利時に選ぶ切り札スート (1=spade, 2=club, 3=heart, 4=diamond)
	TrumpSuit *int `json:"trumpSuit,omitempty"`
	// AceSuit はエース呼びフェーズで指名するエースのスート。
	AceSuit *int `json:"aceSuit,omitempty"`
	// Config ゲーム設定
	Config *GermanSoloWebConfig `json:"config,omitempty"`
}

// GermanSoloWebConfig ジャーマン・ソロのWeb設定
type GermanSoloWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// GermanSoloWebOutputPlayer ジャーマン・ソロのWebアウトプットプレイヤー
type GermanSoloWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// GermanSoloWebOutput ジャーマン・ソロのWebアウトプット
type GermanSoloWebOutput struct {
	Players          []*GermanSoloWebOutputPlayer    `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	CurrentBidderIdx int                             `json:"currentBidderIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	DealerIdx        int                             `json:"dealerIdx"`
	ForehandIdx      int                             `json:"forehandIdx"`
	DeclarerIdx      int                             `json:"declarerIdx"`
	WinningBid       int                             `json:"winningBid"`
	TrumpSuit        int                             `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	PlayerScores     [domain.GermanSoloPlayerCnt]int `json:"playerScores"`
	LastTrickWinner  int                             `json:"lastTrickWinner"`
	Outcome          int                             `json:"outcome"`
	Result           int                             `json:"result"`
	PlayableIndices  []int                           `json:"playableIndices"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerPlayer     int                             `json:"winnerPlayer"`
	IsHumanTurn      bool                            `json:"isHumanTurn"`
	IsHumanBidTurn   bool                            `json:"isHumanBidTurn"`
	// HighestBid は競り中に上回るべき宣言、BiddableBids は今宣言できるビッド。
	// RequiredTricks は確定した契約の必要トリック数 (Tout=8, それ以外=5)。
	HighestBid     int   `json:"highestBid"`
	BiddableBids   []int `json:"biddableBids"`
	RequiredTricks int   `json:"requiredTricks"`
	DeclarerTricks int   `json:"declarerTricks"`
	DefenderTricks int   `json:"defenderTricks"`
	// エース呼び。CalledAceSuit は呼ばれたエースのスート (-1=未指名)、PartnerIdx は
	// **そのエースが場に出るまで -1** (誰が味方かは伏せる)。PlaysAlone は単独契約
	// (Solo / Tout) と、呼べるエースが無かった Frage。
	IsHumanAceCallTurn bool               `json:"isHumanAceCallTurn"`
	CalledAceSuit      int                `json:"calledAceSuit"`
	CallableAceSuits   []int              `json:"callableAceSuits"`
	PartnerIdx         int                `json:"partnerIdx"`
	PlaysAlone         bool               `json:"playsAlone"`
	Hint               *WebOutputCardHint `json:"hint,omitempty"`
	// HintAceSuit はエース呼びフェーズでヒントが勧めるスート (0=なし)。共有の
	// WebOutputCardHint は札の索引しか運べないので、スートはここに載せる。
	HintAceSuit int `json:"hintAceSuit"`
	WebOutputBase
	Config GermanSoloWebOutputConfig `json:"config"`
}

// GermanSoloWebOutputConfig ジャーマン・ソロの設定アウトプット
type GermanSoloWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds an GermanSoloConfig from the nested web config, applying bounds checking.
func (c *GermanSoloWebConfig) ToConfig() domain.GermanSoloConfig {
	cfg := domain.DefaultGermanSoloConfig()
	cfg.CpuDifficulty = domain.GermanSoloCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.GermanSoloCpuDifficultyEasy), int(domain.GermanSoloCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 1000)
	return cfg
}

// ToConfig builds an GermanSoloConfig from the web input.
func (p GermanSoloWebInput) ToConfig() domain.GermanSoloConfig {
	return configOrDefault(p.Config, (*GermanSoloWebConfig).ToConfig, domain.DefaultGermanSoloConfig())
}

// GermanSoloWebController ジャーマン・ソロのWebコントローラークラス
type GermanSoloWebController = GameWebController[usecase.GermanSoloInteractorIF, GermanSoloWebInput, *GermanSoloWebOutput]

// NewGermanSoloWebController and NewGermanSoloWebControllerWithProvider are
// the standard and provider-backed constructors for GermanSoloWebController.
var NewGermanSoloWebController, NewGermanSoloWebControllerWithProvider = webControllerPair[usecase.GermanSoloInteractorIF, GermanSoloWebInput, *GermanSoloWebOutput](
	newGermanSoloDefaultOutput, germanSoloDispatch,
)

func newGermanSoloDefaultOutput(msg string) *GermanSoloWebOutput {
	return &GermanSoloWebOutput{
		Players:          make([]*GermanSoloWebOutputPlayer, 0),
		CurrentTrick:     make([]*WebOutputTrickCard, 0),
		PlayableIndices:  make([]int, 0),
		BiddableBids:     make([]int, 0),
		CallableAceSuits: make([]int, 0),
		CalledAceSuit:    -1,
		PartnerIdx:       -1,
		DeclarerIdx:      -1,
		TrumpSuit:        -1,
		LastTrickWinner:  -1,
		WinnerPlayer:     -1,
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

func germanSoloDispatch(bc *baseController, w http.ResponseWriter, di usecase.GermanSoloInteractorIF, param GermanSoloWebInput, newDefault func(string) *GermanSoloWebOutput) bool {
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
		bc.writePresenterResponse(w, di.Bid(domain.GermanSoloBid(*param.Bid), trump))
	case "a", "ace":
		if !requireParam(bc, w, newDefault, param.AceSuit == nil, "param error: aceSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.CallAce(*param.AceSuit))
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
