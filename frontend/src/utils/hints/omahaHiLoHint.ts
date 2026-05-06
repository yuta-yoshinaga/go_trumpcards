import type { Card, OmahaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { HoldemPhase } from '../../types/phases';
import { getHoldemBaseHint } from './holdemBaseHint';

/** Maximum rank that counts toward an 8-or-better low hand. Aces play low (value 1). */
const LOW_RANK_MAX = 8;

/** Minimum unique low ranks the board needs at showdown for the low pot to qualify. */
const BOARD_LOW_REQUIRED = 3;

/** Minimum unique low ranks the player must hold to play exactly two low cards
 * for the low hand (Omaha rule: exactly 2 hole + 3 board). */
const HOLE_LOW_REQUIRED = 2;

/** Total community cards on the board at the river (used to count cards still to come). */
const TOTAL_COMMUNITY_CARDS = 5;

/** Hand rank threshold above which the high side alone justifies a raise. */
const STRONG_HAND_RANK = 3;

/** Win-probability margin above pot odds that signals a +EV raise. */
const EV_RAISE_MARGIN = 0.1;

/** Returns a frontend HintResult for Omaha Hi-Lo (8 or Better), or null
 * when no suggestion is available. Layers low-side strategy and scoop
 * detection on top of the Hold'em base hint. */
export function getOmahaHiLoHint(state: OmahaResponse): HintResult | null {
  if (
    state.phase === HoldemPhase.INIT ||
    state.phase === HoldemPhase.SHOWDOWN ||
    state.phase === HoldemPhase.END ||
    state.phase === HoldemPhase.REBUY
  ) {
    return null;
  }

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn) return null;

  const baseHint = getHoldemBaseHint(state);
  if (!state.isHiLo) return baseHint;

  const lowDraw = evaluateLowDraw(human.cards, state.communityCards, state.phase);
  const highStrong = isHighStrong(state, human.handRank);

  if (lowDraw === 'qualified' && !highStrong && (!baseHint || baseHint.targetAction === 'fold')) {
    return { targetAction: 'call', reason: 'hint.lowDraw', confidence: 'moderate' };
  }

  if (lowDraw !== 'none' && highStrong) {
    return { targetAction: 'raise', reason: 'hint.scoopChance', confidence: 'strong' };
  }

  return baseHint;
}

/** Returns true when the high side alone is strong enough to want a raise —
 * either positive equity vs pot odds, or hand rank ≥ three of a kind. */
function isHighStrong(state: OmahaResponse, handRank: number): boolean {
  if (state.equity?.winProbability != null && state.potOdds != null) {
    return state.equity.winProbability > state.potOdds + EV_RAISE_MARGIN;
  }
  return handRank >= STRONG_HAND_RANK;
}

/** Low-draw quality classifier:
 *  - 'qualified': board already has ≥3 unique low ranks, so the low pot will qualify
 *  - 'viable':    fewer board low ranks, but enough cards still to come that low can still complete
 *  - 'none':      hole cards or board make a qualifying low impossible
 */
function evaluateLowDraw(hole: Card[], community: Card[], phase: number): 'qualified' | 'viable' | 'none' {
  const holeLowRanks = uniqueLowRanks(hole);
  if (holeLowRanks < HOLE_LOW_REQUIRED) return 'none';

  const boardLowRanks = uniqueLowRanks(community);
  if (boardLowRanks >= BOARD_LOW_REQUIRED) return 'qualified';

  const cardsToCome = TOTAL_COMMUNITY_CARDS - communityCardCount(phase);
  if (boardLowRanks + cardsToCome < BOARD_LOW_REQUIRED) return 'none';

  return 'viable';
}

/** Counts unique ranks ≤ LOW_RANK_MAX in the given card set. Ace = 1 plays low. */
function uniqueLowRanks(cards: Card[]): number {
  const seen = new Set<number>();
  for (const card of cards) {
    if (card.value >= 1 && card.value <= LOW_RANK_MAX) {
      seen.add(card.value);
    }
  }
  return seen.size;
}

/** Number of community cards visible in each phase (matches HoldemPhase). */
function communityCardCount(phase: number): number {
  switch (phase) {
    case HoldemPhase.PRE_FLOP:
      return 0;
    case HoldemPhase.FLOP:
      return 3;
    case HoldemPhase.TURN:
      return 4;
    case HoldemPhase.RIVER:
      return 5;
    default:
      return 0;
  }
}
