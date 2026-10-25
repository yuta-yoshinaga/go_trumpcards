import type { JassResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['BidTrump', 'BidPartner', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUITS: Record<number, string> = { 0: 'No trump', 1: 'Spades', 2: 'Clubs', 3: 'Hearts', 4: 'Diamonds' };

function formatJassCard(card: JassResponse['players'][number]['cards'][number]): string {
  return formatCard(card);
}

/** Format a Jass game state as terminal text. */
export function formatJassState(state: JassResponse): string {
  const lines = [formatHeader('Jass')];
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(
    `trump: ${SUITS[state.trumpSuit] ?? state.trumpSuit}  scores: team0=${state.teamScores[0]} team1=${state.teamScores[1]} (first to ${state.config.targetScore})`,
  );
  lines.push('');
  for (const player of state.players) {
    lines.push(
      `${formatPlayerName(player.id, player.isHuman)} (team ${player.team}): cards=${player.cardCount} tricks=${player.trickCount}`,
    );
    if (player.isHuman && player.cards.length > 0) {
      lines.push(`  ${player.cards.map((card, i) => `[${i}]${formatJassCard(card)}`).join('  ')}`);
    }
  }
  lines.push('----------');
  if (state.currentTrick.length > 0) {
    lines.push(
      `trick: ${state.currentTrick.map(({ playerIdx, card }) => `${formatPlayerName(playerIdx, state.players[playerIdx]?.isHuman ?? false)}=${formatJassCard(card)}`).join(', ')}`,
    );
  }
  if (state.lastTrick.length > 0) {
    lines.push(
      `last trick: ${state.lastTrick.map(({ playerIdx, card }) => `${formatPlayerName(playerIdx, state.players[playerIdx]?.isHuman ?? false)}=${formatJassCard(card)}`).join(', ')} (winner P${state.lastTrickWinner})`,
    );
  }
  if (isRequestedHint(state) && state.hint)
    lines.push(
      `HINT: ${state.hint.cardIndex != null ? `card [${state.hint.cardIndex}]` : state.hint.schieben ? 'schieben' : `suit ${state.hint.suit}`} (${state.hint.reason})`,
    );
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag)
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
