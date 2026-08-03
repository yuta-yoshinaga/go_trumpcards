import type { ViraResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_SYMBOLS = ['none', '♠', '♣', '♥', '♦'];
const CONTRACT_NAMES = ['Pass', 'Six', 'Misère', 'Seven', 'Eight'];

/** Format a Vira game state as terminal text. */
export function formatViraState(state: ViraResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Vira'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}`);
  if (state.declarerIdx >= 0) {
    const name = formatPlayerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false);
    lines.push(`declarer: ${name} — ${CONTRACT_NAMES[state.contract] ?? '?'}`);
  } else {
    lines.push(`bids: ${state.bids.map((b, i) => `P${i}=${CONTRACT_NAMES[b] ?? b}`).join('  ')}`);
  }
  lines.push(`scores: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDeclarer ? 'Declarer' : 'Defender';
    lines.push(`${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} score=${p.score}`);
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

  if (state.phase === 3 || state.phase === 4) {
    const tricks = state.roundTricks.map((v, i) => `P${i}=${v}`).join(' ');
    lines.push(`round result: tricks ${tricks}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerPlayer >= 0) {
    lines.push(`Game Over! Winner: Player ${state.winnerPlayer}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
