//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MusWebInput ムスのWebインプット
type MusWebInput struct {
	BaseWebInput
	// Mus Mus フェーズ: true=Mus宣言, false=Corte宣言
	Mus *bool `json:"mus,omitempty"`
	// DiscardIndices Discard フェーズ: 交換するカードインデックス一覧
	DiscardIndices []int `json:"discardIndices,omitempty"`
	// BetAction 賭けフェーズ: アクション (0=paso,1=envido,2=ordago,3=quiero,4=noquiero)
	BetAction *int `json:"betAction,omitempty"`
	// BetAmount Envido の賭け額
	BetAmount *int `json:"betAmount,omitempty"`
	// Config ゲーム設定
	Config *MusWebConfig `json:"config,omitempty"`
}

// MusWebConfig ムスのWeb設定
type MusWebConfig struct {
	CpuDifficulty   *int `json:"cpuDifficulty,omitempty"`
	TargetAmarrakos *int `json:"targetAmarrakos,omitempty"`
}

// MusWebOutputPlayer ムスのWebアウトプットプレイヤー
type MusWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	TeamScore int              `json:"teamScore"`
}

// MusWebOutputResult 賭けラウンド結果
type MusWebOutputResult struct {
	Kind  int `json:"kind"`
	Stake int `json:"stake"`
	Team  int `json:"team"`
}

// MusWebOutputHint ヒント出力
type MusWebOutputHint struct {
	Mus     bool   `json:"mus"`
	Action  int    `json:"action"`
	Amount  int    `json:"amount"`
	Indices []int  `json:"indices"`
	Reason  string `json:"reason"`
}

// MusWebOutput ムスのWebアウトプット
type MusWebOutput struct {
	Players        []*MusWebOutputPlayer                  `json:"players"`
	Phase          int                                    `json:"phase"`
	RoundNumber    int                                    `json:"roundNumber"`
	ManoIdx        int                                    `json:"manoIdx"`
	BetTeam        int                                    `json:"betTeam"`
	PendingStake   int                                    `json:"pendingStake"`
	LastBettorTeam int                                    `json:"lastBettorTeam"`
	MusTurn        int                                    `json:"musTurn"`
	DiscardTurn    int                                    `json:"discardTurn"`
	MusCycle       int                                    `json:"musCycle"`
	Amarrakos      [domain.MusTeamCnt]int                 `json:"amarrakos"`
	Results        [domain.MusRoundCnt]MusWebOutputResult `json:"results"`
	GameEndFlag    bool                                   `json:"gameEndFlag"`
	WinnerTeam     int                                    `json:"winnerTeam"`
	HumanTeam      int                                    `json:"humanTeam"`
	IsHumanTurn    bool                                   `json:"isHumanTurn"`
	CanPaso        bool                                   `json:"canPaso"`
	CanEnvido      bool                                   `json:"canEnvido"`
	CanOrdago      bool                                   `json:"canOrdago"`
	CanQuiero      bool                                   `json:"canQuiero"`
	CanNoQuiero    bool                                   `json:"canNoQuiero"`
	Hint           *MusWebOutputHint                      `json:"hint,omitempty"`
	Config         MusWebOutputConfig                     `json:"config"`
	WebOutputBase
}

// MusWebOutputConfig ムスの設定アウトプット
type MusWebOutputConfig struct {
	CpuDifficulty   int `json:"cpuDifficulty"`
	TargetAmarrakos int `json:"targetAmarrakos"`
}

// ToConfig builds a MusConfig from the nested web config, applying bounds checking.
func (c *MusWebConfig) ToConfig() domain.MusConfig {
	cfg := domain.DefaultMusConfig()
	cfg.CpuDifficulty = domain.MusCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.MusCpuDifficultyEasy), int(domain.MusCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetAmarrakos, c.TargetAmarrakos, 1, 1000)
	return cfg
}

// ToConfig builds a MusConfig from the web input.
func (p MusWebInput) ToConfig() domain.MusConfig {
	return configOrDefault(p.Config, (*MusWebConfig).ToConfig, domain.DefaultMusConfig())
}

// MusWebController ムスのWebコントローラークラス
type MusWebController = GameWebController[usecase.MusInteractorIF, MusWebInput, *MusWebOutput]

// NewMusWebController and NewMusWebControllerWithProvider are
// the standard and provider-backed constructors for MusWebController.
var NewMusWebController, NewMusWebControllerWithProvider = webControllerPair[usecase.MusInteractorIF, MusWebInput, *MusWebOutput](
	newMusDefaultOutput, musDispatch,
)

func newMusDefaultOutput(msg string) *MusWebOutput {
	return &MusWebOutput{
		Players:       make([]*MusWebOutputPlayer, 0),
		WinnerTeam:    -1,
		HumanTeam:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func musDispatch(bc *baseController, w http.ResponseWriter, mi usecase.MusInteractorIF, param MusWebInput, newDefault func(string) *MusWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, mi.ResetWithConfig(param.ToConfig()))
	case "mus":
		if !requireParam(bc, w, newDefault, param.Mus == nil, "param error: mus is required.") {
			return true
		}
		bc.writePresenterResponse(w, mi.Mus(*param.Mus))
	case "discard":
		// discardIndices may be empty (keep all cards) — nil vs empty slice distinction
		if param.DiscardIndices == nil {
			bc.writePresenterResponse(w, mi.Discard([]int{}))
		} else {
			bc.writePresenterResponse(w, mi.Discard(param.DiscardIndices))
		}
	case "bet":
		if !requireParam(bc, w, newDefault, param.BetAction == nil, "param error: betAction is required.") {
			return true
		}
		amount := 0
		if param.BetAmount != nil {
			amount = *param.BetAmount
		}
		bc.writePresenterResponse(w, mi.Bet(*param.BetAction, amount))
	case "n", "next":
		bc.writePresenterResponse(w, mi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, mi.Hint, mi.ActionLog)
	}
	return true
}
