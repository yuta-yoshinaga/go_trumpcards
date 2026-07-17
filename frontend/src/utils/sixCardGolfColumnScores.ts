import type { Card, SixCardGolfSlot } from '../types/card';

/** Number of card columns in a Six Card Golf grid (2 rows × 3 columns). */
export const SIX_CARD_GOLF_COLUMNS = 3;

/** Score of a single Six Card Golf card: K=0, A=1, J/Q=10, otherwise face value. */
export function sixCardGolfCardScore(card: Card | null): number {
  if (!card) return 0;
  switch (card.value) {
    case 13:
      return 0;
    case 1:
      return 1;
    case 11:
    case 12:
      return 10;
    default:
      return card.value;
  }
}

/** Score contribution of one column, mirroring the backend ScorePlayer rule. */
export interface SixCardGolfColumnScore {
  /** Combined points the column contributes (a matched pair contributes 0). */
  score: number;
  /** True when both cards are face up and share the same value (cancels to 0). */
  isPair: boolean;
  /** True when at least one card in the column is still face down (score not yet certain). */
  hasHidden: boolean;
}

/**
 * Computes the per-column score breakdown for a Six Card Golf grid.
 *
 * Columns are the vertical pairs (top index `c`, bottom index `c + 3`). A column
 * whose two face-up cards share a value cancels to 0; otherwise each face-up
 * card adds its {@link sixCardGolfCardScore}. Face-down cards score nothing.
 */
export function sixCardGolfColumnScores(grid: SixCardGolfSlot[]): SixCardGolfColumnScore[] {
  const result: SixCardGolfColumnScore[] = [];
  for (let col = 0; col < SIX_CARD_GOLF_COLUMNS; col++) {
    const top = grid[col];
    const bot = grid[col + SIX_CARD_GOLF_COLUMNS];
    // A slot with no face-up card (face down, or no card yet) keeps the column uncertain.
    const hasHidden = !(top?.faceUp && top.card != null) || !(bot?.faceUp && bot.card != null);
    const isPair =
      !!top?.faceUp && !!bot?.faceUp && top.card != null && bot.card != null && top.card.value === bot.card.value;
    if (isPair) {
      result.push({ score: 0, isPair: true, hasHidden: false });
      continue;
    }
    const topScore = top?.faceUp ? sixCardGolfCardScore(top.card) : 0;
    const botScore = bot?.faceUp ? sixCardGolfCardScore(bot.card) : 0;
    result.push({ score: topScore + botScore, isPair: false, hasHidden });
  }
  return result;
}
