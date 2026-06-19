import type { Card, FaroResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  1: 'Betting',
  2: 'Turn',
  3: 'Call',
  4: 'RoundEnd',
  5: 'GameEnd',
};

/** Render a single card, or a dash when null. */
function card(c: Card | null): string {
  return c ? formatCard(c) : '-';
}

/** Format a Faro game state as terminal text. */
export function formatFaroState(state: FaroResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Faro'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? state.phase}  chips: ${state.chips}`);
  lines.push(`turns: ${state.turnsPlayed}/${state.turnsTotal}  cards left: ${state.remaining}`);
  if (state.soda) lines.push(`soda (burned): ${card(state.soda)}`);
  lines.push('----------');

  if (state.bets.length > 0) {
    lines.push('BETS');
    for (const b of state.bets) {
      lines.push(`  rank ${b.rank}: ${b.amount}${b.copper ? ' (copper)' : ''}`);
    }
  } else {
    lines.push('no bets placed');
  }

  if (state.losingCard || state.winningCard) {
    lines.push('----------');
    lines.push(`losing card: ${card(state.losingCard)}`);
    lines.push(`winning card: ${card(state.winningCard)}`);
    if (state.split) lines.push('split (bank takes half)');
  }

  if (state.callCards.length > 0) {
    lines.push('----------');
    lines.push(`call (last 3): ${state.callCards.map(formatCard).join(', ')}`);
  }

  if (state.phase === 4 || state.phase === 5) {
    if (state.callOrder.length > 0) {
      lines.push(state.callWon ? 'Call won!' : 'Call lost.');
    }
    lines.push(`round net: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Out of chips. Game over.');

  lines.push(formatSeparator());
  return lines.join('\n');
}
