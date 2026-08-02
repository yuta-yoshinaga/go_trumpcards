import type { LooResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Decide', 'Play', 'TrickEnd', 'RoundEnd'];
const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'];

/** Format a Loo (Lanterloo) game state as terminal text. */
export function formatLooState(state: LooResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Loo'));
  lines.push(
    `deal: ${state.roundNumber}  trick: ${state.trickNumber}/${state.totalTricks}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`pot: ${state.pot}  trump: ${SUIT_SYMBOLS[state.trumpSuit] ?? '-'}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status = p.playing ? 'Play' : 'Pass';
    lines.push(`${name}: ${status}  tricks=${p.trickCount}  chips=${p.chips}`);
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

  if (state.phase === 3 && state.lastDealDetail) {
    const d = state.lastDealDetail;
    const gained = state.players.map((p) => `P${p.id}=${d.gained[p.id] ?? 0}`).join(' ');
    lines.push(`deal result: gained ${gained}`);
    if (d.looed.length > 0) lines.push(`looed: ${d.looed.map((i) => `P${i}`).join(', ')}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
