//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// UltiWebInput ウルティ (Ulti) のWebインプット
type UltiWebInput struct {
	BaseWebInput
	// CardIndex プレイするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices 捨てるカードのインデックス (discard で使用)
	CardIndices []int `json:"cardIndices,omitempty"`
	// Contract コントラクト ("party" / "betli" / "durchmarsch")
	Contract *string `json:"contract,omitempty"`
	// TrumpSuit Party 時に選ぶ切り札スート (1=spade, 2=club, 3=heart, 4=diamond)
	TrumpSuit *int `json:"trumpSuit,omitempty"`
	// Config ゲーム設定
	Config *UltiWebConfig `json:"config,omitempty"`
}

// UltiWebConfig ウルティのWeb設定
type UltiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// UltiWebOutputPlayer ウルティのWebアウトプットプレイヤー
type UltiWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	CardPoints int              `json:"cardPoints"`
	Coins      int              `json:"coins"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// UltiWebOutput ウルティのWebアウトプット
type UltiWebOutput struct {
	Players          []*UltiWebOutputPlayer    `json:"players"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	TrickNumber      int                       `json:"trickNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                       `json:"leadPlayerIdx"`
	DealerIdx        int                       `json:"dealerIdx"`
	DeclarerIdx      int                       `json:"declarerIdx"`
	Contract         int                       `json:"contract"`
	TrumpSuit        int                       `json:"trumpSuit"`
	TalonCount       int                       `json:"talonCount"`
	TalonTaken       bool                      `json:"talonTaken"`
	DiscardCount     int                       `json:"discardCount"`
	CurrentTrick     []*WebOutputTrickCard     `json:"currentTrick"`
	PlayerCoins      [domain.UltiPlayerCnt]int `json:"playerCoins"`
	// LastDealCoins は直近ディールの精算による符号付き増減。累積からは読めない。
	LastDealCoins   [domain.UltiPlayerCnt]int `json:"lastDealCoins"`
	LastTrickWinner int                       `json:"lastTrickWinner"`
	Outcome         int                       `json:"outcome"`
	Result          int                       `json:"result"`
	PlayableIndices []int                     `json:"playableIndices"`
	GameEndFlag     bool                      `json:"gameEndFlag"`
	WinnerPlayer    int                       `json:"winnerPlayer"`
	IsHumanTurn     bool                      `json:"isHumanTurn"`
	IsHumanBidTurn  bool                      `json:"isHumanBidTurn"`
	Hint            *WebOutputCardHint        `json:"hint,omitempty"`
	WebOutputBase
	Config UltiWebOutputConfig `json:"config"`
}

// UltiWebOutputConfig ウルティの設定アウトプット
type UltiWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds an UltiConfig from the nested web config, applying bounds checking.
func (c *UltiWebConfig) ToConfig() domain.UltiConfig {
	cfg := domain.DefaultUltiConfig()
	cfg.CpuDifficulty = domain.UltiCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.UltiCpuDifficultyEasy), int(domain.UltiCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, 1, 1000)
	return cfg
}

// ToConfig builds an UltiConfig from the web input.
func (p UltiWebInput) ToConfig() domain.UltiConfig {
	return configOrDefault(p.Config, (*UltiWebConfig).ToConfig, domain.DefaultUltiConfig())
}

// ultiParseContract コントラクト文字列を UltiContract に変換する (None=不正)。
func ultiParseContract(s string) domain.UltiContract {
	switch s {
	case "party", "p":
		return domain.UltiContractParty
	case "betli", "b":
		return domain.UltiContractBetli
	case "durchmarsch", "d":
		return domain.UltiContractDurchmarsch
	case "ulti", "u":
		return domain.UltiContractUlti
	default:
		return domain.UltiContractNone
	}
}

// UltiWebController ウルティのWebコントローラークラス
type UltiWebController = GameWebController[usecase.UltiInteractorIF, UltiWebInput, *UltiWebOutput]

// NewUltiWebController and NewUltiWebControllerWithProvider are
// the standard and provider-backed constructors for UltiWebController.
var NewUltiWebController, NewUltiWebControllerWithProvider = webControllerPair[usecase.UltiInteractorIF, UltiWebInput, *UltiWebOutput](
	newUltiDefaultOutput, ultiDispatch,
)

func newUltiDefaultOutput(msg string) *UltiWebOutput {
	return &UltiWebOutput{
		Players:         make([]*UltiWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     0,
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func ultiDispatch(bc *baseController, w http.ResponseWriter, di usecase.UltiInteractorIF, param UltiWebInput, newDefault func(string) *UltiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Contract == nil, "param error: contract is required.") {
			return true
		}
		trump := -1
		if param.TrumpSuit != nil {
			trump = *param.TrumpSuit
		}
		bc.writePresenterResponse(w, di.Bid(ultiParseContract(*param.Contract), trump))
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
