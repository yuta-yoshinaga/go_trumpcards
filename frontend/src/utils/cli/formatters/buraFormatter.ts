import type { BuraPlayer, BuraResponse } from '../../../types/card';
import { formatCard, formatHeader } from '../formatterBase';

/**
 * Render one seat. Visibility comes from the server's `hidden` flag, never
 * from re-deriving it here: a hidden seat arrives with no cards at all, only a
 * count, so the backs are drawn from `cardCount`.
 */
function formatSeat(p: BuraPlayer): string {
  const cards = p.hidden
    ? Array.from({ length: p.cardCount }, () => '[??]').join(' ')
    : p.cards.map((c, i) => `${i}:${formatCard(c)}`).join(' ');
  return `${p.isHuman ? 'you' : 'cpu'}: ${p.points} pt  ${cards}`;
}

/** Format a Bura game state as terminal text. */
export function formatBuraState(state: BuraResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Bura'));
  lines.push(`trick ${state.trickNumber} / stock ${state.stockRemaining} / target ${state.winThreshold}`);
  lines.push(state.trumpCard ? `trump: ${formatCard(state.trumpCard)}` : `trump suit: ${state.trumpSuit}`);
  lines.push('----------');

  for (const p of state.players) {
    lines.push(formatSeat(p));
  }

  if (state.currentLead.length > 0) {
    lines.push(`led: ${state.currentLead.map(formatCard).join(' ')}`);
  }

  if (state.gameEndFlag) {
    if (state.isDraw || state.winnerIdx < 0) {
      lines.push('draw -- nobody claimed');
    } else {
      lines.push(state.winnerIdx === 0 ? 'you win' : 'you lose');
    }
  }

  return lines.join('\n');
}
