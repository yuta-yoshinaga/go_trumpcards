//go:build !js || !wasm || solo

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CometWebInput はコメットの Web インプット。
type CometWebInput struct {
	BaseWebInput
	// HandIndex は出す手札のインデックス。
	HandIndex *int `json:"handIndex,omitempty"`
	// Config はゲーム設定。
	Config *CometWebConfig `json:"config,omitempty"`
}

// CometWebConfig はコメットの Web 設定。
type CometWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	Players       *int `json:"players,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// CometWebOutputPlayer は 1 席ぶんの出力。
type CometWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Cards は手札 (人間のみ公開)。
	Cards     []*WebOutputCard `json:"cards"`
	CardCount int              `json:"cardCount"`
	Score     int              `json:"score"`
	IsDealer  bool             `json:"isDealer"`
}

// CometWebOutputResult は 1 局の集計結果。
type CometWebOutputResult struct {
	WinnerIdx int   `json:"winnerIdx"`
	CardsLeft []int `json:"cardsLeft"`
	Gained    []int `json:"gained"`
	// UnplayedKings は出なかった K の枚数 (1 枚 2 点)。
	UnplayedKings int `json:"unplayedKings"`
	// HeldWildIdx はコメットを抱えたまま終わった席 (-1 = なし)。
	HeldWildIdx int `json:"heldWildIdx"`
}

// CometWebOutput はコメットの Web アウトプット。
type CometWebOutput struct {
	Players     []*CometWebOutputPlayer `json:"players"`
	Phase       string                  `json:"phase"`
	RoundNumber int                     `json:"roundNumber"`
	DealerIdx   int                     `json:"dealerIdx"`
	// CurrentPlayerIdx は手番の席。
	CurrentPlayerIdx int `json:"currentPlayerIdx"`
	// Pile は今の連なりに出た札。
	Pile []*WebOutputCard `json:"pile"`
	// Need は次に要るランク (0 = 連なりの先頭で何でも出せる)。
	Need int `json:"need"`
	// DeadCount は伏せた死に手の枚数。**中身は見えない。**
	DeadCount int `json:"deadCount"`
	// LastPlayerIdx は最後に札を出した席。ストップのあとここから再開する。
	LastPlayerIdx int `json:"lastPlayerIdx"`
	// PlayableIdxs は人間が出せる手札の位置。空ならパスするしかない。
	PlayableIdxs []int                 `json:"playableIdxs"`
	LastResult   *CometWebOutputResult `json:"lastResult,omitempty"`
	GameEndFlag  bool                  `json:"gameEndFlag"`
	WinnerIdx    int                   `json:"winnerIdx"`
	IsHumanTurn  bool                  `json:"isHumanTurn"`
	// HintHandIdx は勧める手札 (-1 = なし)。
	HintHandIdx int `json:"hintHandIdx"`
	// HintReason は理由の識別子。
	HintReason string `json:"hintReason"`
	WebOutputBase
	Config CometWebOutputConfig `json:"config"`
}

// CometWebOutputConfig は設定アウトプット。
type CometWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	Players       int `json:"players"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *CometWebConfig) ToConfig() domain.CometConfig {
	cfg := domain.DefaultCometConfig()
	cfg.CpuDifficulty = domain.CometCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.CometCpuDifficultyEasy),
		int(domain.CometCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.Players, c.Players,
		domain.CometMinPlayers, domain.CometMaxPlayers)
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore,
		domain.CometMinTarget, domain.CometMaxTarget)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p CometWebInput) ToConfig() domain.CometConfig {
	return configOrDefault(p.Config, (*CometWebConfig).ToConfig, domain.DefaultCometConfig())
}

// CometWebController はコメットの Web コントローラー。
type CometWebController = GameWebController[usecase.CometInteractorIF, CometWebInput, *CometWebOutput]

// NewCometWebController, NewCometWebControllerWithProvider are the standard
// and provider-backed constructors.
var NewCometWebController, NewCometWebControllerWithProvider = webControllerPair[usecase.CometInteractorIF, CometWebInput, *CometWebOutput](
	newCometDefaultOutput, cometDispatch,
)

func newCometDefaultOutput(msg string) *CometWebOutput {
	return &CometWebOutput{
		Players:       make([]*CometWebOutputPlayer, 0),
		Pile:          make([]*WebOutputCard, 0),
		PlayableIdxs:  make([]int, 0),
		LastPlayerIdx: -1,
		WinnerIdx:     -1,
		HintHandIdx:   -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func cometDispatch(bc *baseController, w http.ResponseWriter, di usecase.CometInteractorIF, param CometWebInput, newDefault func(string) *CometWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.HandIndex == nil, "param error: handIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.HandIndex))
	case "pass":
		bc.writePresenterResponse(w, di.Pass())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
