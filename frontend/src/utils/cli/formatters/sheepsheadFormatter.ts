import type { SheepsheadResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Pick', 'Bury', 'Call', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_NAMES = ['none', '♠', '♣', '♥'];

/** Format a Sheepshead game state as terminal text. */
export function formatSheepsheadState(state: SheepsheadResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Sheepshead'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );
  lines.push(`blind: ${state.blindCount} cards`);

  const pickerLabel = state.pickerIdx >= 0 ? formatPlayerName(state.pickerIdx, false) : '-';
  const partnerLabel = state.partnerRevealed && state.partnerIdx >= 0 ? formatPlayerName(state.partnerIdx, false) : '?';
  lines.push(`picker: ${pickerLabel}  partner: ${partnerLabel}  calledSuit: ${SUIT_NAMES[state.calledSuit] ?? '-'}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: cards=${p.cardCount} tricks=${p.trickCount} chips=${p.chips}`);
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

  if (state.callableSuits.length > 0) {
    lines.push(`callable suits: ${state.callableSuits.map((s) => SUIT_NAMES[s] ?? s).join(', ')}`);
  }

  if (state.phase === 5 || state.phase === 6) {
    lines.push(
      `round result: picker points=${state.roundPickerPoints} multiplier=x${state.roundMultiplier} ` +
        `pickerWon=${state.roundPickerWon ? 'yes' : 'no'}`,
    );
    if (state.buried.length > 0) lines.push(`buried: ${formatIndexedCards(state.buried)}`);
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
