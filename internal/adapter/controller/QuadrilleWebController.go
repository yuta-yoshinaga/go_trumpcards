//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// QuadrilleWebInput カドリール (Quadrille) のWebインプット
type QuadrilleWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Bid ビッド宣言 (0=pass, 1=entrar, 2=solo)
	Bid *int `json:"bid,omitempty"`
	// TrumpSuit ビッド勝利時に選ぶ切り札スート (1=spade, 2=club, 3=heart, 4=diamond)
	TrumpSuit *int `json:"trumpSuit,omitempty"`
	// KingSuit は王呼びフェーズで指名する王のスート。
	KingSuit *int `json:"kingSuit,omitempty"`
	// Config ゲーム設定
	Config *QuadrilleWebConfig `json:"config,omitempty"`
}

// QuadrilleWebConfig カドリールのWeb設定
type QuadrilleWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// QuadrilleWebOutputPlayer カドリールのWebアウトプットプレイヤー
type QuadrilleWebOutputPlayer struct {
	ID          int              `json:"id"`
	IsHuman     bool             `json:"isHuman"`
	CardCount   int              `json:"cardCount"`
	Cards       []*WebOutputCard `json:"cards"`
	TrickCount  int              `json:"trickCount"`
	Score       int              `json:"score"`
	IsQuadrille bool             `json:"isQuadrille"`
}

// QuadrilleWebOutput カドリールのWebアウトプット
type QuadrilleWebOutput struct {
	Players          []*QuadrilleWebOutputPlayer    `json:"players"`
	Phase            int                            `json:"phase"`
	RoundNumber      int                            `json:"roundNumber"`
	TrickNumber      int                            `json:"trickNumber"`
	CurrentPlayerIdx int                            `json:"currentPlayerIdx"`
	CurrentBidderIdx int                            `json:"currentBidderIdx"`
	LeadPlayerIdx    int                            `json:"leadPlayerIdx"`
	DealerIdx        int                            `json:"dealerIdx"`
	ForehandIdx      int                            `json:"forehandIdx"`
	QuadrilleIdx     int                            `json:"quadrilleIdx"`
	WinningBid       int                            `json:"winningBid"`
	TrumpSuit        int                            `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard          `json:"currentTrick"`
	PlayerScores     [domain.QuadrillePlayerCnt]int `json:"playerScores"`
	LastTrickWinner  int                            `json:"lastTrickWinner"`
	Outcome          int                            `json:"outcome"`
	Result           int                            `json:"result"`
	PlayableIndices  []int                          `json:"playableIndices"`
	GameEndFlag      bool                           `json:"gameEndFlag"`
	WinnerPlayer     int                            `json:"winnerPlayer"`
	IsHumanTurn      bool                           `json:"isHumanTurn"`
	IsHumanBidTurn   bool                           `json:"isHumanBidTurn"`
	// 王呼び。CalledKingSuit は呼ばれた王のスート (-1=未指名)、PartnerIdx は
	// **その王が場に出るまで -1** (誰が味方かは伏せる)。RoiSeul は落札者が
	// 王 4 枚を全部持っていた単独プレイ。
	IsHumanKingCallTurn bool               `json:"isHumanKingCallTurn"`
	CalledKingSuit      int                `json:"calledKingSuit"`
	CallableKingSuits   []int              `json:"callableKingSuits"`
	PartnerIdx          int                `json:"partnerIdx"`
	RoiSeul             bool               `json:"roiSeul"`
	Hint                *WebOutputCardHint `json:"hint,omitempty"`
	WebOutputBase
	Config QuadrilleWebOutputConfig `json:"config"`
}

// QuadrilleWebOutputConfig カドリールの設定アウトプット
type QuadrilleWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds an QuadrilleConfig from the nested web config, applying bounds checking.
func (c *QuadrilleWebConfig) ToConfig() domain.QuadrilleConfig {
	cfg := domain.DefaultQuadrilleConfig()
	cfg.CpuDifficulty = domain.QuadrilleCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.QuadrilleCpuDifficultyEasy), int(domain.QuadrilleCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 1000)
	return cfg
}

// ToConfig builds an QuadrilleConfig from the web input.
func (p QuadrilleWebInput) ToConfig() domain.QuadrilleConfig {
	return configOrDefault(p.Config, (*QuadrilleWebConfig).ToConfig, domain.DefaultQuadrilleConfig())
}

// QuadrilleWebController カドリールのWebコントローラークラス
type QuadrilleWebController = GameWebController[usecase.QuadrilleInteractorIF, QuadrilleWebInput, *QuadrilleWebOutput]

// NewQuadrilleWebController and NewQuadrilleWebControllerWithProvider are
// the standard and provider-backed constructors for QuadrilleWebController.
var NewQuadrilleWebController, NewQuadrilleWebControllerWithProvider = webControllerPair[usecase.QuadrilleInteractorIF, QuadrilleWebInput, *QuadrilleWebOutput](
	newQuadrilleDefaultOutput, quadrilleDispatch,
)

func newQuadrilleDefaultOutput(msg string) *QuadrilleWebOutput {
	return &QuadrilleWebOutput{
		Players:         make([]*QuadrilleWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		QuadrilleIdx:    -1,
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func quadrilleDispatch(bc *baseController, w http.ResponseWriter, di usecase.QuadrilleInteractorIF, param QuadrilleWebInput, newDefault func(string) *QuadrilleWebOutput) bool {
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
		bc.writePresenterResponse(w, di.Bid(domain.QuadrilleBid(*param.Bid), trump))
	case "k", "king":
		if !requireParam(bc, w, newDefault, param.KingSuit == nil, "param error: kingSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.CallKing(*param.KingSuit))
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
