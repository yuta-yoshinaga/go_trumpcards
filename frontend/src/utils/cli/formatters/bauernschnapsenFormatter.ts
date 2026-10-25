import type { BauernschnapsenResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const CONTRACTS: Record<number, string> = { 0: 'Undeclared', 1: 'Rufer', 2: 'Farbenzwang', 3: 'Bettel' };
const SUITS: Record<number, string> = { 1: 'Spades', 2: 'Clubs', 3: 'Hearts', 4: 'Diamonds' };

/** Format a Bauernschnapsen game state as terminal text. */
export function formatBauernschnapsenState(state: BauernschnapsenResponse): string {
  const lines = [
    formatHeader('Bauernschnapsen'),
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${state.phase}`,
    `contract: ${CONTRACTS[state.contract] ?? 'Unknown'}${state.trumpSuit ? `  trump: ${SUITS[state.trumpSuit] ?? '?'}` : ''}`,
    `score: Team0=${state.teamScores[0]} Team1=${state.teamScores[1]} (target ${state.config.targetScore})`,
    '',
  ];
  for (const p of state.players) {
    lines.push(`${formatPlayerName(p.id, p.isHuman)}: team=${p.team} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length) lines.push(`  ${formatIndexedCards(p.cards)}  (use p <idx>)`);
  }
  lines.push('----------');
  if (state.currentTrick.length)
    lines.push(
      `trick: ${state.currentTrick.map((tc) => `${formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false)}=${formatCard(tc.card)}`).join(', ')}`,
    );
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      state.hint.isMarriage
        ? `HINT: declare a marriage with card [${state.hint.cardIndex ?? '?'}] (${state.hint.reason})`
        : state.hint.cardIndex !== undefined
          ? `HINT: play card [${state.hint.cardIndex}] (${state.hint.reason})`
          : `HINT: ${state.hint.reason}`,
    );
  }
  if (state.message) lines.push(state.message);
  if (state.gameEndFlag)
    lines.push(state.winnerTeam >= 0 ? `Game Over! Team ${state.winnerTeam} wins!` : 'Game Over! Draw.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
