package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EgyptianRatscrewWebInput エジプシャン・ラットスクリュー Web 入力
type EgyptianRatscrewWebInput struct {
	BaseWebInput
	Config *EgyptianRatscrewWebConfig `json:"config,omitempty"`
}

// EgyptianRatscrewWebConfig エジプシャン・ラットスクリューの設定リクエスト
type EgyptianRatscrewWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// EgyptianRatscrewWebPlayer プレイヤー出力
type EgyptianRatscrewWebPlayer struct {
	Name      string `json:"name"`
	IsHuman   bool   `json:"isHuman"`
	StockSize int    `json:"stockSize"`
}

// EgyptianRatscrewWebFaceChances は絵札ごとのチャンス回数。
type EgyptianRatscrewWebFaceChances struct {
	Jack  int `json:"jack"`
	Queen int `json:"queen"`
	King  int `json:"king"`
	Ace   int `json:"ace"`
}

// NewEgyptianRatscrewWebFaceChances はドメインの FaceCardChances から回数を組む。
//
// 既定の応答とプレゼンタの両方が同じものを返す必要があるので、literal を 2 箇所に
// 書かない ── 片方だけ直すと、盤面のある応答と無い応答で規則の説明が食い違う。
func NewEgyptianRatscrewWebFaceChances() *EgyptianRatscrewWebFaceChances {
	return &EgyptianRatscrewWebFaceChances{
		Jack:  domain.FaceCardChances(domain.EgyptianRatscrewJackValue),
		Queen: domain.FaceCardChances(domain.EgyptianRatscrewQueenValue),
		King:  domain.FaceCardChances(domain.EgyptianRatscrewKingValue),
		Ace:   domain.FaceCardChances(domain.EgyptianRatscrewAceValue),
	}
}

// EgyptianRatscrewWebOutput エジプシャン・ラットスクリュー Web 出力
type EgyptianRatscrewWebOutput struct {
	Phase           int                          `json:"phase"`
	GameEndFlag     bool                         `json:"gameEndFlag"`
	WinnerIdx       int                          `json:"winnerIdx"`
	CurrentTurnIdx  int                          `json:"currentTurnIdx"`
	IsHumanTurn     bool                         `json:"isHumanTurn"`
	IsTopFaceCard   bool                         `json:"isTopFaceCard"`
	IsSlappable     bool                         `json:"isSlappable"`
	CenterPileSize  int                          `json:"centerPileSize"`
	TopCard         *WebOutputCard               `json:"topCard,omitempty"`
	Players         []*EgyptianRatscrewWebPlayer `json:"players"`
	CpuDifficulty   int                          `json:"cpuDifficulty"`
	ChanceRemaining int                          `json:"chanceRemaining"`
	// FaceChances は絵札ごとに相手へ課すチャンスの回数 (#5580)。規則の説明を
	// 画面に書くのに要る。数字を訳文に焼き込むと、回数を変えたとき説明だけが嘘になる。
	FaceChances        *EgyptianRatscrewWebFaceChances `json:"faceChances"`
	ChanceFromIdx      int                             `json:"chanceFromIdx"`
	PendingKind        int                             `json:"pendingKind"`
	PendingDeadlineMs  int64                           `json:"pendingDeadlineMs"`
	LastEventKind      int                             `json:"lastEventKind"`
	LastEventPlayerIdx int                             `json:"lastEventPlayerIdx"`
	LastSlapReason     int                             `json:"lastSlapReason"`
	WebOutputBase
}

// EgyptianRatscrewWebController エジプシャン・ラットスクリュー Web コントローラー
type EgyptianRatscrewWebController = GameWebController[usecase.EgyptianRatscrewInteractorIF, EgyptianRatscrewWebInput, *EgyptianRatscrewWebOutput]

// NewEgyptianRatscrewWebController and NewEgyptianRatscrewWebControllerWithProvider are
// the standard and provider-backed constructors for EgyptianRatscrewWebController.
var NewEgyptianRatscrewWebController, NewEgyptianRatscrewWebControllerWithProvider = webControllerPair[usecase.EgyptianRatscrewInteractorIF, EgyptianRatscrewWebInput, *EgyptianRatscrewWebOutput](
	newEgyptianRatscrewDefaultOutput, egyptianRatscrewDispatch,
)

func newEgyptianRatscrewDefaultOutput(msg string) *EgyptianRatscrewWebOutput {
	return &EgyptianRatscrewWebOutput{
		WinnerIdx:     -1,
		ChanceFromIdx: -1,
		CpuDifficulty: int(domain.EgyptianRatscrewCpuNormal),
		Players:       make([]*EgyptianRatscrewWebPlayer, 0),
		// 回数は盤面が無くても規則なので、既定の応答にも乗せる。
		FaceChances:   NewEgyptianRatscrewWebFaceChances(),
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func egyptianRatscrewDispatch(bc *baseController, w http.ResponseWriter, ei usecase.EgyptianRatscrewInteractorIF, param EgyptianRatscrewWebInput, _ func(string) *EgyptianRatscrewWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, ei.ResetWithConfig(egyptianRatscrewConfigFromInput(ei.GetConfig(), param.Config)))
		} else {
			bc.writePresenterResponse(w, ei.Reset())
		}
		return true
	case "s", "step":
		bc.writePresenterResponse(w, ei.Step())
		return true
	case "j", "slap":
		// Web エンドポイントは人間 (idx=0) の slap 専用。
		// クライアント値は信用せずサーバ側で固定する (CPU を強制 slap させないため)。
		bc.writePresenterResponse(w, ei.Slap(0))
		return true
	case "tick":
		bc.writePresenterResponse(w, ei.Tick())
		return true
	case "log", "l":
		bc.writePresenterResponse(w, ei.ActionLog())
		return true
	}
	return false
}

// egyptianRatscrewConfigFromInput merges the partial Web config request into the current
// config so missing fields default to existing values rather than zero.
func egyptianRatscrewConfigFromInput(current domain.EgyptianRatscrewConfig, in *EgyptianRatscrewWebConfig) domain.EgyptianRatscrewConfig {
	out := current
	if in.CpuDifficulty != nil {
		out.CpuDifficulty = domain.EgyptianRatscrewCpuDifficulty(*in.CpuDifficulty)
	}
	return out
}
