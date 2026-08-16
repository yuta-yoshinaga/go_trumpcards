package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HeartsWebInput ハーツWebインプット
type HeartsWebInput struct {
	BaseWebInput
	CardIndices []int            `json:"cardIndices,omitempty"`
	CardIndex   *int             `json:"cardIndex,omitempty"`
	Config      *HeartsWebConfig `json:"config,omitempty"`
}

// HeartsWebConfig ハーツWeb設定
type HeartsWebConfig struct {
	CpuDifficulty *int  `json:"cpuDifficulty,omitempty"`
	PointLimit    *int  `json:"pointLimit,omitempty"`
	OmnibusJD     *bool `json:"omnibusJD,omitempty"`
}

// HeartsWebOutputPlayer ハーツWebアウトプットプレイヤー
type HeartsWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
	// PenaltyCards はこのプレイヤーがこれまでに獲得したペナルティカード
	// (全ハート + Q♠。J♦は含まない) の一覧。フロントエンドが「♥×n / Q♠」の
	// 内訳表示に用いる。トリックテイキングでは公開情報のため秘匿ゲートは不要。
	PenaltyCards []*WebOutputCard `json:"penaltyCards"`
	// TookOmnibusJD はオムニバス規則が有効で、かつこのプレイヤーが J♦ を獲得
	// 済みのとき真。J♦ は -10 点というスイングを持つがボーナスなので
	// PenaltyCards には入らず、入れると「ペナルティ」の意味が変わってしまう。
	// 規則が無効なら常に偽 (そのとき J♦ はただの札)。
	TookOmnibusJD bool `json:"tookOmnibusJD"`
}

// HeartsWebOutput ハーツWebアウトプット
type HeartsWebOutput struct {
	Players          []*HeartsWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	HeartsBroken     bool                     `json:"heartsBroken"`
	PassDirection    int                      `json:"passDirection"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	Hint             *WebOutputCardHint       `json:"hint,omitempty"`
	WebOutputBase
	Config HeartsWebOutputConfig `json:"config"`
}

// HeartsWebOutputConfig ハーツ設定アウトプット
type HeartsWebOutputConfig struct {
	CpuDifficulty int  `json:"cpuDifficulty"`
	PointLimit    int  `json:"pointLimit"`
	OmnibusJD     bool `json:"omnibusJD"`
}

// ToConfig builds a HeartsConfig from the nested web config, applying bounds checking.
func (c *HeartsWebConfig) ToConfig() domain.HeartsConfig {
	cfg := domain.DefaultHeartsConfig()
	cfg.CpuDifficulty = domain.HeartsCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.HeartsCpuDifficultyEasy), int(domain.HeartsCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	webutil.ApplyBool(&cfg.OmnibusJD, c.OmnibusJD)
	return cfg
}

// ToConfig builds a HeartsConfig from the web input.
func (p HeartsWebInput) ToConfig() domain.HeartsConfig {
	return configOrDefault(p.Config, (*HeartsWebConfig).ToConfig, domain.DefaultHeartsConfig())
}

// HeartsWebController ハーツWebコントローラークラス
type HeartsWebController = GameWebController[usecase.HeartsInteractorIF, HeartsWebInput, *HeartsWebOutput]

// NewHeartsWebController and NewHeartsWebControllerWithProvider are
// the standard and provider-backed constructors for HeartsWebController.
var NewHeartsWebController, NewHeartsWebControllerWithProvider = webControllerPair[usecase.HeartsInteractorIF, HeartsWebInput, *HeartsWebOutput](
	newHeartsDefaultOutput, heartsDispatch,
)

func newHeartsDefaultOutput(msg string) *HeartsWebOutput {
	return &HeartsWebOutput{
		Players:       make([]*HeartsWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func heartsDispatch(bc *baseController, w http.ResponseWriter, hi usecase.HeartsInteractorIF, param HeartsWebInput, newDefault func(string) *HeartsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, hi.ResetWithConfig(param.ToConfig()))
	case "pass":
		if !requireParam(bc, w, newDefault, len(param.CardIndices) != 3, "param error: pass requires exactly 3 card indices.") {
			return true
		}
		bc.writePresenterResponse(w, hi.Pass(param.CardIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, hi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, hi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, hi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, hi.Hint, hi.ActionLog)
	}
	return true
}
