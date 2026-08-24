//go:build !js || !wasm || classic

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// UnsunKarutaWebInput はうんすんカルタの Web インプット。
type UnsunKarutaWebInput struct {
	BaseWebInput
	// CardIndex は出す札のインデックス。
	CardIndex *int `json:"cardIndex,omitempty"`
	// Declare はリードでメリ / モンチ を宣言するか。
	Declare *bool `json:"declare,omitempty"`
	// Config はゲーム設定。
	Config *UnsunKarutaWebConfig `json:"config,omitempty"`
}

// UnsunKarutaWebConfig はうんすんカルタの Web 設定。
type UnsunKarutaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetDeals   *int `json:"targetDeals,omitempty"`
}

// UnsunKarutaWebOutputPlayer は 1 席ぶんの出力。
type UnsunKarutaWebOutputPlayer struct {
	ID      int  `json:"id"`
	IsHuman bool `json:"isHuman"`
	// Team は 0 / 1。**敵味方が交互に座る**ので、席番号だけでは味方が分からない。
	Team       int              `json:"team"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	IsDealer   bool             `json:"isDealer"`
}

// UnsunKarutaWebOutput はうんすんカルタの Web アウトプット。
type UnsunKarutaWebOutput struct {
	Players          []*UnsunKarutaWebOutputPlayer `json:"players"`
	Phase            int                           `json:"phase"`
	RoundNumber      int                           `json:"roundNumber"`
	TrickNumber      int                           `json:"trickNumber"`
	TrickCount       int                           `json:"trickCount"`
	CurrentPlayerIdx int                           `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                           `json:"leadPlayerIdx"`
	DealerIdx        int                           `json:"dealerIdx"`
	HumanTeam        int                           `json:"humanTeam"`
	// TrumpSuit はこのディールの切り札スート (1..5)。
	TrumpSuit int `json:"trumpSuit"`
	// TrumpSuitName は切り札スートの識別子 ("pao" 等、i18n キーに使う)。
	TrumpSuitName string `json:"trumpSuitName"`
	// TrumpCard は表に返した札。
	TrumpCard *WebOutputCard `json:"trumpCard,omitempty"`
	// MustFollow はこのトリックにフォロー義務があるか。
	MustFollow bool `json:"mustFollow"`
	// Declared はこのトリックで宣言が行われたか。
	Declared bool `json:"declared"`
	// CanDeclare は人間がいま宣言できるか (リードの手番か)。
	CanDeclare      bool                  `json:"canDeclare"`
	CurrentTrick    []*WebOutputTrickCard `json:"currentTrick"`
	TeamTricks      []int                 `json:"teamTricks"`
	TeamScores      []int                 `json:"teamScores"`
	LastTrickWinner int                   `json:"lastTrickWinner"`
	Result          int                   `json:"result"`
	PlayableIndices []int                 `json:"playableIndices"`
	GameEndFlag     bool                  `json:"gameEndFlag"`
	WinnerTeam      int                   `json:"winnerTeam"`
	IsHumanTurn     bool                  `json:"isHumanTurn"`
	Hint            *WebOutputCardHint    `json:"hint,omitempty"`
	WebOutputBase
	Config UnsunKarutaWebOutputConfig `json:"config"`
}

// UnsunKarutaWebOutputConfig は設定アウトプット。
type UnsunKarutaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetDeals   int `json:"targetDeals"`
}

// ToConfig は Web 設定から domain の設定を組み立てる (境界チェック付き)。
func (c *UnsunKarutaWebConfig) ToConfig() domain.UnsunKarutaConfig {
	cfg := domain.DefaultUnsunKarutaConfig()
	cfg.CpuDifficulty = domain.UnsunKarutaCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty,
		int(domain.UnsunKarutaCpuDifficultyEasy),
		int(domain.UnsunKarutaCpuDifficultyHard),
		int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetDeals, c.TargetDeals,
		domain.UnsunKarutaMinDeals, domain.UnsunKarutaMaxDeals)
	return cfg
}

// ToConfig は Web インプットから domain の設定を組み立てる。
func (p UnsunKarutaWebInput) ToConfig() domain.UnsunKarutaConfig {
	return configOrDefault(p.Config, (*UnsunKarutaWebConfig).ToConfig, domain.DefaultUnsunKarutaConfig())
}

// UnsunKarutaWebController はうんすんカルタの Web コントローラー。
type UnsunKarutaWebController = GameWebController[usecase.UnsunKarutaInteractorIF, UnsunKarutaWebInput, *UnsunKarutaWebOutput]

// NewUnsunKarutaWebController, NewUnsunKarutaWebControllerWithProvider are the
// standard and provider-backed constructors.
var NewUnsunKarutaWebController, NewUnsunKarutaWebControllerWithProvider = webControllerPair[usecase.UnsunKarutaInteractorIF, UnsunKarutaWebInput, *UnsunKarutaWebOutput](
	newUnsunKarutaDefaultOutput, unsunKarutaDispatch,
)

func newUnsunKarutaDefaultOutput(msg string) *UnsunKarutaWebOutput {
	return &UnsunKarutaWebOutput{
		Players:         make([]*UnsunKarutaWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		TeamTricks:      make([]int, 0),
		TeamScores:      make([]int, 0),
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func unsunKarutaDispatch(bc *baseController, w http.ResponseWriter, di usecase.UnsunKarutaInteractorIF, param UnsunKarutaWebInput, newDefault func(string) *UnsunKarutaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		// **宣言は札と一緒に届く。** 別命令にすると「宣言したが札を出していない」
		// 盤面が生まれる。省略時は宣言なし。
		declare := param.Declare != nil && *param.Declare
		bc.writePresenterResponse(w, di.Play(*param.CardIndex, declare))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
