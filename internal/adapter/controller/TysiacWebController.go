//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TysiacWebInput サウザンド (Tysiąc) のWebインプット
type TysiacWebInput struct {
	BaseWebInput
	// CardIndex プレイ/ディスカードするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Raise ビッドする (true=+10 raise, false=pass)
	Raise *bool `json:"raise,omitempty"`
	// Config ゲーム設定
	Config *TysiacWebConfig `json:"config,omitempty"`
}

// TysiacWebConfig サウザンドのWeb設定
type TysiacWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// TysiacWebOutputPlayer サウザンドのWebアウトプットプレイヤー
type TysiacWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// TysiacWebOutput サウザンドのWebアウトプット
type TysiacWebOutput struct {
	Players          []*TysiacWebOutputPlayer    `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	DealerIdx        int                         `json:"dealerIdx"`
	ForehandIdx      int                         `json:"forehandIdx"`
	DeclarerIdx      int                         `json:"declarerIdx"`
	Contract         int                         `json:"contract"`
	CurrentBid       int                         `json:"currentBid"`
	TrumpSuit        int                         `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	PlayerScores     [domain.TysiacPlayerCnt]int `json:"playerScores"`
	RoundCardPoints  [domain.TysiacPlayerCnt]int `json:"roundCardPoints"`
	RoundMarriage    [domain.TysiacPlayerCnt]int `json:"roundMarriage"`
	LastTrickWinner  int                         `json:"lastTrickWinner"`
	PlayableIndices  []int                       `json:"playableIndices"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerPlayer     int                         `json:"winnerPlayer"`
	IsHumanTurn      bool                        `json:"isHumanTurn"`
	Hint             *WebOutputCardHint          `json:"hint,omitempty"`
	WebOutputBase
	Config TysiacWebOutputConfig `json:"config"`
}

// TysiacWebOutputConfig サウザンドの設定アウトプット
type TysiacWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a TysiacConfig from the nested web config, applying bounds checking.
func (c *TysiacWebConfig) ToConfig() domain.TysiacConfig {
	cfg := domain.DefaultTysiacConfig()
	cfg.CpuDifficulty = domain.TysiacCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TysiacCpuDifficultyEasy), int(domain.TysiacCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a TysiacConfig from the web input.
func (p TysiacWebInput) ToConfig() domain.TysiacConfig {
	return configOrDefault(p.Config, (*TysiacWebConfig).ToConfig, domain.DefaultTysiacConfig())
}

// TysiacWebController サウザンドのWebコントローラークラス
type TysiacWebController = GameWebController[usecase.TysiacInteractorIF, TysiacWebInput, *TysiacWebOutput]

// NewTysiacWebController and NewTysiacWebControllerWithProvider are
// the standard and provider-backed constructors for TysiacWebController.
var NewTysiacWebController, NewTysiacWebControllerWithProvider = webControllerPair[usecase.TysiacInteractorIF, TysiacWebInput, *TysiacWebOutput](
	newTysiacDefaultOutput, tysiacDispatch,
)

func newTysiacDefaultOutput(msg string) *TysiacWebOutput {
	return &TysiacWebOutput{
		Players:         make([]*TysiacWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func tysiacDispatch(bc *baseController, w http.ResponseWriter, di usecase.TysiacInteractorIF, param TysiacWebInput, newDefault func(string) *TysiacWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Raise == nil, "param error: raise is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(*param.Raise))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(*param.CardIndex))
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
