import type { MonteCarloResponse } from '../types/card';
import { MonteCarloPhase } from '../types/phases';

/** A row and column position on the Monte Carlo board. */
export interface MonteCarloCellPos {
  row: number;
  col: number;
}

/** A removable pair represented by the board's flat, zero-based indices. */
export type MonteCarloPair = readonly [number, number];

/** Reports whether two different board positions are neighbors, including diagonals. */
export function isMonteCarloAdjacent(a: MonteCarloCellPos, b: MonteCarloCellPos): boolean {
  return Math.abs(a.row - b.row) <= 1 && Math.abs(a.col - b.col) <= 1 && !(a.row === b.row && a.col === b.col);
}

/** Reports whether a player can deal while the game is in its playing phase. */
export function canMonteCarloDeal(state: MonteCarloResponse | null | undefined): boolean {
  return state?.phase === MonteCarloPhase.PLAYING && state.stockCount > 0;
}

/** Reports whether the current state allows undo. */
export function canMonteCarloUndo(state: MonteCarloResponse | null | undefined): boolean {
  return state?.phase === MonteCarloPhase.PLAYING && state.canUndo;
}

/** Lists all adjacent, same-rank occupied-cell pairs in flat board-index form. */
export function listMonteCarloRemovablePairs(state: MonteCarloResponse): MonteCarloPair[] {
  const pairs: MonteCarloPair[] = [];
  state.board.forEach((row, rowIndex) => {
    row.forEach((cell, colIndex) => {
      if (cell.card) {
        const first = { row: rowIndex, col: colIndex };
        for (let otherRow = rowIndex; otherRow < state.board.length; otherRow += 1) {
          const startCol = otherRow === rowIndex ? colIndex + 1 : 0;
          for (let otherCol = startCol; otherCol < (state.board[otherRow]?.length ?? 0); otherCol += 1) {
            const otherCard = state.board[otherRow]?.[otherCol]?.card;
            if (otherCard?.value === cell.card.value && isMonteCarloAdjacent(first, { row: otherRow, col: otherCol })) {
              pairs.push([rowIndex * 5 + colIndex, otherRow * 5 + otherCol]);
            }
          }
        }
      }
    });
  });
  return pairs;
}
