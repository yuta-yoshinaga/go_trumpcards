import type { MonteCarloBoardCell } from '../types/card';

/**
 * Counts the removable pairs currently present on a Monte Carlo Solitaire board.
 *
 * A removable pair is two filled cells that are adjacent (8-way: horizontally,
 * vertically, or diagonally — Chebyshev distance of 1) and hold cards of equal
 * rank (`card.value`). This mirrors the backend rule in `internal/domain/MonteCarlo.go`
 * (`Remove` rejects `absInt(r1-r2) > 1 || absInt(c1-c2) > 1` and unequal `GetValue`).
 *
 * Each pair is counted exactly once: for every filled cell only the four
 * "forward" neighbours (right, down-left, down, down-right) are inspected,
 * matching the domain's `findAdjacentPair` scan order.
 *
 * @param board - The 5x5 Monte Carlo board grid.
 * @returns The number of distinct removable pairs on the board.
 */
export function countRemovablePairs(board: MonteCarloBoardCell[][]): number {
  // Forward-only directions so each unordered pair is visited once.
  const dirs: ReadonlyArray<readonly [number, number]> = [
    [0, 1], // right
    [1, -1], // down-left
    [1, 0], // down
    [1, 1], // down-right
  ];
  let count = 0;
  for (let r = 0; r < board.length; r++) {
    const row = board[r];
    for (let c = 0; c < row.length; c++) {
      const a = row[c]?.card;
      if (!a) continue;
      for (const [dr, dc] of dirs) {
        const nr = r + dr;
        const nc = c + dc;
        const b = board[nr]?.[nc]?.card;
        if (!b) continue;
        if (a.value === b.value) count++;
      }
    }
  }
  return count;
}
