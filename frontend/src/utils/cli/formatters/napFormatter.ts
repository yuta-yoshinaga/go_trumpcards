import type { NapResponse } from '../../../types/card';
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

/** Maps a Nap contract/bid value (0/2/3/4/5) to its display name. */
function contractName(value: number): string {
  switch (value) {
    case 0:
      return 'Pass';
    case 2:
      return 'Two';
    case 3:
      return 'Three';
    case 4:
      return 'Four';
    case 5:
      return 'Nap';
    default:
      return String(value);
  }
}

/** Format a Nap (Napoleon) game state as terminal text. */
export function formatNapState(state: NapResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Nap'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}`);
  if (state.declarerIdx >= 0) {
    const name = formatPlayerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false);
    lines.push(`declarer: ${name} — ${contractName(state.contract)}`);
  } else {
    lines.push(`bids: ${state.bids.map((b, i) => `P${i}=${contractName(b)}`).join('  ')}`);
  }
  lines.push(`chips: ${state.playerScores.map((s, i) => `P${i}=${s}`).join('  ')}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const role = p.isDeclarer ? 'Declarer' : 'Defender';
    lines.push(`${name} (${role}): cards=${p.cardCount} tricks=${p.trickCount} chips=${p.score}`);
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
