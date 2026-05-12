import type { Card, MississippiStudResponse } from '../../types/card';
import { isMaskedCard } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MississippiStudPhase } from '../../types/phases';

const TENS_VALUE = 10;
const ACE_VALUE = 1;

/**
 * Returns a Mississippi Stud betting hint for the 3rd / 4th / 5th street
 * decision phases. The recommendation is a "raise 3x" when the made hand
 * already pays (pair of 6+, two pair, etc.) or a strong draw exists,
 * a "raise 1x" for marginal hands with reasonable draws, and "fold"
 * otherwise.
 *
 * `state.handRank` is not populated until resolution, so hand quality is
 * evaluated directly from the revealed cards (hole + revealed community).
 */
export function getMississippiStudHint(state: MississippiStudResponse): HintResult | null {
  if (
    state.phase !== MississippiStudPhase.THIRD_STREET &&
    state.phase !== MississippiStudPhase.FOURTH_STREET &&
    state.phase !== MississippiStudPhase.FIFTH_STREET
  ) {
    return null;
  }
  if (!state.playerHand || state.playerHand.length === 0) return null;

  const revealed = state.communityCards.filter((c): c is Card => !isMaskedCard(c));
  const cards: Card[] = [...state.playerHand, ...revealed];

  if (hasMadeHandAtLeastMidPair(cards)) {
    return { targetAction: 'play3x', reason: 'hint.raiseBig', confidence: 'strong' };
  }

  if (hasReasonableDraw(cards, state.phase)) {
    return { targetAction: 'play1x', reason: 'hint.raiseSmall', confidence: 'moderate' };
  }

  return { targetAction: 'fold', reason: 'hint.fold', confidence: 'moderate' };
}

function valueCounts(cards: Card[]): Map<number, number> {
  const counts = new Map<number, number>();
  for (const c of cards) {
    counts.set(c.value, (counts.get(c.value) ?? 0) + 1);
  }
  return counts;
}

function hasMadeHandAtLeastMidPair(cards: Card[]): boolean {
  const counts = valueCounts(cards);
  let maxCount = 0;
  let pairCount = 0;
  let qualifyingPair = false;
  for (const [value, count] of counts) {
    if (count > maxCount) maxCount = count;
    if (count >= 2) {
      pairCount++;
      if (value === ACE_VALUE || value >= 6) qualifyingPair = true;
    }
  }
  return maxCount >= 3 || pairCount >= 2 || qualifyingPair;
}

function hasReasonableDraw(cards: Card[], phase: number): boolean {
  const totalRevealed = cards.length;
  const slotsLeft = 5 - totalRevealed;

  if (hasFlushDraw(cards, slotsLeft)) return true;
  if (hasOpenStraightDraw(cards, slotsLeft)) return true;

  if (phase === MississippiStudPhase.THIRD_STREET) {
    const highCards = cards.filter((c) => c.value === ACE_VALUE || c.value >= TENS_VALUE);
    if (highCards.length >= 2) return true;
  }

  return false;
}

function hasFlushDraw(cards: Card[], slotsLeft: number): boolean {
  const bySuit = new Map<string, number>();
  for (const c of cards) bySuit.set(c.design, (bySuit.get(c.design) ?? 0) + 1);
  for (const count of bySuit.values()) {
    if (count >= 3 && count + slotsLeft >= 5) return true;
  }
  return false;
}

function hasOpenStraightDraw(cards: Card[], slotsLeft: number): boolean {
  const set = new Set<number>();
  for (const c of cards) {
    set.add(c.value);
    if (c.value === ACE_VALUE) set.add(14);
  }
  const values = [...set].sort((a, b) => a - b);
  let bestRun = 1;
  for (let i = 1; i < values.length; i++) {
    if (values[i] === values[i - 1] + 1) {
      bestRun++;
    } else {
      bestRun = 1;
    }
    if (bestRun + slotsLeft >= 5) return true;
  }
  return false;
}
