import type { MonteCarloResponse } from '../../../types/card';
import { MonteCarloPhase } from '../../../types/phases';
import { canMonteCarloDeal, canMonteCarloUndo, listMonteCarloRemovablePairs } from '../../monteCarloRules';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Readonly<Record<number, string>> = {
  [MonteCarloPhase.PLAYING]: 'Playing',
  [MonteCarloPhase.GAME_CLEAR]: 'Game Clear',
  [MonteCarloPhase.GAME_OVER]: 'Game Over',
};

/** Format a Monte Carlo Solitaire state for terminal play. */
export function formatMonteCarloState(state: MonteCarloResponse): string {
  const lines = [formatHeader('Monte Carlo'), `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`];
  lines.push(`stock: ${state.stockCount} | removed: ${state.removedCount}/52 | deals: ${state.dealCount}`);
  const pairs = listMonteCarloRemovablePairs(state);
  const canDeal = canMonteCarloDeal(state);
  const canUndo = canMonteCarloUndo(state);
  lines.push(`removable pairs: ${pairs.length} | undo available: ${canUndo ? 'yes' : 'no'}`);
  lines.push('Board (choose two adjacent matching indices with p <a> <b>):');
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
  if (state.phase === MonteCarloPhase.PLAYING) {
    const actions = [
      ...pairs.map(([a, b]) => `p ${a} ${b}`),
      ...(canDeal ? ['d (deal)'] : []),
      ...(canUndo ? ['u (undo)'] : []),
      'h (hint)',
      'g (give up)',
      'r (reset)',
    ];
    lines.push(`Available commands: ${actions.join(', ')}`);
  } else {
    if (state.phase === MonteCarloPhase.GAME_CLEAR) lines.push('All cards removed. Game clear!');
    if (state.phase === MonteCarloPhase.GAME_OVER) lines.push('No more moves. Game over.');
    lines.push('Available command: r (reset)');
  }
  if (state.messageCode === 'montecarlo.noHint') lines.push('No hint available.');
  if (state.hint && isRequestedHint(state)) {
    if (state.hint.action === 'deal') {
      if (canDeal) lines.push('HINT: Deal cards from the stock.');
    } else if (
      state.hint.fromR !== undefined &&
      state.hint.fromC !== undefined &&
      state.hint.toR !== undefined &&
      state.hint.toC !== undefined
    ) {
      lines.push(
        `HINT: Remove [${state.hint.fromR * 5 + state.hint.fromC}] and [${state.hint.toR * 5 + state.hint.toC}].`,
      );
    }
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
