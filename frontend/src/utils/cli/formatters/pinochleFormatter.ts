import type { PinochleResponse } from '../../../types/card';
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

/** Format a Pinochle game state as terminal text. */
export function formatPinochleState(state: PinochleResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Pinochle'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}`);
  if (state.trumpSuit > 0) lines.push(`trump: ${SUIT_NAMES[state.trumpSuit] ?? '?'}`);
  if (state.highestBid > 0) lines.push(`bid: ${state.highestBid}`);
  if (state.teamScores) lines.push(`score: Team0=${state.teamScores[0]} Team1=${state.teamScores[1]}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const parts = [`team=${p.team}`, `meld=${p.meldScore}`, `tricks=${p.trickPoints}`];
    if (p.bid > 0) parts.push(`bid=${p.bid}`);
    if (p.hasPassed) parts.push('[Passed]');
    lines.push(`${name}: ${parts.join(' ')}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.playerMelds) {
    for (let i = 0; i < state.playerMelds.length; i++) {
      const melds = state.playerMelds[i];
      if (melds.length > 0) {
        const name = formatPlayerName(i, state.players[i]?.isHuman ?? false);
        const meldStrs = melds.map((m) => `${m.points}pts(${formatCardList(m.cards)})`);
        lines.push(`${name} melds: ${meldStrs.join(', ')}`);
      }
    }
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
