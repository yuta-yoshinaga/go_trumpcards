import type { EuchreResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Format a Euchre game state as terminal text. */
export function formatEuchreState(state: EuchreResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Euchre'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}`);
  if (state.trumpSuit > 0) lines.push(`trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}`);
  if (state.teamScores.length >= 2) lines.push(`score: Team0=${state.teamScores[0]} Team1=${state.teamScores[1]}`);
  if (state.goingAlone) lines.push('Going alone!');
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: team=${p.team} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.faceUpCard) lines.push(`face-up: ${formatCard(state.faceUpCard)}`);

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.hint && isRequestedHint(state)) lines.push(`HINT: ${state.hint.reason}`);

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game Over! Winner: Team ${state.winnerTeam}`);

  lines.push(formatSeparator());
  return lines.join('\n');
}
