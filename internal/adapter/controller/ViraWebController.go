//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ViraWebInput ヴィーラのWebインプット
type ViraWebInput struct {
	BaseWebInput
	// Bid 入札種別 (0=Pass,1=Six,2=Misère,3=Seven,4=Eight)
	Bid *int `json:"bid,omitempty"`
	// CardIndex プレイするカードのインデックス (play コマンド)
	CardIndex *int `json:"cardIndex,omitempty"`
	// Config ゲーム設定
	Config *ViraWebConfig `json:"config,omitempty"`
}

// ViraWebConfig ヴィーラのWeb設定
type ViraWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetRounds  *int `json:"targetRounds,omitempty"`
}

// ViraWebOutputPlayer ヴィーラのWebアウトプットプレイヤー
type ViraWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Score      int              `json:"score"`
	IsDeclarer bool             `json:"isDeclarer"`
}

// ViraWebOutput ヴィーラのWebアウトプット
type ViraWebOutput struct {
	Players          []*ViraWebOutputPlayer `json:"players"`
	Phase            int                    `json:"phase"`
	RoundNumber      int                    `json:"roundNumber"`
	TrickNumber      int                    `json:"trickNumber"`
	CurrentPlayerIdx int                    `json:"currentPlayerIdx"`
	LeadPlayerIdx    int                    `json:"leadPlayerIdx"`
	DealerIdx        int                    `json:"dealerIdx"`
	DeclarerIdx      int                    `json:"declarerIdx"`
	Contract         int                    `json:"contract"`
	TrumpSuit        int                    `json:"trumpSuit"`
	// Pot 現在のポット。**次局へ持ち越されるので画面に出す。**これが見えないと、
	// 同じ契約でも見返りが違う理由がプレイヤーに分からない。
	Pot             int                       `json:"pot"`
	LastRoundDelta  [domain.ViraPlayerCnt]int `json:"lastRoundDelta"`
	LastRoundMade   bool                      `json:"lastRoundMade"`
	Bids            [domain.ViraPlayerCnt]int `json:"bids"`
	CurrentTrick    []*WebOutputTrickCard     `json:"currentTrick"`
	PlayerScores    [domain.ViraPlayerCnt]int `json:"playerScores"`
	RoundTricks     [domain.ViraPlayerCnt]int `json:"roundTricks"`
	PlayableIndices []int                     `json:"playableIndices"`
	GameEndFlag     bool                      `json:"gameEndFlag"`
	WinnerPlayer    int                       `json:"winnerPlayer"`
	IsHumanTurn     bool                      `json:"isHumanTurn"`
	IsHumanBidTurn  bool                      `json:"isHumanBidTurn"`
	Hint            *WebOutputCardHint        `json:"hint,omitempty"`
	WebOutputBase
	Config ViraWebOutputConfig `json:"config"`
}

// ViraWebOutputConfig ヴィーラの設定アウトプット
type ViraWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetRounds  int `json:"targetRounds"`
}

// ToConfig builds a ViraConfig from the nested web config, applying bounds checking.
func (c *ViraWebConfig) ToConfig() domain.ViraConfig {
	cfg := domain.DefaultViraConfig()
	cfg.CpuDifficulty = domain.ViraCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.ViraCpuDifficultyEasy), int(domain.ViraCpuDifficultyHard), int(cfg.CpuDifficulty)))
	// **下限は ViraPlayerCnt。**1 を許すと境界検査は通るが ViraConfig.Validate が
	// 落とすので、リセットが黙って無視される。倍数条件は Validate 側が見る。
	webutil.ApplyBoundedInt(&cfg.TargetRounds, c.TargetRounds, domain.ViraPlayerCnt, 1000000)
	return cfg
}

// ToConfig builds a ViraConfig from the web input.
func (p ViraWebInput) ToConfig() domain.ViraConfig {
	return configOrDefault(p.Config, (*ViraWebConfig).ToConfig, domain.DefaultViraConfig())
}

// ViraWebController ヴィーラのWebコントローラークラス
type ViraWebController = GameWebController[usecase.ViraInteractorIF, ViraWebInput, *ViraWebOutput]

// NewViraWebController and NewViraWebControllerWithProvider are
// the standard and provider-backed constructors for ViraWebController.
var NewViraWebController, NewViraWebControllerWithProvider = webControllerPair[usecase.ViraInteractorIF, ViraWebInput, *ViraWebOutput](
	newViraDefaultOutput, viraDispatch,
)

func newViraDefaultOutput(msg string) *ViraWebOutput {
	return &ViraWebOutput{
		Players:         make([]*ViraWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		DeclarerIdx:     -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func viraDispatch(bc *baseController, w http.ResponseWriter, di usecase.ViraInteractorIF, param ViraWebInput, newDefault func(string) *ViraWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(*param.Bid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, di.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, di.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, di.Hint, di.ActionLog)
	}
	return true
}
