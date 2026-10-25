import type { LaBelleLucieResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASES: Record<number, string> = { 0: 'Playing', 1: 'Game Clear', 2: 'Game Over' };

/** Format a La Belle Lucie game state as terminal text. */
export function formatLaBelleLucieState(state: LaBelleLucieResponse): string {
  const lines = [
    formatHeader('La Belle Lucie'),
    `phase: ${PHASES[state.phase] ?? state.phase} | moves: ${state.moveCount} | redeals: ${state.redealsLeft}`,
    'Foundation:',
  ];
  lines.push(
    state.foundation.map((pile, i) => `  ${i}: ${pile.length ? formatCard(pile[pile.length - 1]) : '[  ]'}`).join('\n'),
  );
  lines.push('Fans:');
  state.fans.forEach((fan, i) => {
    lines.push(`  ${i}: ${formatIndexedCards(fan) || '[empty]'}${state.movableFans[i] ? ' *' : ''}`);
  });
  if (state.phase === 0) {
    lines.push('Moves: use p <from> <to> or f <idx> for a fan marked *');
    lines.push(state.redealsLeft > 0 ? 'Other actions: rd, ac, u, h, g' : 'Other actions: ac, u, h, g');
  } else {
    lines.push('Game finished. Use r to start a new game.');
  }
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      state.hint.toFoundation
        ? `HINT: move fan ${state.hint.fromFan} to foundation`
        : `HINT: move fan ${state.hint.fromFan} to fan ${state.hint.toFan}`,
    );
  }
  if (state.message) lines.push(state.message);
  if (state.phase === 1) lines.push('Congratulations! You win!');
  if (state.phase === 2) lines.push('Game over.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
