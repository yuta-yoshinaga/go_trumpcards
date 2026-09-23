import type { Card } from '../../../types/common';
import type { BassetResponse } from '../../../types/games/basset';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'Betting', 2: 'Turn', 3: 'Decision', 4: 'RoundEnd', 5: 'GameEnd' };
const card = (value: Card | null) => (value ? formatCard(value) : '-');

/** Formats a Basset response for the terminal. */
export function formatBassetState(state: BassetResponse): string {
  const lines = [
    formatHeader('Basset'),
    `phase: ${PHASE_NAMES[state.phase] ?? state.phase}  chips: ${state.chips}`,
    `turns: ${state.turnsPlayed}/${state.turnsTotal}  cards left: ${state.remaining}`,
    '----------',
  ];
  if (state.bet) lines.push(`bet: rank ${state.bet.rank} amount ${state.bet.amount} stage ${state.bet.stage}`);
  else lines.push('no bet placed');
  if (state.bankerCard || state.playerCard)
    lines.push(`banker card: ${card(state.bankerCard)}`, `player card: ${card(state.playerCard)}`);
  if (state.phase === 3) lines.push('decision: take or paroli');
  if (state.message) lines.push(state.message);
  lines.push(`round net: ${state.totalPayout}`, formatSeparator());
  return lines.join('\n');
}
