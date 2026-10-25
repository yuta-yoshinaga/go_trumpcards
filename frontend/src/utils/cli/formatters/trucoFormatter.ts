import type { TrucoResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASES = ['Play', 'Respond', 'TrickEnd', 'HandEnd', 'GameEnd'];

/** Format a Truco game state as terminal text. */
export function formatTrucoState(state: TrucoResponse): string {
  const lines = [
    formatHeader('Truco'),
    `hand: ${state.handNumber}  trick: ${state.trickNumber}  phase: ${PHASES[state.phase] ?? state.phase}`,
    `score: You=${state.matchPoints[0]}  CPU=${state.matchPoints[1]} (first to ${state.matchTarget})`,
    `stake: ${state.handStake}  level: ${state.acceptedLevel}`,
    '',
  ];
  state.players.forEach((player) => {
    lines.push(`${formatPlayerName(player.id, player.isHuman)}: cards=${player.cardCount} tricks=${player.trickCount}`);
    if (player.isHuman) lines.push(`  ${player.cards.map((card, idx) => `[${idx}]${formatCard(card)}`).join('  ')}`);
  });
  lines.push(formatSeparator());
  if (state.currentTrick.length)
    lines.push(
      `trick: ${state.currentTrick.map(({ playerIdx, card }) => `${formatPlayerName(playerIdx, state.players[playerIdx]?.isHuman ?? false)}=${formatCard(card)}`).join(', ')}`,
    );
  if (state.hint && isRequestedHint(state)) {
    const { action, cardIndex, reason } = state.hint;
    const card = cardIndex === undefined ? undefined : state.players.find((p) => p.isHuman)?.cards[cardIndex];
    const suggestion =
      action === 'play' && cardIndex !== undefined
        ? `play card [${cardIndex}]${card ? ` ${formatCard(card)}` : ''}`
        : action;
    lines.push(`HINT: ${suggestion} (${reason})`);
  }
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game Over! ${state.winnerIdx === 0 ? 'You win!' : 'CPU wins!'}`);
  lines.push(formatSeparator());
  return lines.join('\n');
}
