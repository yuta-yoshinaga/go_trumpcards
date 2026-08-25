//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DilotiWebInput はディロティの Web インプット。
type DilotiWebInput struct {
	BaseWebInput
	// HandIndex は出す手札のインデックス。
	HandIndex *int `json:"handIndex,omitempty"`
	// Action は手の種類 (capture / declare / trail)。
	Action string `json:"action,omitempty"`
	// TableIndices は巻き込む場札のインデックス。
	TableIndices []int `json:"tableIndices,omitempty"`
	// DeclIndices は巻き込む宣言のインデックス。
	DeclIndices []int `json:"declIndices,omitempty"`
	// DeclValue は宣言値 (declare のときだけ)。
	DeclValue *int `json:"declValue,omitempty"`
	// Config はゲーム設定。
	Config *DilotiWebConfig `json:"config,omitempty"`
}

// DilotiWebConfig はディロティの Web 設定。
type DilotiWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// DilotiWebOutputDeclaration は場に積まれた 1 つの宣言。
type DilotiWebOutputDeclaration struct {
	OwnerIdx int `json:"ownerIdx"`
	Value    int `json:"value"`
	// Groups は束ごとの札。**グループ宣言は丸ごとしか取れない**ので、束の区切りを渡す。
	Groups  [][]*WebOutputCard `json:"groups"`
	IsGroup bool               `json:"isGroup"`
}

// DilotiWebOutputPlayer は 1 席ぶんの出力。
type DilotiWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Cards は手札 (人間のみ公開)。
	Cards     []*WebOutputCard `json:"cards"`
	CardCount int              `json:"cardCount"`
	// CapturedCount はこの局で取った枚数。
	CapturedCount int `json:"capturedCount"`
	// Xeri はこの局のクセリ回数。**1 回 10 点。**
	Xeri     int  `json:"xeri"`
	Score    int  `json:"score"`
	IsDealer bool `json:"isDealer"`
}

// DilotiWebOutputTake は 1 つの取り手。
type DilotiWebOutputTake struct {
	TableIdxs []int `json:"tableIdxs"`
	DeclIdxs  []int `json:"declIdxs"`
}

// DilotiWebOutputDeclCandidate は作れる宣言の候補。
type DilotiWebOutputDeclCandidate struct {
	Value     int   `json:"value"`
	TableIdxs []int `json:"tableIdxs"`
}

// DilotiWebOutputScoreLine は 1 つの得点項目。
type DilotiWebOutputScoreLine struct {
	Key    string `json:"key"`
	Points []int  `json:"points"`
}

// DilotiWebOutputResult は 1 局の集計結果。
type DilotiWebOutputResult struct {
	Lines      []*DilotiWebOutputScoreLine `json:"lines"`
	Totals     []int                       `json:"totals"`
	CardCounts []int                       `json:"cardCounts"`
	Xeris      []int                       `json:"xeris"`
}

// DilotiWebOutput はディロティの Web アウトプット。
type DilotiWebOutput struct {
	Players          []*DilotiWebOutputPlayer      `json:"players"`
	Phase            string                        `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	DealerIdx        int                           `json:"dealerIdx"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	Table            []*WebOutputCard              `json:"table"`
	Declarations     []*DilotiWebOutputDeclaration `json:"declarations"`
	// DeckRemaining は山の残り枚数。
	DeckRemaining int `json:"deckRemaining"`
	// LastCapturer は最後に取った席。**山が尽きたときの場札はここへ渡る。**
	LastCapturer int `json:"lastCapturer"`
	// TakeOptions は人間の手札ごとの取り手 (手札の並び順)。
	TakeOptions [][]*DilotiWebOutputTake `json:"takeOptions"`
	// DeclareOptions は人間の手札ごとの宣言候補 (手札の並び順)。
	DeclareOptions [][]*DilotiWebOutputDeclCandidate `json:"declareOptions"`
	// CanTrail は人間の手札ごとに場へ置けるか (手札の並び順)。
	CanTrail    []bool                 `json:"canTrail"`
	LastResult  *DilotiWebOutputResult `json:"lastResult,omitempty"`
	GameEndFlag bool                   `json:"gameEndFlag"`
	WinnerIdx   int                    `json:"winnerIdx"`
	IsHumanTurn bool                   `json:"isHumanTurn"`
	// HintHandIdx は勧める手札 (-1 = なし)。
	HintHandIdx int `json:"hintHandIdx"`
	// HintAction は勧める手の種類。
	HintAction string `json:"hintAction"`
	// HintTableIdxs は勧める場札。
	HintTableIdxs []int `json:"hintTableIdxs"`
	// HintDeclIdxs は勧める宣言。
	HintDeclIdxs []int `json:"hintDeclIdxs"`
	// HintDeclValue は勧める宣言値。
	HintDeclValue int `json:"hintDeclValue"`
	// HintReason は理由の識別子。
	HintReason string `json:"hintReason"`
	WebOutputBase
	Config DilotiWebOutputConfig `json:"config"`
}

// DilotiWebOutputConfig は設定アウトプット。
type DilotiWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *DilotiWebConfig) ToConfig() domain.DilotiConfig {
	cfg := domain.DefaultDilotiConfig()
	cfg.CpuDifficulty = domain.DilotiCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.DilotiCpuDifficultyEasy),
		int(domain.DilotiCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore,
		domain.DilotiMinTarget, domain.DilotiMaxTarget)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p DilotiWebInput) ToConfig() domain.DilotiConfig {
	return configOrDefault(p.Config, (*DilotiWebConfig).ToConfig, domain.DefaultDilotiConfig())
}

// DilotiWebController はディロティの Web コントローラー。
type DilotiWebController = GameWebController[usecase.DilotiInteractorIF, DilotiWebInput, *DilotiWebOutput]

// NewDilotiWebController, NewDilotiWebControllerWithProvider are the standard
// and provider-backed constructors.
var NewDilotiWebController, NewDilotiWebControllerWithProvider = webControllerPair[usecase.DilotiInteractorIF, DilotiWebInput, *DilotiWebOutput](
	newDilotiDefaultOutput, dilotiDispatch,
)

func newDilotiDefaultOutput(msg string) *DilotiWebOutput {
	return &DilotiWebOutput{
		Players:        make([]*DilotiWebOutputPlayer, 0),
		Table:          make([]*WebOutputCard, 0),
		Declarations:   make([]*DilotiWebOutputDeclaration, 0),
		TakeOptions:    make([][]*DilotiWebOutputTake, 0),
		DeclareOptions: make([][]*DilotiWebOutputDeclCandidate, 0),
		CanTrail:       make([]bool, 0),
		HintTableIdxs:  make([]int, 0),
		HintDeclIdxs:   make([]int, 0),
		LastCapturer:   -1,
		WinnerIdx:      -1,
		HintHandIdx:    -1,
		WebOutputBase:  WebOutputBase{Message: msg},
	}
}

func dilotiDispatch(bc *baseController, w http.ResponseWriter, di usecase.DilotiInteractorIF, param DilotiWebInput, newDefault func(string) *DilotiWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.HandIndex == nil, "param error: handIndex is required.") {
			return true
		}
		// **取る札・宣言値は同じ要求に乗る。** 別便にすると「出したが取っていない」
		// 盤面が生まれる。
		if !requireParam(bc, w, newDefault, param.Action == "", "param error: action is required.") {
			return true
		}
		declValue := 0
		if param.DeclValue != nil {
			declValue = *param.DeclValue
		}
		bc.writePresenterResponse(w,
			di.Play(*param.HandIndex, param.Action, param.TableIndices, param.DeclIndices, declValue))
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
