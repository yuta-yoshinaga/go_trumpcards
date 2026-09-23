import type { SakuraResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Play', 'RoundEnd', 'GameEnd'];

/** Format a Sakura game state as terminal text. */
export function formatSakuraState(state: SakuraResponse): string {
  const lines = [formatHeader('Sakura')];
  lines.push(
    `round: ${state.round}/${state.totalRounds}  stock: ${state.stockCount}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`field: ${state.fieldCards.length ? state.fieldCards.map(formatCard).join(' ') : '(empty)'}`);
  for (const p of state.players) {
    lines.push(
      `${formatPlayerName(p.id, p.isHuman)}: hand=${p.cardCount} taken=${p.takenCount} points=${p.totalPoints} score=${p.score}`,
    );
    if (p.isHuman && p.cards.length) lines.push(`  ${formatIndexedCards(p.cards)}`);
  }
  if (state.phase === 1 && state.lastResult)
    lines.push(
      `result: winner=${state.lastResult.winner >= 0 ? formatPlayerName(state.lastResult.winner, state.players[state.lastResult.winner]?.isHuman ?? false) : 'draw'}`,
    );
  if (state.phase === 2)
    lines.push(
      `winner: ${state.winner >= 0 ? formatPlayerName(state.winner, state.players[state.winner]?.isHuman ?? false) : 'draw'}`,
    );
  if (state.hint && state.hint.cardIndex >= 0)
    lines.push(
      `HINT: play ${state.hint.cardIndex}${state.hint.fieldIndex >= 0 ? ` field ${state.hint.fieldIndex}` : ''} (${state.hint.reason})`,
    );
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
