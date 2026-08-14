//go:build !js || !wasm || casino

package domain

// trackPreFlopStats プリフロップのHUDスタッツを追跡
func (h *Holdem) trackPreFlopStats(playerIdx, action int) {
	if h.phase != HoldemPhasePreFlop {
		return
	}

	isVPIPAction := false
	isPFRAction := false

	switch action {
	case HoldemActionCall:
		isVPIPAction = true
	case HoldemActionBet, HoldemActionRaise, HoldemActionAllIn:
		isVPIPAction = true
		isPFRAction = true
	}

	if isVPIPAction && !h.vpipTracked[playerIdx] {
		h.players[playerIdx].IncrementVPIP()
		h.vpipTracked[playerIdx] = true
	}

	if isPFRAction && !h.pfrTracked[playerIdx] {
		h.players[playerIdx].IncrementPFR()
		h.pfrTracked[playerIdx] = true
	}

	// 3Bet追跡: raiseCount >= 1 (既にレイズがある) かつ未追跡
	if h.raiseCount >= 1 && !h.threeBetTracked[playerIdx] {
		h.players[playerIdx].IncrementThreeBetOpportunity()
		if action == HoldemActionRaise || action == HoldemActionAllIn {
			h.players[playerIdx].IncrementThreeBet()
		}
		h.threeBetTracked[playerIdx] = true
	}
}

// trackPostFlopStats ポストフロップのAFスタッツを追跡
func (h *Holdem) trackPostFlopStats(playerIdx, action int) {
	if h.phase < HoldemPhaseFlop || h.phase > HoldemPhaseRiver {
		return
	}

	switch action {
	case HoldemActionBet, HoldemActionRaise, HoldemActionAllIn:
		h.players[playerIdx].IncrementPostFlopBetRaise()
	case HoldemActionCall:
		h.players[playerIdx].IncrementPostFlopCall()
	}
}

// resolveLastPlayer 全員フォールドで最後のプレイヤーが勝利
func (h *Holdem) resolveLastPlayer() {
	for i, p := range h.players {
		if !p.GetFolded() {
			p.AddChips(h.pot)
			h.roundResults = []HoldemResult{{
				PlayerIdx: i,
				WonAmount: h.pot,
			}}
			h.pot = 0
			break
		}
	}
	h.phase = HoldemPhaseEnd
	h.gameEndFlag = true
	h.dealerIdx = (h.dealerIdx + 1) % len(h.players)
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (h *Holdem) resolveShowdown() {
	// ハンド評価
	for _, p := range h.players {
		if !p.GetFolded() {
			p.EvalBestHand(h.communityCards)
		}
	}

	// サイドポット計算・配分
	bp := toBettingPlayers(h.players)
	h.sidePots = CalculateSidePots(bp, h.pot, h.startingChips)
	wonAmounts := DistributePots(bp, h.sidePots)

	// 結果を構築
	h.roundResults = make([]HoldemResult, 0)
	humanLost := false
	for i, p := range h.players {
		if p.GetFolded() {
			continue
		}
		result := HoldemResult{
			PlayerIdx: i,
			HandRank:  p.GetHandRank(),
			HandName:  h.getHandName(p.GetHandRank()),
			BestHand:  p.GetBestHand(),
			Kickers:   ExtractKickers(p.GetBestHand(), p.GetHandRank()),
			WonAmount: wonAmounts[i],
		}
		h.roundResults = append(h.roundResults, result)
		if p.GetIsHuman() && wonAmounts[i] == 0 {
			humanLost = true
		}
	}

	// 人間が負けた場合、マック選択のためSHOWDOWNフェーズに留まる
	if humanLost {
		return
	}

	h.finalizeShowdown()
}

// finalizeShowdown ショーダウンを完了し、END フェーズに遷移する
func (h *Holdem) finalizeShowdown() {
	// **配り終えたポットは 0 にする。** resolveShowdown は DistributePots で
	// チップを配るが pot を戻さないので、END でも配り終えた額が残り続けていた
	// ── 全員降りて終わる resolveLastPlayer は 0 にしており、同じ「ハンドが
	// 終わった」状態なのに片方だけ残るのは読み手を誤らせる (実測 200 ハンド中
	// 176 で残っていた)。次の Reset が作り直すので、消して失うものは無い。
	h.pot = 0
	h.phase = HoldemPhaseEnd
	h.gameEndFlag = true
	h.dealerIdx = (h.dealerIdx + 1) % len(h.players)
}

// Muck 人間プレイヤーがハンドをマックする (公開せずに伏せる)
func (h *Holdem) Muck() error {
	if h.phase != HoldemPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Muck is not available now.")
	}
	for i := range h.roundResults {
		if h.players[h.roundResults[i].PlayerIdx].GetIsHuman() {
			h.roundResults[i].Mucked = true
			break
		}
	}
	h.finalizeShowdown()
	return nil
}

// ShowHand 人間プレイヤーがハンドを公開する
func (h *Holdem) ShowHand() error {
	if h.phase != HoldemPhaseShowdown {
		return NewDomainError(ErrWrongPhase, "Show hand is not available now.")
	}
	h.finalizeShowdown()
	return nil
}

// IsMuckAvailable 人間プレイヤーがマック可能かどうか
func (h *Holdem) IsMuckAvailable() bool {
	if h.phase != HoldemPhaseShowdown {
		return false
	}
	for _, r := range h.roundResults {
		if h.players[r.PlayerIdx].GetIsHuman() && r.WonAmount == 0 {
			return true
		}
	}
	return false
}

// getHandName ハンドランクから名前を返す
func (h *Holdem) getHandName(rank int) string {
	return pokerHandName(rank)
}
