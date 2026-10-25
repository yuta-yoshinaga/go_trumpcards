import type { BeloteResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['BidPickUp', 'BidCallTrump', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUITS = ['None', 'Spades', 'Clubs', 'Hearts', 'Diamonds'];

/** Format a Belote game state as terminal text. */
export function formatBeloteState(state: BeloteResponse): string {
  const lines = [
    formatHeader('Belote'),
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
    `trump: ${SUITS[state.trumpSuit] ?? state.trumpSuit}  scores: team0=${state.teamScores[0]} team1=${state.teamScores[1]}`,
    `round points: team0=${state.roundPoints[0]} team1=${state.roundPoints[1]}  belote: team0=${state.roundBeloteBonus[0]} team1=${state.roundBeloteBonus[1]}`,
    '',
  ];
  for (const player of state.players) {
    lines.push(
      `${formatPlayerName(player.id, player.isHuman)} (team ${player.team}): cards=${player.cardCount} tricks=${player.trickCount}`,
    );
    if (player.isHuman && player.cards.length) lines.push(`  ${formatIndexedCards(player.cards)}`);
  }
  lines.push('----------');
  if (state.faceUpCard) lines.push(`face up: ${formatCard(state.faceUpCard)}`);
  if (state.currentTrick.length) {
    lines.push(
      `trick: ${state.currentTrick.map(({ playerIdx, card }) => `${formatPlayerName(playerIdx, state.players[playerIdx]?.isHuman ?? false)}=${formatCard(card)}`).join(', ')}`,
    );
  }
  if (state.hint && isRequestedHint(state)) {
    const advice =
      state.hint.cardIndex !== undefined
        ? `card index ${state.hint.cardIndex}`
        : state.hint.orderUp
          ? 'order up'
          : state.hint.suit !== undefined
            ? `call trump ${state.hint.suit}`
            : 'pass';
    lines.push(`HINT: ${advice} (${state.hint.reason})`);
  }
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag)
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
