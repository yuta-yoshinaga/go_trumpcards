import type { KingResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'];

/** Contract names indexed by contract number (0..6). */
const CONTRACT_NAMES = [
  'No Tricks',
  'No Hearts',
  'No Queens',
  'No King of Hearts',
  'No Last Two',
  'No Men',
  'King (Trump)',
];

/** Format a King game state as terminal text. */
export function formatKingState(state: KingResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('King'));
  lines.push(`deal: ${state.dealNumber + 1}/${state.totalDeals}  trick: ${state.trickNumber}  phase: ${state.phase}`);

  const contract = state.currentContract >= 0 ? (CONTRACT_NAMES[state.currentContract] ?? state.currentContract) : '-';
  const trump = state.trumpSuit >= 1 ? (SUIT_SYMBOLS[state.trumpSuit] ?? '-') : '-';
  lines.push(`contract: ${contract}  trump: ${trump}  dealer: P${state.dealerIdx}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: cards=${p.cardCount} tricks=${p.trickCount} score=${p.totalScore}`);
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

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.roundWinners.length > 0) {
    lines.push(`Game Over! Winner(s): ${state.roundWinners.map((w) => `Player ${w}`).join(', ')}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
