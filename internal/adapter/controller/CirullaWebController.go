//go:build !js || !wasm || extra3

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CirullaWebInput はチルッラの Web インプット。
type CirullaWebInput struct {
	BaseWebInput
	// HandIndex は出す手札のインデックス。
	HandIndex *int `json:"handIndex,omitempty"`
	// CaptureIndices は取る場札のインデックス (空なら場に置く)。
	CaptureIndices []int `json:"captureIndices,omitempty"`
	// Config はゲーム設定。
	Config *CirullaWebConfig `json:"config,omitempty"`
}

// CirullaWebConfig はチルッラの Web 設定。
type CirullaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// CirullaWebOutputPlayer は 1 席ぶんの出力。
type CirullaWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Cards は手札 (人間のみ公開)。
	Cards     []*WebOutputCard `json:"cards"`
	CardCount int              `json:"cardCount"`
	// CapturedCount はこのラウンドで取った枚数。
	CapturedCount int `json:"capturedCount"`
	// DenariCount は取ったデナリの枚数。
	DenariCount int `json:"denariCount"`
	// HasSetteBello は 7♦ を取ったか。
	HasSetteBello bool `json:"hasSetteBello"`
	Scope         int  `json:"scope"`
	BonusPoints   int  `json:"bonusPoints"`
	Score         int  `json:"score"`
	IsDealer      bool `json:"isDealer"`
	// LastBonus は直近の配札ボーナス識別子 ("" = 無し)。
	LastBonus string `json:"lastBonus"`
}

// CirullaWebOutputScoreLine は 1 つの得点項目。
type CirullaWebOutputScoreLine struct {
	Key    string `json:"key"`
	Points []int  `json:"points"`
}

// CirullaWebOutputResult は 1 ラウンドの集計結果。
type CirullaWebOutputResult struct {
	Lines  []*CirullaWebOutputScoreLine `json:"lines"`
	Totals []int                        `json:"totals"`
	// SweptDenari は全デナリを取った席 (-1 = なし)。
	SweptDenari int `json:"sweptDenari"`
}

// CirullaWebOutput はチルッラの Web アウトプット。
type CirullaWebOutput struct {
	Players     []*CirullaWebOutputPlayer `json:"players"`
	Phase       string                    `json:"phase"`
	RoundNumber int                       `json:"roundNumber"`
	DealerIdx   int                       `json:"dealerIdx"`
	// CurrentPlayerIdx は手番の席。
	CurrentPlayerIdx int              `json:"currentPlayerIdx"`
	Table            []*WebOutputCard `json:"table"`
	// DeckRemaining は山の残り枚数。
	DeckRemaining int `json:"deckRemaining"`
	// LastCapturer は最後に捕獲した席。**ラウンド末の場札はここへ渡る。**
	LastCapturer int `json:"lastCapturer"`
	// CaptureOptions は人間の手札ごとの捕獲候補 (手札の並び順)。
	CaptureOptions [][][]int               `json:"captureOptions"`
	LastResult     *CirullaWebOutputResult `json:"lastResult,omitempty"`
	GameEndFlag    bool                    `json:"gameEndFlag"`
	WinnerIdx      int                     `json:"winnerIdx"`
	IsHumanTurn    bool                    `json:"isHumanTurn"`
	// HintHandIdx は勧める手札 (-1 = なし)。
	HintHandIdx int `json:"hintHandIdx"`
	// HintCaptureIdxs は勧める捕獲対象。
	HintCaptureIdxs []int `json:"hintCaptureIdxs"`
	// HintReason は理由の識別子。
	HintReason string `json:"hintReason"`
	WebOutputBase
	Config CirullaWebOutputConfig `json:"config"`
}

// CirullaWebOutputConfig は設定アウトプット。
type CirullaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *CirullaWebConfig) ToConfig() domain.CirullaConfig {
	cfg := domain.DefaultCirullaConfig()
	cfg.CpuDifficulty = domain.CirullaCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.CirullaCpuDifficultyEasy),
		int(domain.CirullaCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore,
		domain.CirullaMinTarget, domain.CirullaMaxTarget)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p CirullaWebInput) ToConfig() domain.CirullaConfig {
	return configOrDefault(p.Config, (*CirullaWebConfig).ToConfig, domain.DefaultCirullaConfig())
}

// CirullaWebController はチルッラの Web コントローラー。
type CirullaWebController = GameWebController[usecase.CirullaInteractorIF, CirullaWebInput, *CirullaWebOutput]

// NewCirullaWebController, NewCirullaWebControllerWithProvider are the standard
// and provider-backed constructors.
var NewCirullaWebController, NewCirullaWebControllerWithProvider = webControllerPair[usecase.CirullaInteractorIF, CirullaWebInput, *CirullaWebOutput](
	newCirullaDefaultOutput, cirullaDispatch,
)

func newCirullaDefaultOutput(msg string) *CirullaWebOutput {
	return &CirullaWebOutput{
		Players:         make([]*CirullaWebOutputPlayer, 0),
		Table:           make([]*WebOutputCard, 0),
		CaptureOptions:  make([][][]int, 0),
		HintCaptureIdxs: make([]int, 0),
		LastCapturer:    -1,
		WinnerIdx:       -1,
		HintHandIdx:     -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func cirullaDispatch(bc *baseController, w http.ResponseWriter, di usecase.CirullaInteractorIF, param CirullaWebInput, newDefault func(string) *CirullaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.HandIndex == nil, "param error: handIndex is required.") {
			return true
		}
		// **取る札は同じ要求に乗る。** 別便にすると「出したが取っていない」
		// 盤面が生まれる。省略時は場に置く。
		bc.writePresenterResponse(w, di.Play(*param.HandIndex, param.CaptureIndices))
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
