import type { WhistResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Format a Whist game state as terminal text. */
export function formatWhistState(state: WhistResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Whist'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}`);
  if (state.trumpSuit > 0) lines.push(`trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}`);
  lines.push(`score: Team0=${state.teamScores[0]} Team1=${state.teamScores[1]}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: team=${p.team} total=${p.cumulativeScore} round=${p.roundScore} tricks=${p.trickCount}`);
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

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.cardIndex !== undefined) lines.push(`HINT: play [${state.hint.cardIndex}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game Over! Winner: Team ${state.winnerTeam}`);

  lines.push(formatSeparator());
  return lines.join('\n');
}
