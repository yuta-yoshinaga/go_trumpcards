import type { MonteCarloResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Playing', 'Game Clear', 'Game Over'];

/** Format a Monte Carlo Solitaire state for terminal play. */
export function formatMonteCarloState(state: MonteCarloResponse): string {
  const lines = [formatHeader('Monte Carlo'), `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`];
  lines.push(`stock: ${state.stockCount} | removed: ${state.removedCount}/52 | deals: ${state.dealCount}`);
  lines.push(`removable pairs: ${state.removablePairs} | undo available: ${state.canUndo ? 'yes' : 'no'}`);
  lines.push('Board (choose two adjacent matching indices with p <idx1> <idx2>):');
  state.board.forEach((row, rowIndex) => {
    lines.push(
      row
        .map((cell, colIndex) => {
          const index = rowIndex * 5 + colIndex;
          return `[${index}]${cell.card ? formatCard(cell.card) : '[empty]'}`;
        })
        .join(' '),
    );
  });
  if (state.phase === 0) {
    lines.push(
      state.removablePairs > 0
        ? 'Next: remove a matching adjacent pair.'
        : state.stockCount > 0
          ? 'Next: deal cards with d.'
          : 'Next: undo or give up.',
    );
  } else if (state.phase === 1) {
    lines.push('All cards removed. Game clear!');
  } else if (state.phase === 2) {
    lines.push('No more moves. Game over.');
  }
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      state.hint.action === 'deal'
        ? 'HINT: Deal cards from the stock.'
        : `HINT: Remove [${(state.hint.fromR ?? 0) * 5 + (state.hint.fromC ?? 0)}] and [${(state.hint.toR ?? 0) * 5 + (state.hint.toC ?? 0)}].`,
    );
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
