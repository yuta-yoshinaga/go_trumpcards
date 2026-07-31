//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KilleWebInput キッレ Webインプット
type KilleWebInput struct {
	BaseWebInput
	Config *KilleWebConfig `json:"config,omitempty"`
}

// KilleWebConfig キッレ Web設定
type KilleWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	Stake         *int `json:"stake,omitempty"`
}

// KilleWebOutputPlayer キッレ Webアウトプットプレイヤー
type KilleWebOutputPlayer struct {
	ID      int            `json:"id"`
	IsHuman bool           `json:"isHuman"`
	Card    *WebOutputCard `json:"card"`
	// Strength は Harlequin の反転を織り込んだ実効的な強さ (非公開なら 0)。
	Strength int `json:"strength"`
	Chips    int `json:"chips"`
	// Reentries はこれまでに買い戻した回数。
	Reentries int `json:"reentries"`
	// ReentryCost は次に買い戻すのに要る額 (もう買い戻せないなら 0)。
	ReentryCost int  `json:"reentryCost"`
	CanReenter  bool `json:"canReenter"`
	IsOut       bool `json:"isOut"`
	// KnockedBy は落とした効果 ("hussar" / "pig" / "" は最弱による脱落)。
	KnockedBy     string `json:"knockedBy"`
	IsSatisfied   bool   `json:"isSatisfied"`
	IsFinished    bool   `json:"isFinished"`
	IsCurrentTurn bool   `json:"isCurrentTurn"`
}

// KilleWebOutputEvent キッレ Webアウトプット交換記録
type KilleWebOutputEvent struct {
	Kind   string `json:"kind"`
	Actor  int    `json:"actor"`
	Target int    `json:"target"`
}

// KilleWebOutput キッレ Webアウトプット
type KilleWebOutput struct {
	Players          []*KilleWebOutputPlayer `json:"players"`
	Phase            int                     `json:"phase"`
	RoundNumber      int                     `json:"roundNumber"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	DealerIdx        int                     `json:"dealerIdx"`
	StockCount       int                     `json:"stockCount"`
	Pot              int                     `json:"pot"`
	Events           []*KilleWebOutputEvent  `json:"events"`
	LoserIdxs        []int                   `json:"loserIdxs"`
	GameEndFlag      bool                    `json:"gameEndFlag"`
	WinnerIdx        int                     `json:"winnerIdx"`
	WebOutputBase
	Config KilleWebOutputConfig `json:"config"`
}

// KilleWebOutputConfig キッレ設定アウトプット
type KilleWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	Stake         int `json:"stake"`
}

// ToConfig builds a KilleConfig from the nested web config, applying bounds checking.
func (c *KilleWebConfig) ToConfig() domain.KilleConfig {
	cfg := domain.DefaultKilleConfig()
	cfg.CpuDifficulty = domain.KilleCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.KilleCpuDifficultyNormal), int(domain.KilleCpuDifficultyNormal), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.Stake, c.Stake, 1, 100)
	return cfg
}

// ToConfig builds a KilleConfig from the web input.
func (p KilleWebInput) ToConfig() domain.KilleConfig {
	return configOrDefault(p.Config, (*KilleWebConfig).ToConfig, domain.DefaultKilleConfig())
}

// KilleWebController キッレ Webコントローラークラス
type KilleWebController = GameWebController[usecase.KilleInteractorIF, KilleWebInput, *KilleWebOutput]

// NewKilleWebController and NewKilleWebControllerWithProvider are
// the standard and provider-backed constructors for KilleWebController.
var NewKilleWebController, NewKilleWebControllerWithProvider = webControllerPair[usecase.KilleInteractorIF, KilleWebInput, *KilleWebOutput](
	newKilleDefaultOutput, killeDispatch,
)

func newKilleDefaultOutput(msg string) *KilleWebOutput {
	return &KilleWebOutput{
		Players:       make([]*KilleWebOutputPlayer, 0),
		Events:        make([]*KilleWebOutputEvent, 0),
		LoserIdxs:     make([]int, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func killeDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KilleInteractorIF, param KilleWebInput, _ func(string) *KilleWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ki.ResetWithConfig(param.ToConfig()))
	case "e", "exchange":
		bc.writePresenterResponse(w, ki.Exchange())
	case "s", "satisfied":
		bc.writePresenterResponse(w, ki.Satisfied())
	case "re", "reenter":
		bc.writePresenterResponse(w, ki.Reenter())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ki.NextRound())
	default:
		return dispatchLog(param.Command, bc, w, ki.ActionLog)
	}
	return true
}
