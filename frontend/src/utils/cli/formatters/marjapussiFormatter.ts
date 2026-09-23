import type { MarjapussiResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'];

/** Format a Marjapussi game state as terminal text. */
export function formatMarjapussiState(state: MarjapussiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Marjapussi'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '-'}`);
  const team0Score = state.teamScores[0] ?? 0;
  const team1Score = state.teamScores[1] ?? 0;
  lines.push(`team scores: Team0=${team0Score}  Team1=${team1Score} (target: ${state.config.targetPoints})`);
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push(`pussi: ${state.pussiCount} cards`);
  if (state.pussiWinnerTeam >= 0) {
    lines.push(`pussi won by: Team ${state.pussiWinnerTeam}`);
  }
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const team = `Team ${p.teamId}`;
    lines.push(`${name} (${team}): cards=${p.cardCount} tricks=${p.trickCount} score=${p.score}`);
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
    const cardPts = state.roundCardPoints.map((v, i) => `Team${i}=${v}`).join(' ');
    const marriage = state.roundMarriage.map((v, i) => `Team${i}=${v}`).join(' ');
    lines.push(`round result: card pts ${cardPts}`);
    lines.push(`marriage: ${marriage}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game Over! Winner: Team ${state.winnerTeam}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
