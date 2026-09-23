import type { OmiResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Format an Omi game state as terminal text. */
export function formatOmiState(state: OmiResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Omi'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}`);
  if (state.trumpSuit > 0) {
    const caller = formatPlayerName(state.trumpCallerIdx, state.players[state.trumpCallerIdx]?.isHuman ?? false);
    lines.push(`trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}  caller: ${caller}`);
  }
  if (state.teamScores.length >= 2) lines.push(`score: Team0=${state.teamScores[0]} Team1=${state.teamScores[1]}`);
  if (state.teamTricks?.length >= 2) lines.push(`tricks: Team0=${state.teamTricks[0]} Team1=${state.teamTricks[1]}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const dealInfo = state.dealStage === 1 ? ' (4 cards — awaiting trump)' : '';
    lines.push(`${name}: team=${p.team} tricks=${p.trickCount}${dealInfo}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      const suitName = tc.card.design ? tc.card.design.charAt(0) : '?';
      return `${name}=${suitName}${tc.card.value}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  // Scoring rules (shown once after trump is set)
  if (state.trumpSuit > 0) {
    lines.push('--- scoring ---');
    lines.push('5+ tricks: 1pt  8 tricks (Omi!): 2pts  4-4: 0pts each');
  }

  if (state.hint && isRequestedHint(state)) lines.push(`HINT: ${state.hint.reason}`);

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game Over! Winner: Team ${state.winnerTeam}`);

  lines.push(formatSeparator());
  return lines.join('\n');
}
