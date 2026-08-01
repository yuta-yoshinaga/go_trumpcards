import type { KlaverjasResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_SYMBOLS = ['?', '♠', '♣', '♥', '♦'];

/** Format a Klaverjas game state as terminal text. */
export function formatKlaverjasState(state: KlaverjasResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Klaverjas'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}`);
  lines.push(`score: A=${state.teamScores[0] ?? 0}  B=${state.teamScores[1] ?? 0}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: cards=${p.cardCount} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trickParts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  if (state.phase === 2 || state.phase === 3) {
    lines.push(`round result: A card pts=${state.roundCardPoints[0] ?? 0} B card pts=${state.roundCardPoints[1] ?? 0}`);
    lines.push(`roem: A=${state.roundRoem[0] ?? 0} B=${state.roundRoem[1] ?? 0}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game Over! Winner: Team ${state.winnerTeam === 0 ? 'A' : 'B'}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
