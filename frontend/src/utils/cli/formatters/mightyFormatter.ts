import type { MightyResponse } from '../../../types/card';
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

/** Format a Mighty game state as terminal text. */
export function formatMightyState(state: MightyResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Mighty'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}`);
  if (state.winningBidNoTrump) {
    lines.push('trump: No-Trump');
  } else if (state.trumpSuit > 0) {
    lines.push(`trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}`);
  }
  if (state.highestBid > 0) lines.push(`bid: ${state.highestBid}`);
  if (state.partnerCard) lines.push(`partner: ${formatCard(state.partnerCard)}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const roles: string[] = [];
    if (p.isDeclarer) roles.push('Declarer');
    if (p.isPartner && p.partnerRevealed) roles.push('Partner');
    const roleStr = roles.length > 0 ? ` [${roles.join(', ')}]` : '';
    lines.push(`${name}: pts=${p.pointCards} tricks=${p.trickCount} total=${p.cumulativeScore}${roleStr}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.kitty && state.kitty.length > 0 && state.phase === 2) {
    lines.push(`kitty: ${formatCardList(state.kitty)}`);
  }

  if (state.currentTrick.length > 0) {
    const parts = state.currentTrick.map((tc) => {
      const playerName = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${playerName}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${parts.join(', ')}`);
  }

  if (state.hint && isRequestedHint(state)) lines.push(`HINT: ${state.hint.reason}`);

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) lines.push(`Game Over! Winner: Team ${state.winnerTeam}`);

  lines.push(formatSeparator());
  return lines.join('\n');
}
