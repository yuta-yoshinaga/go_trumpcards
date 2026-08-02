import type { Card, ContractRummyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** フェーズ番号 (sync: internal/domain/ContractRummy.go)。 */
const PHASE_DRAW = 0;
const PHASE_PLAY = 1;

/**
 * Returns a frontend {@link HintResult} for Contract Rummy, or null when no
 * suggestion is available.
 *
 * A player's `melds` is what they have laid down, not what their hand could
 * make, so this uses the same shallow connects-with-something test as the other
 * rummies rather than proving a meld.
 *
 * What is specific here is the contract. Until it is met the whole contract has
 * to go down in one turn, so the heavy cards are material and throwing them is
 * wrong — the same inversion Kalooki has, but driven by a per-round slot list
 * rather than a point threshold, which is why the hint names the slots left.
 */
export function getContractRummyHint(state: ContractRummyResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === PHASE_DRAW) {
    const top = state.discardTop;
    return top && connects(top, human.cards)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.contractrummyTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.contractrummyDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== PHASE_PLAY) return null;

  // **契約を満たすまでは崩さない。**一度に全部そろえる必要があるので、
  // 重い札こそ材料になる。
  if (!human.contractMet && state.contractSlots.length > 0) {
    return { targetAction: 'meld', reason: 'frontendHint.contractrummyMeetContract', confidence: 'moderate' };
  }

  const idx = heaviestLoose(human.cards);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.contractrummyDiscardHeavy', confidence: 'moderate' };
}

/** 同じランクがあるか、同じスートで隣のランクがあるか。メルドの証明ではない。 */
function connects(c: Card, hand: Card[]): boolean {
  return hand.some((o) => o.value === c.value || (o.design === c.design && Math.abs(o.value - c.value) === 1));
}

/** 繋がっていない札のうち一番重いものの位置。全部繋がっていれば一番重い札。 */
function heaviestLoose(hand: Card[]): number {
  const loose = hand
    .map((_, i) => i)
    .filter(
      (i) =>
        !connects(
          hand[i],
          hand.filter((_, j) => j !== i),
        ),
    );
  const pool = loose.length > 0 ? loose : hand.map((_, i) => i);
  let best = pool[0];
  for (const i of pool) {
    if (hand[i].value > hand[best].value) best = i;
  }
  return best;
}
