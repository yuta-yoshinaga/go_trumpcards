import type { NapoleonResponse } from '../../../types/card';
import {
  formatCard,
  formatCardList,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Format a Napoleon game state as terminal text. */
export function formatNapoleonState(state: NapoleonResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Napoleon'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}`);
  if (state.trumpSuit > 0) lines.push(`trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}`);
  if (state.highestBid > 0) lines.push(`bid: ${state.highestBid}`);
  if (state.adjutantCard) lines.push(`adjutant: ${formatCard(state.adjutantCard)}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const roles: string[] = [];
    if (p.isNapoleon) roles.push('Napoleon');
    if (p.isAdjutant && p.adjutantRevealed) roles.push('Adjutant');
    const roleStr = roles.length > 0 ? ` [${roles.join(', ')}]` : '';
    lines.push(`${name}: pic=${p.pictureCards} tricks=${p.trickCount} total=${p.cumulativeScore}${roleStr}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.kitty.length > 0 && state.phase === 2) {
    lines.push(`kitty: ${formatCardList(state.kitty)}`);
  }

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
