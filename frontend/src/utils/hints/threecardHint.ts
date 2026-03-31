import type { ThreeCardResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ThreeCardPhase } from '../../types/phases';

/** Hand rank threshold: pair or better means always play. */
const RANK_PAIR = 2;

/** Q-6-4 rule: play with Queen-high or better when hand rank is high card. */
const Q_6_4_THRESHOLD = [12, 6, 4] as const;

/**
 * Converts card value to a comparable rank (Ace=14, King=13, ..., 2=2).
 * Card values: 1=Ace, 2-10=pip, 11=Jack, 12=Queen, 13=King.
 */
function cardRank(value: number): number {
  return value === 1 ? 14 : value;
}

/**
 * Checks if a high-card hand meets or exceeds Q-6-4 using lexicographic comparison.
 * The Q-6-4 rule is the optimal play/fold threshold in Three Card Poker.
 */
function meetsQ64(ranks: number[]): boolean {
  const sorted = [...ranks].sort((a, b) => b - a);
  for (let i = 0; i < Q_6_4_THRESHOLD.length; i++) {
    const threshold = Q_6_4_THRESHOLD[i];
    const rank = sorted[i] ?? 0;
    if (rank > threshold) return true;
    if (rank < threshold) return false;
  }
  return true;
}

/**
 * Returns a hint for Three Card Poker during the action phase (play/fold decision).
 * Uses the Q-6-4 rule: play with Queen-6-4 or better, fold otherwise.
 * Hands with pair or better always recommend play with strong confidence.
 */
export function getThreeCardHint(state: ThreeCardResponse): HintResult | null {
  if (state.phase !== ThreeCardPhase.ACTION) return null;

  if (state.playerHandRank >= RANK_PAIR) {
    return {
      targetAction: 'play',
      reason: 'hintReason.strongHand',
      confidence: 'strong',
    };
  }

  const ranks = state.playerHand.map((c) => cardRank(c.value));
  if (meetsQ64(ranks)) {
    return {
      targetAction: 'play',
      reason: 'hintReason.queenHighPlay',
      confidence: 'moderate',
    };
  }

  return {
    targetAction: 'fold',
    reason: 'hintReason.weakHand',
    confidence: 'moderate',
  };
}
