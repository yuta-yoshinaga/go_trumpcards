//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BauernschnapsenWebInput バウエルンシュナプセンWebインプット
type BauernschnapsenWebInput struct {
	BaseWebInput
	CardIndex *int `json:"cardIndex,omitempty"`
	// Contract は契約フェーズの宣言 (0=パス 1=通常 2=同スート縛り 3=ベテル)、
	// TrumpSuit はそれに添える切り札スート。ベテルは切り札を取らない。
	Contract  *int                      `json:"contract,omitempty"`
	TrumpSuit *int                      `json:"trumpSuit,omitempty"`
	Config    *BauernschnapsenWebConfig `json:"config,omitempty"`
}

// BauernschnapsenWebConfig バウエルンシュナプセンWeb設定
type BauernschnapsenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// BauernschnapsenWebOutputPlayer バウエルンシュナプセンWebアウトプットプレイヤー
type BauernschnapsenWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	Team       int              `json:"team"`
	TrickCount int              `json:"trickCount"`
}

// BauernschnapsenWebOutputHint ヒント出力
type BauernschnapsenWebOutputHint struct {
	CardIndex  *int   `json:"cardIndex,omitempty"`
	Reason     string `json:"reason"`
	IsMarriage bool   `json:"isMarriage"`
}

// BauernschnapsenWebOutput バウエルンシュナプセンWebアウトプット
type BauernschnapsenWebOutput struct {
	Players          []*BauernschnapsenWebOutputPlayer `json:"players"`
	Phase            int                               `json:"phase"`
	RoundNumber      int                               `json:"roundNumber"`
	TrickNumber      int                               `json:"trickNumber"`
	CurrentPlayerIdx int                               `json:"currentPlayerIdx"`
	DealerIdx        int                               `json:"dealerIdx"`
	TrumpSuit        int                               `json:"trumpSuit"`
	// Contract は採用された契約、DeclarerIdx はそれを宣言した席 (-1 = 未確定)。
	// クローン元にあった trumpCard / stockRemaining は、20 枚を配り切って
	// 山札も表向きの切り札表示カードも持たないこのゲームには存在しない。
	Contract      int                   `json:"contract"`
	DeclarerIdx   int                   `json:"declarerIdx"`
	CurrentTrick  []*WebOutputTrickCard `json:"currentTrick"`
	TeamScores    [2]int                `json:"teamScores"`
	RoundPoints   [2]int                `json:"roundPoints"`
	RoundMarriage [2]int                `json:"roundMarriage"`
	// ValidPlayIndices は現手番の席が出せる手札位置。**追従はトリック 1 から必須**
	// なので、画面はこれを見ないと出せない札を押せてしまう。
	ValidPlayIndices []int                         `json:"validPlayIndices"`
	MarriageIndices  []int                         `json:"marriageIndices"`
	GameEndFlag      bool                          `json:"gameEndFlag"`
	WinnerTeam       int                           `json:"winnerTeam"`
	LeadPlayerIdx    int                           `json:"leadPlayerIdx"`
	Hint             *BauernschnapsenWebOutputHint `json:"hint,omitempty"`
	WebOutputBase
	Config BauernschnapsenWebOutputConfig `json:"config"`
}

// BauernschnapsenWebOutputConfig バウエルンシュナプセン設定アウトプット
type BauernschnapsenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a BauernschnapsenConfig from the nested web config, applying bounds checking.
func (c *BauernschnapsenWebConfig) ToConfig() domain.BauernschnapsenConfig {
	cfg := domain.DefaultBauernschnapsenConfig()
	cfg.CpuDifficulty = domain.BauernschnapsenCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.BauernschnapsenCpuDifficultyEasy), int(domain.BauernschnapsenCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 1, 10000)
	return cfg
}

// ToConfig builds a BauernschnapsenConfig from the web input.
func (p BauernschnapsenWebInput) ToConfig() domain.BauernschnapsenConfig {
	return configOrDefault(p.Config, (*BauernschnapsenWebConfig).ToConfig, domain.DefaultBauernschnapsenConfig())
}

// BauernschnapsenWebController バウエルンシュナプセンWebコントローラークラス
type BauernschnapsenWebController = GameWebController[usecase.BauernschnapsenInteractorIF, BauernschnapsenWebInput, *BauernschnapsenWebOutput]

// NewBauernschnapsenWebController and NewBauernschnapsenWebControllerWithProvider are
// the standard and provider-backed constructors for BauernschnapsenWebController.
var NewBauernschnapsenWebController, NewBauernschnapsenWebControllerWithProvider = webControllerPair[usecase.BauernschnapsenInteractorIF, BauernschnapsenWebInput, *BauernschnapsenWebOutput](
	newBauernschnapsenDefaultOutput, bauernschnapsenDispatch,
)

func newBauernschnapsenDefaultOutput(msg string) *BauernschnapsenWebOutput {
	return &BauernschnapsenWebOutput{
		Players:          make([]*BauernschnapsenWebOutputPlayer, 0),
		CurrentTrick:     make([]*WebOutputTrickCard, 0),
		ValidPlayIndices: make([]int, 0),
		MarriageIndices:  make([]int, 0),
		WinnerTeam:       -1,
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

func bauernschnapsenDispatch(bc *baseController, w http.ResponseWriter, gi usecase.BauernschnapsenInteractorIF, param BauernschnapsenWebInput, newDefault func(string) *BauernschnapsenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, gi.ResetWithConfig(param.ToConfig()))
	case "c", "contract":
		if !requireParam(bc, w, newDefault, param.Contract == nil, "param error: contract is required.") {
			return true
		}
		suit := domain.CardDesignSpade
		if param.TrumpSuit != nil {
			suit = *param.TrumpSuit
		}
		bc.writePresenterResponse(w, gi.DeclareContract(domain.BauernschnapsenContract(*param.Contract), suit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.Play(*param.CardIndex))
	case "m", "marriage":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.DeclareMarriage(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, gi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, gi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, gi.Hint, gi.ActionLog)
	}
	return true
}
