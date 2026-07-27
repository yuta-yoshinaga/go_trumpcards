//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SheepsheadWebInput シープスヘッドのWebインプット
type SheepsheadWebInput struct {
	BaseWebInput
	// Pick ピックフェーズでのブラインド取得可否 (pick コマンド)
	Pick *bool `json:"pick,omitempty"`
	// BuryIndices 埋め札のカードインデックス (bury コマンド)
	BuryIndices []int `json:"buryIndices,omitempty"`
	// CallSuit 呼びスートのインデックス (call コマンド; 1=♠ 2=♣ 3=♥)
	CallSuit *int `json:"callSuit,omitempty"`
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *SheepsheadWebConfig `json:"config,omitempty"`
}

// SheepsheadWebConfig シープスヘッドのWeb設定
type SheepsheadWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	BaseChips     *int `json:"baseChips,omitempty"`
	StartChips    *int `json:"startChips,omitempty"`
	TargetChips   *int `json:"targetChips,omitempty"`
}

// SheepsheadWebOutputPlayer シープスヘッドのWebアウトプットプレイヤー
type SheepsheadWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Chips      int              `json:"chips"`
}

// SheepsheadWebOutputHint ヒント出力
type SheepsheadWebOutputHint struct {
	CardIndices []int  `json:"cardIndices"`
	Suit        int    `json:"suit"`
	Pick        bool   `json:"pick"`
	Reason      string `json:"reason"`
}

// SheepsheadWebOutput シープスヘッドのWebアウトプット
type SheepsheadWebOutput struct {
	Players          []*SheepsheadWebOutputPlayer `json:"players"`
	Phase            int                          `json:"phase"`
	RoundNumber      int                          `json:"roundNumber"`
	TrickNumber      int                          `json:"trickNumber"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                          `json:"leadPlayerIdx"`
	DealerIdx        int                          `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard        `json:"currentTrick"`
	// BlindCount ブラインドの枚数 (ピックフェーズ中は枚数のみ公開)
	BlindCount        int                      `json:"blindCount"`
	Buried            []*WebOutputCard         `json:"buried"`
	PickerIdx         int                      `json:"pickerIdx"`
	PartnerIdx        int                      `json:"partnerIdx"`
	CalledSuit        int                      `json:"calledSuit"`
	PartnerRevealed   bool                     `json:"partnerRevealed"`
	PassCount         int                      `json:"passCount"`
	CallableSuits     []int                    `json:"callableSuits"`
	PlayableIndices   []int                    `json:"playableIndices"`
	RoundPickerPoints int                      `json:"roundPickerPoints"`
	RoundMultiplier   int                      `json:"roundMultiplier"`
	RoundPickerWon    bool                     `json:"roundPickerWon"`
	GameEndFlag       bool                     `json:"gameEndFlag"`
	WinnerIdx         int                      `json:"winnerIdx"`
	Hint              *SheepsheadWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config SheepsheadWebOutputConfig `json:"config"`
}

// SheepsheadWebOutputConfig シープスヘッドの設定アウトプット
type SheepsheadWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	BaseChips     int `json:"baseChips"`
	StartChips    int `json:"startChips"`
	TargetChips   int `json:"targetChips"`
}

// ToConfig builds a SheepsheadConfig from the nested web config, applying bounds checking.
func (c *SheepsheadWebConfig) ToConfig() domain.SheepsheadConfig {
	cfg := domain.DefaultSheepsheadConfig()
	cfg.CpuDifficulty = domain.SheepsheadCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SheepsheadCpuDifficultyEasy), int(domain.SheepsheadCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.BaseChips, c.BaseChips, 1, 100000)
	webutil.ApplyBoundedInt(&cfg.StartChips, c.StartChips, 1, 100000)
	webutil.ApplyBoundedInt(&cfg.TargetChips, c.TargetChips, cfg.StartChips+1, 1000000)
	return cfg
}

// ToConfig builds a SheepsheadConfig from the web input.
func (p SheepsheadWebInput) ToConfig() domain.SheepsheadConfig {
	return configOrDefault(p.Config, (*SheepsheadWebConfig).ToConfig, domain.DefaultSheepsheadConfig())
}

// SheepsheadWebController シープスヘッドのWebコントローラークラス
type SheepsheadWebController = GameWebController[usecase.SheepsheadInteractorIF, SheepsheadWebInput, *SheepsheadWebOutput]

// NewSheepsheadWebController and NewSheepsheadWebControllerWithProvider are
// the standard and provider-backed constructors for SheepsheadWebController.
var NewSheepsheadWebController, NewSheepsheadWebControllerWithProvider = webControllerPair[usecase.SheepsheadInteractorIF, SheepsheadWebInput, *SheepsheadWebOutput](
	newSheepsheadDefaultOutput, sheepsheadDispatch,
)

func newSheepsheadDefaultOutput(msg string) *SheepsheadWebOutput {
	return &SheepsheadWebOutput{
		Players:         make([]*SheepsheadWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		Buried:          make([]*WebOutputCard, 0),
		CallableSuits:   make([]int, 0),
		PlayableIndices: make([]int, 0),
		PickerIdx:       -1,
		PartnerIdx:      -1,
		WinnerIdx:       -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func sheepsheadDispatch(bc *baseController, w http.ResponseWriter, si usecase.SheepsheadInteractorIF, param SheepsheadWebInput, newDefault func(string) *SheepsheadWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "pick":
		if !requireParam(bc, w, newDefault, param.Pick == nil, "param error: pick is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Pick(*param.Pick))
	case "bury":
		if !requireParam(bc, w, newDefault, len(param.BuryIndices) == 0, "param error: buryIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Bury(param.BuryIndices))
	case "call":
		if !requireParam(bc, w, newDefault, param.CallSuit == nil, "param error: callSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Call(*param.CallSuit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, si.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
