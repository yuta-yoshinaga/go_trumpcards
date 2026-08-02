import type { Card, HandAndFootResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { HandAndFootPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Hand and Foot, or null when no
 * suggestion is available.
 *
 * The response carries no melds for the human, so this uses the same shallow
 * connects-with-something test as the other rummies rather than proving one.
 *
 * Two things are specific here. A **frozen** discard pile can only be taken
 * with a natural pair from hand, so the usual "it connects, take it" advice is
 * wrong while the freeze is on. And a player still on their first hand has a
 * **foot** waiting: going out is not the goal until they have picked it up.
 */
export function getHandAndFootHint(state: HandAndFootResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0 || state.currentPlayerIdx !== human.id) return null;

  if (state.phase === HandAndFootPhase.DRAW) {
    const top = state.discardTop;
    if (!top) {
      return { targetAction: 'drawStock', reason: 'frontendHint.handandfootDrawStock', confidence: 'moderate' };
    }
    // **凍結中は自然のペアが要る。**繋がっているだけでは取れない。
    if (state.isFrozen) {
      const pairs = human.cards.filter((c) => c.value === top.value).length;
      return pairs >= 2
        ? { targetAction: 'takeDiscard', reason: 'frontendHint.handandfootTakeFrozen', confidence: 'moderate' }
        : { targetAction: 'drawStock', reason: 'frontendHint.handandfootFrozenBlocked', confidence: 'moderate' };
    }
    return connects(top, human.cards)
      ? { targetAction: 'takeDiscard', reason: 'frontendHint.handandfootTakeDiscard', confidence: 'moderate' }
      : { targetAction: 'drawStock', reason: 'frontendHint.handandfootDrawStock', confidence: 'moderate' };
  }

  if (state.phase !== HandAndFootPhase.DISCARD) return null;

  // **まだフットを持っていない。**上がるのではなく手を空けるのが目的。
  if (!human.inFoot && human.footCount > 0) {
    return { targetAction: 'discard', reason: 'frontendHint.handandfootReachFoot', confidence: 'moderate' };
  }

  const idx = heaviestLoose(human.cards);
  return { targetAction: `card-${idx}`, reason: 'frontendHint.handandfootDiscardHeavy', confidence: 'moderate' };
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
