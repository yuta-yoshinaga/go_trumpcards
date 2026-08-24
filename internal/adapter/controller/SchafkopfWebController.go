//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SchafkopfWebInput シャーフコップのWebインプット
type SchafkopfWebInput struct {
	BaseWebInput
	// Pick ピックフェーズでのブラインド取得可否 (pick コマンド)
	Pick *bool `json:"pick,omitempty"`
	// Contract 宣言する契約 (pick コマンド; 0=Rufspiel 1=Wenz 2=Solo)
	Contract *int `json:"contract,omitempty"`
	// SoloSuit Solo の切り札スート (pick コマンド; 1=♠ 2=♣ 3=♥ 4=♦)
	SoloSuit *int `json:"soloSuit,omitempty"`
	// CallSuit 呼びスートのインデックス (call コマンド; 1=♠ 2=♣ 3=♥)
	CallSuit *int `json:"callSuit,omitempty"`
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *SchafkopfWebConfig `json:"config,omitempty"`
}

// SchafkopfWebConfig シャーフコップのWeb設定
type SchafkopfWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	BaseChips     *int `json:"baseChips,omitempty"`
	StartChips    *int `json:"startChips,omitempty"`
	TargetChips   *int `json:"targetChips,omitempty"`
}

// SchafkopfWebOutputPlayer シャーフコップのWebアウトプットプレイヤー
type SchafkopfWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Chips      int              `json:"chips"`
}

// SchafkopfWebOutputHint ヒント出力
type SchafkopfWebOutputHint struct {
	CardIndices []int  `json:"cardIndices"`
	Suit        int    `json:"suit"`
	Pick        bool   `json:"pick"`
	Reason      string `json:"reason"`
}

// SchafkopfWebOutput シャーフコップのWebアウトプット
type SchafkopfWebOutput struct {
	Players          []*SchafkopfWebOutputPlayer `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	DealerIdx        int                         `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	PickerIdx        int                         `json:"pickerIdx"`
	// Contract は切り札の構成そのもの。返さないと、Wenz の盤面で Ober が
	// 切り札でない理由が画面から読み取れない。
	Contract int `json:"contract"`
	SoloSuit int `json:"soloSuit"`
	// BeatableContracts は今この席が宣言できる契約。押せるのに必ず拒否
	// されるボタンを描かないために要る。
	BeatableContracts []int                   `json:"beatableContracts"`
	PartnerIdx        int                     `json:"partnerIdx"`
	CalledSuit        int                     `json:"calledSuit"`
	PartnerRevealed   bool                    `json:"partnerRevealed"`
	PassCount         int                     `json:"passCount"`
	CallableSuits     []int                   `json:"callableSuits"`
	PlayableIndices   []int                   `json:"playableIndices"`
	RoundPickerPoints int                     `json:"roundPickerPoints"`
	RoundMultiplier   int                     `json:"roundMultiplier"`
	RoundPickerWon    bool                    `json:"roundPickerWon"`
	GameEndFlag       bool                    `json:"gameEndFlag"`
	WinnerIdx         int                     `json:"winnerIdx"`
	Hint              *SchafkopfWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config SchafkopfWebOutputConfig `json:"config"`
}

// SchafkopfWebOutputConfig シャーフコップの設定アウトプット
type SchafkopfWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	BaseChips     int `json:"baseChips"`
	StartChips    int `json:"startChips"`
	TargetChips   int `json:"targetChips"`
}

// ToConfig builds a SchafkopfConfig from the nested web config, applying bounds checking.
func (c *SchafkopfWebConfig) ToConfig() domain.SchafkopfConfig {
	cfg := domain.DefaultSchafkopfConfig()
	cfg.CpuDifficulty = domain.SchafkopfCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SchafkopfCpuDifficultyEasy), int(domain.SchafkopfCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.BaseChips, c.BaseChips, 1, 100000)
	webutil.ApplyBoundedInt(&cfg.StartChips, c.StartChips, 1, 100000)
	webutil.ApplyBoundedInt(&cfg.TargetChips, c.TargetChips, cfg.StartChips+1, 1000000)
	return cfg
}

// ToConfig builds a SchafkopfConfig from the web input.
func (p SchafkopfWebInput) ToConfig() domain.SchafkopfConfig {
	return configOrDefault(p.Config, (*SchafkopfWebConfig).ToConfig, domain.DefaultSchafkopfConfig())
}

// SchafkopfWebController シャーフコップのWebコントローラークラス
type SchafkopfWebController = GameWebController[usecase.SchafkopfInteractorIF, SchafkopfWebInput, *SchafkopfWebOutput]

// NewSchafkopfWebController and NewSchafkopfWebControllerWithProvider are
// the standard and provider-backed constructors for SchafkopfWebController.
var NewSchafkopfWebController, NewSchafkopfWebControllerWithProvider = webControllerPair[usecase.SchafkopfInteractorIF, SchafkopfWebInput, *SchafkopfWebOutput](
	newSchafkopfDefaultOutput, schafkopfDispatch,
)

func newSchafkopfDefaultOutput(msg string) *SchafkopfWebOutput {
	return &SchafkopfWebOutput{
		Players:         make([]*SchafkopfWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		CallableSuits:   make([]int, 0),
		PlayableIndices: make([]int, 0),
		PickerIdx:       -1,
		PartnerIdx:      -1,
		WinnerIdx:       -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func schafkopfDispatch(bc *baseController, w http.ResponseWriter, si usecase.SchafkopfInteractorIF, param SchafkopfWebInput, newDefault func(string) *SchafkopfWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "pick":
		if !requireParam(bc, w, newDefault, param.Pick == nil, "param error: pick is required.") {
			return true
		}
		contract := domain.SchafkopfContractRufspiel
		if param.Contract != nil {
			contract = domain.SchafkopfContract(*param.Contract)
		}
		soloSuit := 0
		if param.SoloSuit != nil {
			soloSuit = *param.SoloSuit
		}
		bc.writePresenterResponse(w, si.Declare(*param.Pick, contract, soloSuit))
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
