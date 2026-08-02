import type { TarneebResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Translate trump-suit code (1-4) to a short label. */
function fmtTrump(suit: number): string {
  switch (suit) {
    case 1:
      return '♠';
    case 2:
      return '♣';
    case 3:
      return '♥';
    case 4:
      return '♦';
    default:
      return '-';
  }
}

/** Format a Tarneeb game state as terminal text. */
export function formatTarneebState(state: TarneebResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Tarneeb'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  trump: ${fmtTrump(state.trumpSuit)}  bid: ${state.highestBid || '-'}`,
  );
  lines.push(`team 0: ${state.teamScores[0] ?? 0}  team 1: ${state.teamScores[1] ?? 0}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name} (T${p.team}): bid=${p.bid >= 0 ? p.bid : '-'} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.phase === 0) lines.push('Bidding phase (7-13 or 0 to pass)');
  if (state.phase === 1) lines.push('Trump declaration phase (suit 1=♠ 2=♣ 3=♥ 4=♦)');

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.bid !== undefined) lines.push(`HINT: bid ${state.hint.bid} (${state.hint.reason})`);
    if (state.hint.trumpSuit !== undefined)
      lines.push(`HINT: trump ${fmtTrump(state.hint.trumpSuit)} (${state.hint.reason})`);
    if (state.hint.cardIndex !== undefined) lines.push(`HINT: play [${state.hint.cardIndex}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game Over — team ${state.winnerTeam} wins`);

  lines.push(formatSeparator());
  return lines.join('\n');
}
