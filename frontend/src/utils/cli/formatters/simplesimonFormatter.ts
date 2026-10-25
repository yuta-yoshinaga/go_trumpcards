import type { SimpleSimonResponse } from '../../../types/card';
import { SimpleSimonPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

/** Formats Simple Simon state as terminal text. */
export function formatSimpleSimonState(state: SimpleSimonResponse): string {
  const phase =
    state.phase === SimpleSimonPhase.PLAYING
      ? 'Playing'
      : state.phase === SimpleSimonPhase.GAME_CLEAR
        ? 'GameClear'
        : 'GameOver';
  const lines = [
    formatHeader('Simple Simon'),
    `phase: ${phase}  moves: ${state.moveCount}  completed suits: ${state.completedSuits}`,
  ];
  state.columns.forEach((column, col) => {
    const cards = column.map((card, idx) => {
      const movable = state.phase === SimpleSimonPhase.PLAYING && state.runStarts[col]?.includes(idx);
      return `${movable ? '*' : ' '}${formatCard(card)}[${idx}]`;
    });
    lines.push(`C${col}: ${cards.join(' ') || '(empty)'}`);
  });
  if (state.phase === SimpleSimonPhase.PLAYING) lines.push('Commands: m <fromCol> <cardIndex> <toCol>, h, g, u, r');
  else lines.push('Commands: r');
  if (state.hint && isRequestedHint(state)) {
    lines.push(`HINT: move [${state.hint.fromCol}] [${state.hint.cardIndex}] to column ${state.hint.toCol}.`);
  } else if (state.messageCode === 'simplesimon.noHint') lines.push('No hint available.');
  if (state.phase !== SimpleSimonPhase.PLAYING)
    lines.push(state.phase === SimpleSimonPhase.GAME_CLEAR ? 'Game Clear!' : 'Game Over.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
