import { useRef } from 'react';
import {
  applyTriPeaksScore,
  initialTriPeaksScore,
  type TriPeaksScoreState,
  type TriPeaksSnapshot,
} from '../utils/tripeaksScore';

/**
 * Tracks the running TriPeaks score across board transitions (issue #3087).
 *
 * Watches `moveCount`, `stockCount`, and the number of cleared peaks from the
 * server response and folds each change into the score via `applyTriPeaksScore`.
 * The score is derived synchronously during render (guarded by an input key so
 * each transition is applied exactly once, StrictMode-safe) so it is always
 * current in the same render as the board state — important because the move that
 * clears the last peak also ends the game, and the end-of-game recorder must see
 * that move's points. Undefined inputs (before the first fetch) reset to zero.
 */
export function useTriPeaksScore(
  moveCount: number | undefined,
  stockCount: number | undefined,
  peaksCleared: number,
): TriPeaksScoreState {
  const scoreRef = useRef<TriPeaksScoreState>(initialTriPeaksScore);
  const prevSnapshot = useRef<TriPeaksSnapshot | null>(null);
  const lastKey = useRef<string | null>(null);

  const key = moveCount === undefined || stockCount === undefined ? null : `${moveCount}|${stockCount}|${peaksCleared}`;

  if (key !== lastKey.current) {
    lastKey.current = key;
    if (key === null) {
      scoreRef.current = initialTriPeaksScore;
      prevSnapshot.current = null;
    } else {
      const cur: TriPeaksSnapshot = {
        moveCount: moveCount as number,
        stockCount: stockCount as number,
        peaksCleared,
      };
      scoreRef.current = applyTriPeaksScore(scoreRef.current, prevSnapshot.current, cur);
      prevSnapshot.current = cur;
    }
  }

  return scoreRef.current;
}
