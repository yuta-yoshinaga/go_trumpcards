//go:build !js || !wasm || extra2

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SutdaWebInput はソッタの Web インプット。
type SutdaWebInput struct {
	BaseWebInput
	// Action は打つ手 (call / raise / fold)。
	Action *string `json:"action,omitempty"`
	// Config はゲーム設定。
	Config *SutdaWebConfig `json:"config,omitempty"`
}

// SutdaWebConfig はソッタの Web 設定。
type SutdaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	Seats         *int `json:"seats,omitempty"`
	StartChips    *int `json:"startChips,omitempty"`
}

// SutdaWebOutputPlayer は 1 席ぶんの出力。
type SutdaWebOutputPlayer struct {
	ID        int  `json:"id"`
	IsHuman   bool `json:"isHuman"`
	CardCount int  `json:"cardCount"`
	// Cards は見えている手札。**伏せているうちは自分のぶんだけ。**
	Cards []*WebOutputCard `json:"cards"`
	Chips int              `json:"chips"`
	Bet   int              `json:"bet"`
	// Folded は降りたか。
	Folded bool `json:"folded"`
	// Revealed はショーダウンで開いたか。
	Revealed bool `json:"revealed"`
	// HandName は開いている席の役の識別子 (i18n キー用)。伏せていれば空。
	HandName string `json:"handName"`
	// HandRank は開いている席の役の強さ。伏せていれば 0。
	HandRank int  `json:"handRank"`
	IsDealer bool `json:"isDealer"`
}

// SutdaWebOutputResult は 1 ハンドの結果。
type SutdaWebOutputResult struct {
	Winners []int `json:"winners"`
	Pot     int   `json:"pot"`
	// HandNames は席ごとの役の識別子。
	HandNames []string `json:"handNames"`
	Folded    []bool   `json:"folded"`
}

// SutdaWebOutput はソッタの Web アウトプット。
type SutdaWebOutput struct {
	Players    []*SutdaWebOutputPlayer `json:"players"`
	Phase      string                  `json:"phase"`
	HandNumber int                     `json:"handNumber"`
	DealerIdx  int                     `json:"dealerIdx"`
	// CurrentPlayerIdx は手番の席。
	CurrentPlayerIdx int `json:"currentPlayerIdx"`
	Pot              int `json:"pot"`
	CurrentBet       int `json:"currentBet"`
	// CallAmount は人間がコールに要る額 (0 ならチェック)。
	CallAmount int `json:"callAmount"`
	// CanRaise は人間がいまレイズできるか。**上限と有り金の両方を見た結果。**
	CanRaise   bool `json:"canRaise"`
	RaiseCount int  `json:"raiseCount"`
	MaxRaises  int  `json:"maxRaises"`
	BetUnit    int  `json:"betUnit"`
	// HumanHandName は人間の役の識別子 (自分のぶんは常に見える)。
	HumanHandName string                `json:"humanHandName"`
	LastResult    *SutdaWebOutputResult `json:"lastResult,omitempty"`
	GameEndFlag   bool                  `json:"gameEndFlag"`
	WinnerIdx     int                   `json:"winnerIdx"`
	IsHumanTurn   bool                  `json:"isHumanTurn"`
	// IsShowdown はショーダウン中か。
	IsShowdown bool `json:"isShowdown"`
	// HintAction は勧める行動 (空 = なし)。
	HintAction string `json:"hintAction"`
	// HintReason は理由の識別子。
	HintReason string `json:"hintReason"`
	WebOutputBase
	Config SutdaWebOutputConfig `json:"config"`
}

// SutdaWebOutputConfig は設定アウトプット。
type SutdaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	Seats         int `json:"seats"`
	StartChips    int `json:"startChips"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *SutdaWebConfig) ToConfig() domain.SutdaConfig {
	cfg := domain.DefaultSutdaConfig()
	cfg.CpuDifficulty = domain.SutdaCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.SutdaCpuDifficultyEasy),
		int(domain.SutdaCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.Seats, c.Seats, domain.SutdaMinSeats, domain.SutdaMaxSeats)
	webutil.ApplyBoundedInt(&cfg.StartChips, c.StartChips, domain.SutdaMinChips, domain.SutdaMaxChips)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p SutdaWebInput) ToConfig() domain.SutdaConfig {
	return configOrDefault(p.Config, (*SutdaWebConfig).ToConfig, domain.DefaultSutdaConfig())
}

// SutdaWebController はソッタの Web コントローラー。
type SutdaWebController = GameWebController[usecase.SutdaInteractorIF, SutdaWebInput, *SutdaWebOutput]

// NewSutdaWebController, NewSutdaWebControllerWithProvider are the standard and
// provider-backed constructors.
var NewSutdaWebController, NewSutdaWebControllerWithProvider = webControllerPair[usecase.SutdaInteractorIF, SutdaWebInput, *SutdaWebOutput](
	newSutdaDefaultOutput, sutdaDispatch,
)

func newSutdaDefaultOutput(msg string) *SutdaWebOutput {
	return &SutdaWebOutput{
		Players:       make([]*SutdaWebOutputPlayer, 0),
		WinnerIdx:     -1,
		MaxRaises:     domain.SutdaMaxRaises,
		BetUnit:       domain.SutdaBetUnit,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func sutdaDispatch(bc *baseController, w http.ResponseWriter, di usecase.SutdaInteractorIF, param SutdaWebInput, newDefault func(string) *SutdaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "a", "action":
		if !requireParam(bc, w, newDefault, param.Action == nil, "param error: action is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Action(*param.Action))
	// **よく使う 3 手は名前でも受ける。** call/raise/fold を毎回 action で
	// 包ませると、クライアント側の分岐が 1 段深くなるだけで得が無い。
	case "call", "raise", "fold":
		bc.writePresenterResponse(w, di.Action(param.Command))
	case "nh", "nexthand":
		bc.writePresenterResponse(w, di.NextHand())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
