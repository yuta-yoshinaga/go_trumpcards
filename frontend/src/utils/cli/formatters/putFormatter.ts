import type { PutResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASES = ['Play', 'Respond', 'TrickEnd', 'HandEnd', 'GameEnd'];

/** Format Put state as terminal text. */
export function formatPutState(state: PutResponse): string {
  const lines = [formatHeader('Put')];
  lines.push(`hand: ${state.handNumber}  trick: ${state.trickNumber}  phase: ${PHASES[state.phase] ?? state.phase}`);
  lines.push(
    `score: ${state.matchPoints[0]}-${state.matchPoints[1]} (first to ${state.matchTarget})  stake: ${state.handStake}`,
  );
  for (const player of state.players) {
    lines.push(`${formatPlayerName(player.id, player.isHuman)}: cards=${player.cardCount} tricks=${player.trickCount}`);
    if (player.isHuman && player.cards.length > 0) {
      lines.push(`  ${player.cards.map((card, index) => `[${index}]${formatCard(card)}`).join('  ')}`);
    }
  }
  if (state.currentTrick.length > 0) {
    const trick = state.currentTrick.map(({ playerIdx, card }) => {
      const player = state.players[playerIdx];
      return `${formatPlayerName(playerIdx, player?.isHuman ?? false)}=${formatCard(card)}`;
    });
    lines.push(`trick: ${trick.join(', ')}`);
  }
  if (state.hint && isRequestedHint(state)) {
    const action = state.hint.cardIndex === undefined ? state.hint.action : `card [${state.hint.cardIndex}]`;
    lines.push(`HINT: ${action} (${state.hint.reason})`);
  }
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag)
    lines.push(state.winnerIdx >= 0 ? `Game Over! Player ${state.winnerIdx} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
