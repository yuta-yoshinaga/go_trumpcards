//go:build !js || !wasm || extra4

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PiedmonteseTarotWebInput はピエモンテ・タロッコの Web インプット。
type PiedmonteseTarotWebInput struct {
	BaseWebInput
	// CardIndex はプレイするカードのインデックス。
	CardIndex *int `json:"cardIndex,omitempty"`
	// CardIndices はスカルトで捨てるカードのインデックス (タロンの枚数ぶん)。
	CardIndices []int `json:"cardIndices,omitempty"`
	// Config はゲーム設定。
	Config *PiedmonteseTarotWebConfig `json:"config,omitempty"`
}

// PiedmonteseTarotWebConfig はピエモンテ・タロッコの Web 設定。
type PiedmonteseTarotWebConfig struct {
	Seats         *int `json:"seats,omitempty"`
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// PiedmonteseTarotWebOutputPlayer は 1 席ぶんの出力。
type PiedmonteseTarotWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	// CardThirds は獲得した札の点を 1/3 単位で表したもの。
	CardThirds int `json:"cardThirds"`
	// CardPoints は同じ点を読める形にしたもの ("26 1/3" など)。
	CardPoints string `json:"cardPoints"`
	Score      int    `json:"score"`
	IsDealer   bool   `json:"isDealer"`
}

// PiedmonteseTarotWebOutput はピエモンテ・タロッコの Web アウトプット。
type PiedmonteseTarotWebOutput struct {
	Players          []*PiedmonteseTarotWebOutputPlayer `json:"players"`
	Phase            int                                `json:"phase"`
	RoundNumber      int                                `json:"roundNumber"`
	TrickNumber      int                                `json:"trickNumber"`
	TrickCount       int                                `json:"trickCount"`
	CurrentPlayerIdx int                                `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                                `json:"leadPlayerIdx"`
	DealerIdx        int                                `json:"dealerIdx"`
	ScartoCount      int                                `json:"scartoCount"`
	// TalonSize は親が捨てる枚数 (席数で変わる: 4 人なら 2、3 人なら 3)。
	TalonSize       int                   `json:"talonSize"`
	CurrentTrick    []*WebOutputTrickCard `json:"currentTrick"`
	PlayerScores    []int                 `json:"playerScores"`
	DealScores      []int                 `json:"dealScores"`
	LastTrickWinner int                   `json:"lastTrickWinner"`
	Outcome         int                   `json:"outcome"`
	Result          int                   `json:"result"`
	PlayableIndices []int                 `json:"playableIndices"`
	// DiscardableIndices は親がスカルトに出せる札。**ピップが足りなければ
	// 非オヌール切り札も含む** —— 画面側で色や値から作ると再現できない (#6236)。
	DiscardableIndices []int              `json:"discardableIndices"`
	GameEndFlag        bool               `json:"gameEndFlag"`
	WinnerPlayer       int                `json:"winnerPlayer"`
	IsHumanTurn        bool               `json:"isHumanTurn"`
	IsHumanScarto      bool               `json:"isHumanScarto"`
	Hint               *WebOutputCardHint `json:"hint,omitempty"`
	WebOutputBase
	Config PiedmonteseTarotWebOutputConfig `json:"config"`
}

// PiedmonteseTarotWebOutputConfig は設定アウトプット。
type PiedmonteseTarotWebOutputConfig struct {
	Seats         int `json:"seats"`
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *PiedmonteseTarotWebConfig) ToConfig() domain.PiedmonteseTarotConfig {
	cfg := domain.DefaultPiedmonteseTarotConfig()
	// **席数は一覧に無い数を丸めない。** 配り方が決まっているのは 3 人と 4 人だけで、
	// 5 を 4 に丸めると要求と違う卓が「成功」として返る。
	if c.Seats != nil && domain.PiedmonteseTarotHandSize(*c.Seats) > 0 {
		cfg.Seats = *c.Seats
	}
	cfg.CpuDifficulty = domain.PiedmonteseTarotCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.PiedmonteseTarotCpuDifficultyEasy),
		int(domain.PiedmonteseTarotCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals, 1, 1000)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p PiedmonteseTarotWebInput) ToConfig() domain.PiedmonteseTarotConfig {
	return configOrDefault(p.Config, (*PiedmonteseTarotWebConfig).ToConfig,
		domain.DefaultPiedmonteseTarotConfig())
}

// PiedmonteseTarotWebController はピエモンテ・タロッコの Web コントローラー。
type PiedmonteseTarotWebController = GameWebController[usecase.PiedmonteseTarotInteractorIF, PiedmonteseTarotWebInput, *PiedmonteseTarotWebOutput]

// NewPiedmonteseTarotWebController, NewPiedmonteseTarotWebControllerWithProvider are
// the standard and provider-backed constructors.
var NewPiedmonteseTarotWebController, NewPiedmonteseTarotWebControllerWithProvider = webControllerPair[usecase.PiedmonteseTarotInteractorIF, PiedmonteseTarotWebInput, *PiedmonteseTarotWebOutput](
	newPiedmonteseTarotDefaultOutput, piedmonteseTarotDispatch,
)

func newPiedmonteseTarotDefaultOutput(msg string) *PiedmonteseTarotWebOutput {
	return &PiedmonteseTarotWebOutput{
		Players:         make([]*PiedmonteseTarotWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		PlayerScores:    make([]int, 0),
		DealScores:      make([]int, 0),
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func piedmonteseTarotDispatch(bc *baseController, w http.ResponseWriter, di usecase.PiedmonteseTarotInteractorIF, param PiedmonteseTarotWebInput, newDefault func(string) *PiedmonteseTarotWebOutput) bool {
	return dispatchTarotDiscardPlay(param.Command, bc, w, tarotDiscardPlayFns{
		resetWithConfig: func() string { return di.ResetWithConfig(param.ToConfig()) },
		discard:         di.Discard,
		play:            di.Play,
		nextTrick:       di.NextTrick,
		nextRound:       di.NextRound,
		hint:            di.Hint,
		actionLog:       di.ActionLog,
	}, param.CardIndices, param.CardIndex, newDefault)
}
