import type { Card, PokerSquaresResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PokerSquaresPhase } from '../../types/phases';

/** Board dimension (5x5). */
const SIZE = 5;

/**
 * Returns a Poker Squares placement hint for the currently drawn card, or null
 * when no card is available or the game is no longer in progress.
 *
 * Strategy: score every empty cell by the poker-hand potential it adds to the
 * affected row and column (same-value pair/triple, same-suit flush, adjacent
 * value straight), and recommend the cell with the highest combined score.
 * The `targetAction` is returned as `cell-<row>-<col>` so the page can match
 * it against `data-hint-action` on the target grid button.
 */
export function getPokersquaresHint(state: PokerSquaresResponse): HintResult | null {
  if (state.phase !== PokerSquaresPhase.PLAYING) return null;
  const current = state.currentCard;
  if (!current) return null;

  let best: { row: number; col: number; score: number } | null = null;
  for (let row = 0; row < SIZE; row++) {
    for (let col = 0; col < SIZE; col++) {
      if (state.board[row]?.[col]?.card) continue;
      const score = scorePlacement(state, current, row, col);
      if (!best || score > best.score) {
        best = { row, col, score };
      }
    }
  }

  if (!best) return null;

  return {
    targetAction: `cell-${best.row}-${best.col}`,
    reason: best.score > 0 ? 'hint.placeSynergy' : 'hint.placeAny',
    confidence: best.score >= 6 ? 'strong' : 'moderate',
  };
}

/** Score a candidate placement by evaluating potential in its row and column. */
function scorePlacement(state: PokerSquaresResponse, card: Card, row: number, col: number): number {
  const rowCards = collectLine(state, 'row', row).filter((c) => c !== null) as Card[];
  const colCards = collectLine(state, 'col', col).filter((c) => c !== null) as Card[];
  return scoreLine(card, rowCards) + scoreLine(card, colCards);
}

/** Return all cards (or nulls) in the requested row or column. */
function collectLine(state: PokerSquaresResponse, axis: 'row' | 'col', index: number): (Card | null)[] {
  const line: (Card | null)[] = [];
  for (let i = 0; i < SIZE; i++) {
    const cell = axis === 'row' ? state.board[index]?.[i] : state.board[i]?.[index];
    line.push(cell?.card ?? null);
  }
  return line;
}

/** Score the synergy of placing `card` into a line already containing `existing` cards. */
function scoreLine(card: Card, existing: Card[]): number {
  if (existing.length === 0) return 0;

  const sameValue = existing.filter((c) => c.value === card.value).length;
  const sameSuit = existing.filter((c) => c.design === card.design).length;
  const adjacent = existing.filter((c) => Math.abs(c.value - card.value) === 1).length;

  // Pair/triple potential is the strongest signal; flushes next; straights weakest.
  return sameValue * 4 + sameSuit * 2 + adjacent;
}
