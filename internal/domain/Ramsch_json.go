//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// Ramsch の KV 永続化。
//
// **これが無いと盤が毎リクエスト消える。** `Ramsch` のフィールドは全部
// 非公開なので、素の `json.Marshal` は `{}` を返す ── エラーにもならず、
// Worker では毎回まっさらな盤から始まる。手元の CLI は 1 プロセスなので
// 一切気付けない。実際、クローン元の Skat は今もこの状態で出荷されている
// (2 バイト `{}` になることを実測)。
//
// **新しいフィールドを足したら必ずここにも足すこと。** 追加を忘れても
// コンパイルは通り、テストも（状態を覗くものは）通る。

// ramschJSON is the JSON wire format for Ramsch.
type ramschJSON struct {
	// TrumpCards は配り切ったあと引かれないので、**このゲームでは落としても
	// 観測できない**（32 枚すべてが Reset で手札と伏せ札に行く）。正しさの
	// ために載せてあるだけで、テストで守られてはいない。
	TrumpCards *TrumpCards          `json:"tc"`
	Players    []*RamschPlayer      `json:"pl"`
	Config     RamschConfig         `json:"cf"`
	Round      ramschRoundStateJSON `json:"rd"`
}

// ramschRoundStateJSON is the JSON wire format for the round state.
type ramschRoundStateJSON struct {
	Phase            RamschPhase       `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"cp"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	LeadPlayerIdx    int               `json:"lp"`
	DealerIdx        int               `json:"di"`
	ForehandIdx      int               `json:"fh"`
	MiddlehandIdx    int               `json:"mh"`
	RearhandIdx      int               `json:"rh"`
	Skat             []*Card           `json:"sk"`
	LoserIdx         int               `json:"li"`
	Durchmarsch      bool              `json:"dm"`
	TrickResolved    bool              `json:"tr"`
	DurchmarschIdx   int               `json:"dx"`
	GameEndFlag      bool              `json:"ge"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *Ramsch) MarshalJSON() ([]byte, error) {
	return json.Marshal(ramschJSON{
		TrumpCards: s.trumpCards,
		Players:    s.players,
		Config:     s.config,
		Round: ramschRoundStateJSON{
			Phase:            s.round.phase,
			RoundNumber:      s.round.roundNumber,
			TrickNumber:      s.round.trickNumber,
			CurrentPlayerIdx: s.round.currentPlayerIdx,
			CurrentTrick:     s.round.currentTrick,
			LeadPlayerIdx:    s.round.leadPlayerIdx,
			DealerIdx:        s.round.dealerIdx,
			ForehandIdx:      s.round.forehandIdx,
			MiddlehandIdx:    s.round.middlehandIdx,
			RearhandIdx:      s.round.rearhandIdx,
			Skat:             s.round.skat,
			LoserIdx:         s.round.loserIdx,
			Durchmarsch:      s.round.durchmarsch,
			TrickResolved:    s.round.trickResolved,
			DurchmarschIdx:   s.round.durchmarschIdx,
			GameEndFlag:      s.round.gameEndFlag,
			ActionLog:        s.round.actionLog,
		},
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Ramsch) UnmarshalJSON(data []byte) error {
	var j ramschJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	s.trumpCards = j.TrumpCards
	s.players = j.Players
	s.config = j.Config
	s.round = ramschRoundState{
		phase:            j.Round.Phase,
		roundNumber:      j.Round.RoundNumber,
		trickNumber:      j.Round.TrickNumber,
		currentPlayerIdx: j.Round.CurrentPlayerIdx,
		currentTrick:     j.Round.CurrentTrick,
		leadPlayerIdx:    j.Round.LeadPlayerIdx,
		dealerIdx:        j.Round.DealerIdx,
		forehandIdx:      j.Round.ForehandIdx,
		middlehandIdx:    j.Round.MiddlehandIdx,
		rearhandIdx:      j.Round.RearhandIdx,
		skat:             j.Round.Skat,
		loserIdx:         j.Round.LoserIdx,
		durchmarsch:      j.Round.Durchmarsch,
		trickResolved:    j.Round.TrickResolved,
		durchmarschIdx:   j.Round.DurchmarschIdx,
		gameEndFlag:      j.Round.GameEndFlag,
	}
	s.round.actionLog = j.Round.ActionLog
	return nil
}
