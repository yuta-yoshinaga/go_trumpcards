import type { KnockoutWhistResponse } from '../../../types/card';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_SYMBOLS = ['?', '♠', '♣', '♥', '♦'];

/** Format a Knockout Whist game state as terminal text (shows trump, hand size, Dogbones, and elimination). */
export function formatKnockoutWhistState(state: KnockoutWhistResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Knockout Whist'));
  lines.push(
    `round: ${state.roundNumber}  hand: ${state.handSize}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '?'}  active: ${state.activeCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    if (p.eliminated) {
      lines.push(`${name}: ELIMINATED`);
      continue;
    }
    lines.push(`${name}: cards=${p.cardCount} roundTricks=${p.roundTricks} dogbones=${p.dogbones}`);
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

  if (state.hint) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerPlayer >= 0) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.winnerPlayer, state.players[state.winnerPlayer]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
