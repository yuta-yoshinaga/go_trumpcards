import type { TwoTenJackResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

/** Format a Two Ten Jack game state as terminal text. */
export function formatTwoTenJackState(state: TwoTenJackResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Two Ten Jack'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}  trump: ${state.trumpSuit}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(
      `${name}: total=${p.cumulativeScore} round=${p.roundScore} tricks=${p.trickCount} pts=${p.capturedPoints}`,
    );
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

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push('Game Over');

  lines.push(formatSeparator());
  return lines.join('\n');
}
