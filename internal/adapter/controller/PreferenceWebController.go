//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PreferenceWebInput プレフェランスのWebインプット
type PreferenceWebInput struct {
	BaseWebInput
	// Bid 入札種別 (0=Pass,1=Six,2=Misère,3=Seven,4=Eight)
	Bid *int `json:"bid,omitempty"`
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *PreferenceWebConfig `json:"config,omitempty"`
}

// PreferenceWebConfig プレフェランスのWeb設定
type PreferenceWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// PreferenceWebOutputPlayer プレフェランスのWebアウトプットプレイヤー
type PreferenceWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// PreferenceWebOutput プレフェランスのWebアウトプット
type PreferenceWebOutput struct {
	Players          []*PreferenceWebOutputPlayer    `json:"players"`
	Phase            int                             `json:"phase"`
	RoundNumber      int                             `json:"roundNumber"`
	TrickNumber      int                             `json:"trickNumber"`
	CurrentPlayerIdx int                             `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                             `json:"leadPlayerIdx"`
	DealerIdx        int                             `json:"dealerIdx"`
	DeclarerIdx      int                             `json:"declarerIdx"`
	Contract         int                             `json:"contract"`
	TrumpSuit        int                             `json:"trumpSuit"`
	Bids             [domain.PreferencePlayerCnt]int `json:"bids"`
	CurrentTrick     []*WebOutputTrickCard           `json:"currentTrick"`
	PlayerScores     [domain.PreferencePlayerCnt]int `json:"playerScores"`
	RoundTricks      [domain.PreferencePlayerCnt]int `json:"roundTricks"`
	PlayableIndices  []int                           `json:"playableIndices"`
	GameEndFlag      bool                            `json:"gameEndFlag"`
	WinnerPlayer     int                             `json:"winnerPlayer"`
	IsHumanTurn      bool                            `json:"isHumanTurn"`
	IsHumanBidTurn   bool                            `json:"isHumanBidTurn"`
	Hint             *WebOutputCardHint              `json:"hint,omitempty"`
	WebOutputBase
	Config PreferenceWebOutputConfig `json:"config"`
}

// PreferenceWebOutputConfig プレフェランスの設定アウトプット
type PreferenceWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a PreferenceConfig from the nested web config, applying bounds checking.
func (c *PreferenceWebConfig) ToConfig() domain.PreferenceConfig {
	cfg := domain.DefaultPreferenceConfig()
	cfg.CpuDifficulty = domain.PreferenceCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.PreferenceCpuDifficultyEasy), int(domain.PreferenceCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a PreferenceConfig from the web input.
func (p PreferenceWebInput) ToConfig() domain.PreferenceConfig {
	return configOrDefault(p.Config, (*PreferenceWebConfig).ToConfig, domain.DefaultPreferenceConfig())
}

// PreferenceWebController プレフェランスのWebコントローラークラス
type PreferenceWebController = GameWebController[usecase.PreferenceInteractorIF, PreferenceWebInput, *PreferenceWebOutput]

// NewPreferenceWebController and NewPreferenceWebControllerWithProvider are
// the standard and provider-backed constructors for PreferenceWebController.
var NewPreferenceWebController, NewPreferenceWebControllerWithProvider = webControllerPair[usecase.PreferenceInteractorIF, PreferenceWebInput, *PreferenceWebOutput](
	newPreferenceDefaultOutput, preferenceDispatch,
)

func newPreferenceDefaultOutput(msg string) *PreferenceWebOutput {
	return &PreferenceWebOutput{
		Players:         make([]*PreferenceWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func preferenceDispatch(bc *baseController, w http.ResponseWriter, di usecase.PreferenceInteractorIF, param PreferenceWebInput, newDefault func(string) *PreferenceWebOutput) bool {
	return dispatchBidTrickPlay(param.Command, bc, w, bidTrickPlayFns{
		resetWithConfig: func() string { return di.ResetWithConfig(param.ToConfig()) },
		bid:             di.Bid,
		play:            di.Play,
		nextTrick:       di.NextTrick,
		nextRound:       di.NextRound,
		hint:            di.Hint,
		actionLog:       di.ActionLog,
	}, param.Bid, param.CardIndex, newDefault)
}
