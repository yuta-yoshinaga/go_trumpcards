//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DehlaPakadWebInput はデーラ・パカドの Web インプット。
type DehlaPakadWebInput struct {
	BaseWebInput
	// CardIndex は出す札のインデックス。
	CardIndex *int `json:"cardIndex,omitempty"`
	// TrumpSuit は宣言する切り札スート (1-4)。
	TrumpSuit *int `json:"trumpSuit,omitempty"`
	// Config はゲーム設定。
	Config *DehlaPakadWebConfig `json:"config,omitempty"`
}

// DehlaPakadWebConfig はデーラ・パカドの Web 設定。
type DehlaPakadWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetKots    *int `json:"targetKots,omitempty"`
}

// DehlaPakadWebOutputPlayer は 1 席ぶんの出力。
type DehlaPakadWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0 / 1。**相方は向かい**なので、隣は必ず相手。
	Team      int              `json:"team"`
	CardCount int              `json:"cardCount"`
	Cards     []*WebOutputCard `json:"cards"`
	// GatheredCount はこの席が中央の山を引き取った札の枚数。
	GatheredCount int  `json:"gatheredCount"`
	IsDealer      bool `json:"isDealer"`
	// IsTrumpChooser はこの席が切り札を決めるか。
	IsTrumpChooser bool `json:"isTrumpChooser"`
}

// DehlaPakadWebOutputHand は 1 ハンドの結果。
type DehlaPakadWebOutputHand struct {
	WinnerTeam int    `json:"winnerTeam"`
	TeamTens   []int  `json:"teamTens"`
	Kot        bool   `json:"kot"`
	KotReason  string `json:"kotReason"`
	DealerIdx  int    `json:"dealerIdx"`
	TrumpSuit  int    `json:"trumpSuit"`
}

// DehlaPakadWebOutput はデーラ・パカドの Web アウトプット。
type DehlaPakadWebOutput struct {
	Players    []*DehlaPakadWebOutputPlayer `json:"players"`
	Phase      string                       `json:"phase"`
	HandNumber int                          `json:"handNumber"`
	DealerIdx  int                          `json:"dealerIdx"`
	// TrumpChooserIdx は切り札を決める席 (親の右隣)。
	TrumpChooserIdx int `json:"trumpChooserIdx"`
	// TrumpSuit はこのハンドの切り札 (-1 = 未宣言)。
	TrumpSuit int `json:"trumpSuit"`
	// TrumpSuitName は切り札の識別子 ("spade" 等、i18n キーに使う)。
	TrumpSuitName    string                `json:"trumpSuitName"`
	TrickNumber      int                   `json:"trickNumber"`
	TrickCount       int                   `json:"trickCount"`
	CurrentPlayerIdx int                   `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                   `json:"leadPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard `json:"currentTrick"`
	LastTrick        []*WebOutputTrickCard `json:"lastTrick"`
	LastTrickWinner  int                   `json:"lastTrickWinner"`
	// PrevTrickWinner は直前のトリックを取った席。**次も取れば山ごと引き取る。**
	PrevTrickWinner int `json:"prevTrickWinner"`
	// CentrePileCount はまだ誰も引き取っていない札の枚数。
	CentrePileCount int `json:"centrePileCount"`
	// CentrePileTens はその山に乗っている 10 の枚数。
	CentrePileTens  int   `json:"centrePileTens"`
	PlayableIndices []int `json:"playableIndices"`
	// TeamTens はチーム別に取った 10 の枚数。
	TeamTens []int `json:"teamTens"`
	// TeamKots はチーム別のコート数 (これが勝敗)。
	TeamKots []int `json:"teamKots"`
	// HumanTeam は人間のチーム。
	HumanTeam   int                        `json:"humanTeam"`
	StreakTeam  int                        `json:"streakTeam"`
	StreakCount int                        `json:"streakCount"`
	LastHand    *DehlaPakadWebOutputHand   `json:"lastHand,omitempty"`
	HandHistory []*DehlaPakadWebOutputHand `json:"handHistory"`
	GameEndFlag bool                       `json:"gameEndFlag"`
	WinnerTeam  int                        `json:"winnerTeam"`
	IsHumanTurn bool                       `json:"isHumanTurn"`
	// IsTrumpPhase は切り札の宣言待ちか。
	IsTrumpPhase bool               `json:"isTrumpPhase"`
	Hint         *WebOutputCardHint `json:"hint,omitempty"`
	// HintTrumpSuit は宣言フェーズで勧める切り札 (-1 = なし)。
	HintTrumpSuit int `json:"hintTrumpSuit"`
	WebOutputBase
	Config DehlaPakadWebOutputConfig `json:"config"`
}

// DehlaPakadWebOutputConfig は設定アウトプット。
type DehlaPakadWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetKots    int `json:"targetKots"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *DehlaPakadWebConfig) ToConfig() domain.DehlaPakadConfig {
	cfg := domain.DefaultDehlaPakadConfig()
	cfg.CpuDifficulty = domain.DehlaPakadCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.DehlaPakadCpuDifficultyEasy),
		int(domain.DehlaPakadCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetKots, c.TargetKots,
		domain.DehlaPakadMinKots, domain.DehlaPakadMaxKots)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p DehlaPakadWebInput) ToConfig() domain.DehlaPakadConfig {
	return configOrDefault(p.Config, (*DehlaPakadWebConfig).ToConfig, domain.DefaultDehlaPakadConfig())
}

// DehlaPakadWebController はデーラ・パカドの Web コントローラー。
type DehlaPakadWebController = GameWebController[usecase.DehlaPakadInteractorIF, DehlaPakadWebInput, *DehlaPakadWebOutput]

// NewDehlaPakadWebController, NewDehlaPakadWebControllerWithProvider are the
// standard and provider-backed constructors.
var NewDehlaPakadWebController, NewDehlaPakadWebControllerWithProvider = webControllerPair[usecase.DehlaPakadInteractorIF, DehlaPakadWebInput, *DehlaPakadWebOutput](
	newDehlaPakadDefaultOutput, dehlaPakadDispatch,
)

func newDehlaPakadDefaultOutput(msg string) *DehlaPakadWebOutput {
	return &DehlaPakadWebOutput{
		Players:         make([]*DehlaPakadWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		TeamTens:        make([]int, 0),
		TeamKots:        make([]int, 0),
		HandHistory:     make([]*DehlaPakadWebOutputHand, 0),
		TrumpSuit:       -1,
		LastTrickWinner: -1,
		PrevTrickWinner: -1,
		StreakTeam:      -1,
		WinnerTeam:      -1,
		HintTrumpSuit:   -1,
		TrickCount:      domain.DehlaPakadTrickCount,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func dehlaPakadDispatch(bc *baseController, w http.ResponseWriter, di usecase.DehlaPakadInteractorIF, param DehlaPakadWebInput, newDefault func(string) *DehlaPakadWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "t", "trump":
		if !requireParam(bc, w, newDefault, param.TrumpSuit == nil, "param error: trumpSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.SelectTrump(*param.TrumpSuit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "nh", "nexthand":
		bc.writePresenterResponse(w, di.NextHand())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
