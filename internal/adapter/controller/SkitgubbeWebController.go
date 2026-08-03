//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SkitgubbeWebInput シートグッベWebインプット
type SkitgubbeWebInput struct {
	BaseWebInput
	CardIndex *int                `json:"cardIndex,omitempty"`
	Config    *SkitgubbeWebConfig `json:"config,omitempty"`
}

// SkitgubbeWebConfig シートグッベWeb設定
type SkitgubbeWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// SkitgubbeWebOutputPlayer シートグッベWebアウトプットプレイヤー
type SkitgubbeWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// CardCount は手札の枚数。伏せている間も送る -- 何枚持っているかは公開情報で、
	// 「誰が上がりに近いか」はこのゲームの唯一の読み筋。
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// CollectedCount は第1フェーズで集めた枚数。集めた札はそのまま第2フェーズの
	// 手札になるので、**取りすぎている人が不利**という判断材料になる。
	CollectedCount int  `json:"collectedCount"`
	Finished       bool `json:"finished"`
	Hidden         bool `json:"hidden"`
}

// SkitgubbeWebOutputHint ヒント出力
type SkitgubbeWebOutputHint struct {
	CardIndex *int `json:"cardIndex,omitempty"`
	// PickUp が真なら「引き取る」のが推奨手。
	PickUp bool   `json:"pickUp"`
	Reason string `json:"reason"`
}

// SkitgubbeWebOutput シートグッベWebアウトプット
type SkitgubbeWebOutput struct {
	Players          []*SkitgubbeWebOutputPlayer `json:"players"`
	Phase            int                         `json:"phase"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	StockCount       int                         `json:"stockCount"`
	// TrumpSuit は切札のスート。**山札から最後に引かれた札**で決まるので、
	// 第1フェーズの途中までは -1 (未確定)。
	TrumpSuit int `json:"trumpSuit"`
	// Duel は第1フェーズで場に出ている札。stunsa (同ランクのバウンス) が続くと
	// 2 枚ずつ積み上がる。
	Duel       []*WebOutputCard `json:"duel"`
	DuelLeader int              `json:"duelLeader"`
	// Pile は第2フェーズで場に出ている札。上回れなかった人が引き取る。
	Pile []*WebOutputCard `json:"pile"`
	// ValidIndices は出せる手札の添字。「直前の札を上回る (同スートの上位か
	// 切札)」の判定をクライアントに再実装させない。
	ValidIndices []int `json:"validIndices"`
	// CanPickUp が真なら「引き取る」ボタンを押せる。出せる札があるうちは
	// 引き取れない ("It is never lawful to duck")。
	CanPickUp   bool                    `json:"canPickUp"`
	GameEndFlag bool                    `json:"gameEndFlag"`
	LoserIdx    int                     `json:"loserIdx"`
	Hint        *SkitgubbeWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config SkitgubbeWebOutputConfig `json:"config"`
}

// SkitgubbeWebOutputConfig シートグッベ設定アウトプット
type SkitgubbeWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// ToConfig builds a SkitgubbeConfig from the nested web config, applying bounds checking.
func (c *SkitgubbeWebConfig) ToConfig() domain.SkitgubbeConfig {
	cfg := domain.DefaultSkitgubbeConfig()
	cfg.CpuDifficulty = domain.SkitgubbeCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.SkitgubbeCpuDifficultyNormal), int(domain.SkitgubbeCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	return cfg
}

// ToConfig builds a SkitgubbeConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *SkitgubbeWebConfig and calling the method on it
// would dereference nil.
func (i SkitgubbeWebInput) ToConfig() domain.SkitgubbeConfig {
	return configOrDefault(i.Config, (*SkitgubbeWebConfig).ToConfig, domain.DefaultSkitgubbeConfig())
}

// SkitgubbeWebController シートグッベWebコントローラ
type SkitgubbeWebController = GameWebController[usecase.SkitgubbeInteractorIF, SkitgubbeWebInput, *SkitgubbeWebOutput]

// NewSkitgubbeWebController and NewSkitgubbeWebControllerWithProvider are
// the standard and provider-backed constructors for SkitgubbeWebController.
var NewSkitgubbeWebController, NewSkitgubbeWebControllerWithProvider = webControllerPair[usecase.SkitgubbeInteractorIF, SkitgubbeWebInput, *SkitgubbeWebOutput](
	newSkitgubbeDefaultOutput, skitgubbeDispatch,
)

func newSkitgubbeDefaultOutput(msg string) *SkitgubbeWebOutput {
	return &SkitgubbeWebOutput{
		Players:       make([]*SkitgubbeWebOutputPlayer, 0),
		Duel:          make([]*WebOutputCard, 0),
		Pile:          make([]*WebOutputCard, 0),
		ValidIndices:  make([]int, 0),
		TrumpSuit:     -1,
		LoserIdx:      -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func skitgubbeDispatch(bc *baseController, w http.ResponseWriter, si usecase.SkitgubbeInteractorIF, param SkitgubbeWebInput, newDefault func(string) *SkitgubbeWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "u", "pickup":
		bc.writePresenterResponse(w, si.PickUp())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}

// NewSkitgubbeDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
func NewSkitgubbeDefaultOutputForTest(msg string) *SkitgubbeWebOutput {
	return newSkitgubbeDefaultOutput(msg)
}
