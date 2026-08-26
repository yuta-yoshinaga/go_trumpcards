import type { Card, DramahaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { DramahaPhase } from '../../types/phases';
import { DRAMAHA_HOLE_CARDS } from '../dramahaBestFive';
import { evaluateFiveCardHand, PokerHand } from '../pokerSquaresUtils';
import { getHoldemBaseHint } from './holdemBaseHint';

/** Rank at or above which a half of the pot is worth raising for on its own. */
const STRONG_RANK = PokerHand.ThreeOfAKind;

/** Hole cards kept when nothing pairs — the two highest, as the CPU does. */
const KEEP_WHEN_NOTHING_PAIRS = 2;

/**
 * Returns a frontend HintResult for Dramaha, or null when there is no
 * suggestion to make.
 *
 * Two kinds of advice, because Dramaha asks two different questions:
 *
 *   - in the draw round, *which cards to exchange* (or that standing pat is
 *     right). No betting decision exists there, so the Hold'em base hint —
 *     which reasons about pot odds — has nothing to say;
 *   - while betting, whether the two halves of the split justify putting chips
 *     in. Half the pot is decided by the five hole cards alone, so a hand the
 *     board has missed entirely can still be worth a call.
 */
export function getDramahaHint(state: DramahaResponse): HintResult | null {
  if (
    state.phase === DramahaPhase.INIT ||
    state.phase === DramahaPhase.SHOWDOWN ||
    state.phase === DramahaPhase.END ||
    state.phase === DramahaPhase.REBUY
  ) {
    return null;
  }

  const human = state.players?.find((p) => p.isHuman);
  if (!human || human.folded) return null;

  // Checked before the all-in guard: an all-in seat has no betting decision
  // left but still draws, and the draw decides half the pot it is contesting.
  if (state.phase === DramahaPhase.DRAW) return drawHint(human.cards ?? []);
  if (human.allIn) return null;

  const base = getHoldemBaseHint(state);
  // The draw half is the five hole cards exactly as dealt — the board is not
  // part of this reading, which is why it can be strong on a board that missed.
  const drawRank = evaluateFiveCardHand(human.cards ?? []);
  if (drawRank == null || drawRank < STRONG_RANK) return base;

  // The draw half alone is strong. If the Omaha half agrees, both halves are
  // live and the hand is playing for the scoop; if it does not, half the pot
  // is still in reach, so folding gives away a hand that is already made.
  if (base?.targetAction === 'raise') {
    return { targetAction: 'raise', reason: 'hint.scoopChance', confidence: 'strong' };
  }
  if (!base || base.targetAction === 'fold') {
    return { targetAction: 'call', reason: 'hint.drawHalfOnly', confidence: 'moderate' };
  }
  return base;
}

/**
 * Advice for the draw round: the positions to exchange, or standing pat.
 *
 * Mirrors the backend's `dramahaCPUDiscards` so the tooltip and the CPUs read
 * the hand the same way — keep anything paired or better, and when nothing
 * pairs keep only the two highest cards.
 */
function drawHint(hole: readonly Card[]): HintResult | null {
  if (hole.length !== DRAMAHA_HOLE_CARDS) return null;
  const discards = dramahaDiscardSuggestion(hole);
  if (discards.length === 0) {
    return { targetAction: 'standpat', reason: 'hint.standPat', confidence: 'strong' };
  }
  return {
    targetAction: 'draw',
    reason: 'hint.exchange',
    reasonParams: { count: discards.length },
    targetIndices: discards,
    confidence: 'moderate',
  };
}

/**
 * The hole-card positions worth exchanging, ascending.
 *
 * Empty means stand pat. Exported because the page highlights the same
 * positions the hint names.
 */
export function dramahaDiscardSuggestion(hole: readonly Card[]): number[] {
  if (hole.length !== DRAMAHA_HOLE_CARDS) return [];

  const freq = new Map<number, number>();
  for (const card of hole) freq.set(card.value, (freq.get(card.value) ?? 0) + 1);

  const discards: number[] = [];
  hole.forEach((card, idx) => {
    if ((freq.get(card.value) ?? 0) < 2) discards.push(idx);
  });
  if (discards.length < DRAMAHA_HOLE_CARDS) return discards;

  // Nothing pairs: keep the two highest and draw three, rather than throwing
  // the whole hand away.
  const byRankDesc = hole.map((_, i) => i).sort((a, b) => aceHigh(hole[b]) - aceHigh(hole[a]));
  return byRankDesc.slice(KEEP_WHEN_NOTHING_PAIRS).sort((a, b) => a - b);
}

/** Card height with the ace played high, matching the backend's ranking. */
function aceHigh(card: Card): number {
  return card.value === 1 ? 14 : card.value;
}
