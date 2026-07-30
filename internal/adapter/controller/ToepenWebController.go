//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ToepenWebInput トゥーペンWebインプット
type ToepenWebInput struct {
	BaseWebInput
	CardIndex *int             `json:"cardIndex,omitempty"`
	Stay      *bool            `json:"stay,omitempty"`
	Config    *ToepenWebConfig `json:"config,omitempty"`
}

// ToepenWebConfig トゥーペンWeb設定
type ToepenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PlayerCnt     *int `json:"playerCnt,omitempty"`
}

// ToepenWebOutputPlayer トゥーペンWebアウトプットプレイヤー
type ToepenWebOutputPlayer struct {
	ID        int              `json:"id"`
	IsHuman   bool             `json:"isHuman"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// Lives は累計失点。10 で脱落。
	Lives      int  `json:"lives"`
	Folded     bool `json:"folded"`
	Eliminated bool `json:"eliminated"`
	Hidden     bool `json:"hidden"`
}

// ToepenWebOutputHint ヒント出力
type ToepenWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Fold      *bool  `json:"fold,omitempty"`
	Reason    string `json:"reason"`
}

// ToepenWebOutput トゥーペンWebアウトプット
type ToepenWebOutput struct {
	Players          []*ToepenWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	DealerIdx        int                      `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	// LeadSuit は -1 なら未決。
	LeadSuit    int `json:"leadSuit"`
	TrickNumber int `json:"trickNumber"`
	HandNumber  int `json:"handNumber"`
	// Stake は今このハンドを落としたときの失点。ノックのたびに 1 増える。
	Stake int `json:"stake"`
	// KnockerIdx は toep を宣言した者。応答フェーズ外は -1。
	KnockerIdx int `json:"knockerIdx"`
	// PendingRespondent は追随/降参を答える番のプレイヤー。無ければ -1。
	PendingRespondent int `json:"pendingRespondent"`
	LastTrickWinner   int `json:"lastTrickWinner"`
	MaxLives          int `json:"maxLives"`
	// ValidPlayIndices は人間が出せる手札の添字。フォロー義務の判定を
	// クライアントに再実装させないために送る。
	ValidPlayIndices []int `json:"validPlayIndices"`
	// CanRedeal は人間が貧民 (A/K/Q/J のみ) の配り直しを要求できるか。
	// 判定はサーバーが持つので、クライアントは札種を数え直さない。
	CanRedeal   bool                 `json:"canRedeal"`
	GameEndFlag bool                 `json:"gameEndFlag"`
	WinnerIdx   int                  `json:"winnerIdx"`
	Hint        *ToepenWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config ToepenWebOutputConfig `json:"config"`
}

// ToepenWebOutputConfig トゥーペン設定アウトプット
type ToepenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PlayerCnt     int `json:"playerCnt"`
}

// ToConfig builds a ToepenConfig from the nested web config, applying bounds checking.
func (c *ToepenWebConfig) ToConfig() domain.ToepenConfig {
	cfg := domain.DefaultToepenConfig()
	cfg.CpuDifficulty = domain.ToepenCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty,
		int(domain.ToepenCpuDifficultyNormal), int(domain.ToepenCpuDifficultyNormal),
		int(cfg.CpuDifficulty)))
	cfg.PlayerCnt = webutil.BoundedIntPtr(c.PlayerCnt,
		domain.ToepenMinPlayers, domain.ToepenMaxPlayers, cfg.PlayerCnt)
	return cfg
}

// ToConfig builds a ToepenConfig from the input, falling back to defaults when absent.
//
// Must go through configOrDefault: `config` is optional on the wire, so a plain
// reset arrives with a nil *ToepenWebConfig and calling the method on it would
// dereference nil.
func (i ToepenWebInput) ToConfig() domain.ToepenConfig {
	return configOrDefault(i.Config, (*ToepenWebConfig).ToConfig, domain.DefaultToepenConfig())
}

// ToepenWebController トゥーペンWebコントローラ
type ToepenWebController = GameWebController[usecase.ToepenInteractorIF, ToepenWebInput, *ToepenWebOutput]

// NewToepenWebController and NewToepenWebControllerWithProvider are
// the standard and provider-backed constructors for ToepenWebController.
var NewToepenWebController, NewToepenWebControllerWithProvider = webControllerPair[usecase.ToepenInteractorIF, ToepenWebInput, *ToepenWebOutput](
	newToepenDefaultOutput, toepenDispatch,
)

func newToepenDefaultOutput(msg string) *ToepenWebOutput {
	return &ToepenWebOutput{
		Players:           make([]*ToepenWebOutputPlayer, 0),
		CurrentTrick:      make([]*WebOutputTrickCard, 0),
		ValidPlayIndices:  make([]int, 0),
		LeadSuit:          -1,
		KnockerIdx:        -1,
		PendingRespondent: -1,
		LastTrickWinner:   -1,
		WinnerIdx:         -1,
		Stake:             1,
		MaxLives:          domain.ToepenMaxLives,
		WebOutputBase:     WebOutputBase{Message: msg},
	}
}

func toepenDispatch(bc *baseController, w http.ResponseWriter, ti usecase.ToepenInteractorIF, param ToepenWebInput, newDefault func(string) *ToepenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "t", "toep":
		bc.writePresenterResponse(w, ti.Toep())
	case "a", "answer":
		if !requireParam(bc, w, newDefault, param.Stay == nil, "param error: stay is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Respond(*param.Stay))
	case "d", "redeal":
		bc.writePresenterResponse(w, ti.Redeal())
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextHand())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}

// NewToepenDefaultOutputForTest exposes the default-output builder to the
// external controller_test package.
func NewToepenDefaultOutputForTest(msg string) *ToepenWebOutput { return newToepenDefaultOutput(msg) }
