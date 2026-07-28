//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KoiKoiWebConfig はこいこい (Koi-Koi) の Web 設定。
type KoiKoiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// ToConfig は KoiKoiWebConfig を domain.KoiKoiConfig に変換する (境界チェック付き)。
func (c *KoiKoiWebConfig) ToConfig() domain.KoiKoiConfig {
	cfg := domain.DefaultKoiKoiConfig()
	cfg.CpuDifficulty = domain.KoiKoiCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty, int(domain.KoiKoiCpuDifficultyEasy), int(domain.KoiKoiCpuDifficultyHard), int(cfg.CpuDifficulty)))
	cfg.TargetScore = webutil.BoundedIntPtr(
		c.TargetScore, domain.KoiKoiTargetScoreMin, domain.KoiKoiTargetScoreMax, cfg.TargetScore)
	return cfg
}

// KoiKoiWebInput はこいこい Web インプット。
type KoiKoiWebInput struct {
	BaseWebInput
	CardIndex  *int             `json:"cardIndex,omitempty"`
	FieldIndex *int             `json:"fieldIndex,omitempty"`
	Config     *KoiKoiWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.KoiKoiConfig を構築する。
func (p KoiKoiWebInput) ToConfig() domain.KoiKoiConfig {
	return configOrDefault(p.Config, (*KoiKoiWebConfig).ToConfig, domain.DefaultKoiKoiConfig())
}

// KoiKoiWebOutputYaku は成立役 1 件。
type KoiKoiWebOutputYaku struct {
	Key    string `json:"key"`
	Points int    `json:"points"`
}

// KoiKoiWebOutputPlayer はこいこい Web アウトプットプレイヤー。
type KoiKoiWebOutputPlayer struct {
	ID            int                    `json:"id"`
	IsHuman       bool                   `json:"isHuman"`
	CardCount     int                    `json:"cardCount"`
	Cards         []*WebOutputCard       `json:"cards"`
	Captured      []*WebOutputCard       `json:"captured"`
	CapturedCount int                    `json:"capturedCount"`
	Score         int                    `json:"score"`
	CalledKoiKoi  bool                   `json:"calledKoiKoi"`
	Yaku          []*KoiKoiWebOutputYaku `json:"yaku"`
	YakuPoints    int                    `json:"yakuPoints"`
}

// KoiKoiWebOutputRoundResult は 1 ラウンドの結果。
type KoiKoiWebOutputRoundResult struct {
	Winner      int                    `json:"winner"`
	Yaku        []*KoiKoiWebOutputYaku `json:"yaku"`
	BasePoints  int                    `json:"basePoints"`
	Multiplier  int                    `json:"multiplier"`
	Total       int                    `json:"total"`
	KoikoiCount int                    `json:"koikoiCount"`
}

// KoiKoiWebOutputHint はヒント出力。
type KoiKoiWebOutputHint struct {
	CardIndex  int    `json:"cardIndex"`
	FieldIndex int    `json:"fieldIndex"`
	KoiKoi     int    `json:"koikoi"`
	Reason     string `json:"reason"`
}

// KoiKoiWebOutput はこいこい Web アウトプット。
type KoiKoiWebOutput struct {
	Players         []*KoiKoiWebOutputPlayer    `json:"players"`
	Phase           int                         `json:"phase"`
	RoundNumber     int                         `json:"roundNumber"`
	CurrentTurn     int                         `json:"currentTurn"`
	FieldCards      []*WebOutputCard            `json:"fieldCards"`
	RemainingDeck   int                         `json:"remainingDeck"`
	KoikoiCount     int                         `json:"koikoiCount"`
	PlayableIndices []int                       `json:"playableIndices"`
	CaptureOptions  map[int][]int               `json:"captureOptions"`
	PendingYaku     []*KoiKoiWebOutputYaku      `json:"pendingYaku"`
	PendingPoints   int                         `json:"pendingPoints"`
	RoundWinner     int                         `json:"roundWinner"`
	Winner          int                         `json:"winner"`
	GameEndFlag     bool                        `json:"gameEndFlag"`
	IsHumanTurn     bool                        `json:"isHumanTurn"`
	LastRoundResult *KoiKoiWebOutputRoundResult `json:"lastRoundResult"`
	Hint            *KoiKoiWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config KoiKoiWebConfigOutput `json:"config"`
}

// KoiKoiWebConfigOutput は設定アウトプット。
type KoiKoiWebConfigOutput struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// KoiKoiWebController はこいこい Web コントローラークラス。
type KoiKoiWebController = GameWebController[usecase.KoiKoiInteractorIF, KoiKoiWebInput, *KoiKoiWebOutput]

// NewKoiKoiWebController, NewKoiKoiWebControllerWithProvider are the standard and
// provider-backed constructors for KoiKoiWebController.
var NewKoiKoiWebController, NewKoiKoiWebControllerWithProvider = webControllerPair[usecase.KoiKoiInteractorIF, KoiKoiWebInput, *KoiKoiWebOutput](
	newKoiKoiDefaultOutput, koikoiDispatch,
)

func newKoiKoiDefaultOutput(msg string) *KoiKoiWebOutput {
	return &KoiKoiWebOutput{
		Players:         make([]*KoiKoiWebOutputPlayer, 0),
		FieldCards:      make([]*WebOutputCard, 0),
		PlayableIndices: make([]int, 0),
		CaptureOptions:  make(map[int][]int),
		PendingYaku:     make([]*KoiKoiWebOutputYaku, 0),
		RoundWinner:     -1,
		Winner:          -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func koikoiDispatch(bc *baseController, w http.ResponseWriter, ki usecase.KoiKoiInteractorIF, param KoiKoiWebInput, newDefault func(string) *KoiKoiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ki.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		fieldIdx := -1
		if param.FieldIndex != nil {
			fieldIdx = *param.FieldIndex
		}
		bc.writePresenterResponse(w, ki.Play(*param.CardIndex, fieldIdx))
	case "kk", "koikoi":
		bc.writePresenterResponse(w, ki.Decide(true))
	case "sb", "stop", "shobu":
		bc.writePresenterResponse(w, ki.Decide(false))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ki.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ki.Hint, ki.ActionLog)
	}
	return true
}
