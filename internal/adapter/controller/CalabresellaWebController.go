//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CalabresellaWebInput カラブレセッラ (Calabresella) のWebインプット
type CalabresellaWebInput struct {
	BaseWebInput
	// CardIndex プレイ/ディスカードするカードのインデックス
	CardIndex *int `json:"cardIndex,omitempty"`
	// Bid ビッド宣言 (0=pass, 1=chiamo, 2=solo)
	Bid *int `json:"bid,omitempty"`
	// Config ゲーム設定
	Config *CalabresellaWebConfig `json:"config,omitempty"`
}

// CalabresellaWebConfig カラブレセッラのWeb設定
type CalabresellaWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetPoints  *int `json:"targetPoints,omitempty"`
}

// CalabresellaWebOutputPlayer カラブレセッラのWebアウトプットプレイヤー
type CalabresellaWebOutputPlayer struct {
	ID          int              `json:"id"`
	IsHuman     bool             `json:"isHuman"`
	CardCount   int              `json:"cardCount"`
	Cards       []*WebOutputCard `json:"cards"`
	TrickCount  int              `json:"trickCount"`
	Score       int              `json:"score"`
	IsSoloist   bool             `json:"isSoloist"`
	RoundThirds int              `json:"roundThirds"`
}

// CalabresellaWebOutputTrickCard トリック中の1枚
type CalabresellaWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// CalabresellaWebOutputHint ヒント出力
type CalabresellaWebOutputHint struct {
	CardIndices []int  `json:"cardIndices"`
	Reason      string `json:"reason"`
}

// CalabresellaWebOutput カラブレセッラのWebアウトプット
type CalabresellaWebOutput struct {
	Players          []*CalabresellaWebOutputPlayer    `json:"players"`
	Phase            int                               `json:"phase"`
	RoundNumber      int                               `json:"roundNumber"`
	TrickNumber      int                               `json:"trickNumber"`
	CurrentPlayerIdx int                               `json:"currentPlayerIdx"`
	CurrentBidderIdx int                               `json:"currentBidderIdx"`
	LeadPlayerIdx    int                               `json:"leadPlayerIdx"`
	DealerIdx        int                               `json:"dealerIdx"`
	ForehandIdx      int                               `json:"forehandIdx"`
	SoloistIdx       int                               `json:"soloistIdx"`
	WinningBid       int                               `json:"winningBid"`
	CurrentTrick     []*CalabresellaWebOutputTrickCard `json:"currentTrick"`
	Monte            []*WebOutputCard                  `json:"monte,omitempty"`
	PlayerScores     [domain.CalabresellaPlayerCnt]int `json:"playerScores"`
	RoundThirds      [domain.CalabresellaPlayerCnt]int `json:"roundThirds"`
	LastTrickWinner  int                               `json:"lastTrickWinner"`
	PlayableIndices  []int                             `json:"playableIndices"`
	GameEndFlag      bool                              `json:"gameEndFlag"`
	WinnerPlayer     int                               `json:"winnerPlayer"`
	IsHumanTurn      bool                              `json:"isHumanTurn"`
	Hint             *CalabresellaWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config CalabresellaWebOutputConfig `json:"config"`
}

// CalabresellaWebOutputConfig カラブレセッラの設定アウトプット
type CalabresellaWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetPoints  int `json:"targetPoints"`
}

// ToConfig builds a CalabresellaConfig from the nested web config, applying bounds checking.
func (c *CalabresellaWebConfig) ToConfig() domain.CalabresellaConfig {
	cfg := domain.DefaultCalabresellaConfig()
	cfg.CpuDifficulty = domain.CalabresellaCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CalabresellaCpuDifficultyEasy), int(domain.CalabresellaCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetPoints, c.TargetPoints, 1, 1000000)
	return cfg
}

// ToConfig builds a CalabresellaConfig from the web input.
func (p CalabresellaWebInput) ToConfig() domain.CalabresellaConfig {
	return configOrDefault(p.Config, (*CalabresellaWebConfig).ToConfig, domain.DefaultCalabresellaConfig())
}

// CalabresellaWebController カラブレセッラのWebコントローラークラス
type CalabresellaWebController = GameWebController[usecase.CalabresellaInteractorIF, CalabresellaWebInput, *CalabresellaWebOutput]

// NewCalabresellaWebController and NewCalabresellaWebControllerWithProvider are
// the standard and provider-backed constructors for CalabresellaWebController.
var NewCalabresellaWebController, NewCalabresellaWebControllerWithProvider = webControllerPair[usecase.CalabresellaInteractorIF, CalabresellaWebInput, *CalabresellaWebOutput](
	newCalabresellaDefaultOutput, calabresellaDispatch,
)

func newCalabresellaDefaultOutput(msg string) *CalabresellaWebOutput {
	return &CalabresellaWebOutput{
		Players:         make([]*CalabresellaWebOutputPlayer, 0),
		CurrentTrick:    make([]*CalabresellaWebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		SoloistIdx:      -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func calabresellaDispatch(bc *baseController, w http.ResponseWriter, di usecase.CalabresellaInteractorIF, param CalabresellaWebInput, newDefault func(string) *CalabresellaWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, di.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Bid(domain.CalabresellaBid(*param.Bid)))
	case "d", "discard":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, di.Discard(*param.CardIndex))
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
