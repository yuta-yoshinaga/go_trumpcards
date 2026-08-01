import type { HeartsResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PASS_DIRECTIONS: Record<number, string> = {
  0: 'Left',
  1: 'Right',
  2: 'Across',
  3: 'No Pass',
};

/** Format a Hearts game state as terminal text. */
export function formatHeartsState(state: HeartsResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Hearts'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  hearts broken: ${state.heartsBroken ? 'yes' : 'no'}`,
  );
  lines.push('');

  // Players
  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  // Current trick
  if (state.currentTrick.length > 0) {
    const trickParts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  // Phase-specific info
  if (state.phase === 0) {
    lines.push(`Pass phase: ${PASS_DIRECTIONS[state.passDirection] ?? 'Unknown'}`);
  }

  // Hint
  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
