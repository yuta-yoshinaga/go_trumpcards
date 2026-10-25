import type { BrusquembilleResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

/** Format a Brusquembille game state as terminal text. */
export function formatBrusquembilleState(state: BrusquembilleResponse): string {
  const lines = [formatHeader('Brusquembille')];
  lines.push(`trick: ${state.trickNumber}  stock: ${state.stockRemaining}  phase: ${state.phase}`);
  lines.push(`trump: ${state.trumpCard ? formatCard(state.trumpCard) : 'none'}`);
  for (const player of state.players) {
    lines.push(`${formatPlayerName(player.id, player.isHuman)}: points=${player.points} tricks=${player.trickCount}`);
    if (player.isHuman && player.cards.length > 0) {
      lines.push(`  ${player.cards.map((card, index) => `[${index}]${formatCard(card)}`).join('  ')}`);
    }
  }
  if (state.currentTrick.length) {
    lines.push(
      `trick cards: ${state.currentTrick.map(({ playerIdx, card }) => `${formatPlayerName(playerIdx, state.players[playerIdx]?.isHuman ?? false)}=${formatCard(card)}`).join(', ')}`,
    );
  }
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      state.hint.cardIndex !== undefined
        ? `HINT: play card [${state.hint.cardIndex}] (${state.hint.reason})`
        : `HINT: ${state.hint.reason}`,
    );
  }
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag)
    lines.push(state.winnerIdx >= 0 ? `Game Over! Player ${state.winnerIdx} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
