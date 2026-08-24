import type { SchafkopfResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

const PHASE_NAMES = ['Pick', 'Call', 'Play', 'TrickEnd', 'RoundEnd', 'GameEnd'];
const SUIT_NAMES = ['none', '♠', '♣', '♥', '♦'];

/**
 * Render the contract in play. Solo carries its trump suit, because "Solo"
 * on its own does not say which colour is trump.
 */
/** The command that declares each contract, for the "you may declare" line. */
const CONTRACT_COMMANDS: Readonly<Record<number, string>> = {
  0: 'pick',
  1: 'wenz',
  2: 'solo <suit>',
};

function contractLabel(state: SchafkopfResponse): string {
  switch (state.contract) {
    case 1:
      return 'Wenz (Unters only)';
    case 2:
      return `Solo (${SUIT_NAMES[state.soloSuit] ?? '-'})`;
    default:
      return 'Rufspiel (called ace)';
  }
}

/** Format a Schafkopf game state as terminal text. */
export function formatSchafkopfState(state: SchafkopfResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Schafkopf'));
  lines.push(
    `round: ${state.roundNumber}  trick: ${state.trickNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
  );

  const pickerLabel = state.pickerIdx >= 0 ? formatPlayerName(state.pickerIdx, false) : '-';
  const partnerLabel = state.partnerRevealed && state.partnerIdx >= 0 ? formatPlayerName(state.partnerIdx, false) : '?';
  lines.push(`picker: ${pickerLabel}  partner: ${partnerLabel}  calledSuit: ${SUIT_NAMES[state.calledSuit] ?? '-'}`);
  lines.push(`contract: ${contractLabel(state)}`);
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

  // 上回れる契約だけを案内する。打てば必ず拒否されるコマンドは勧めない。
  if (state.phase === 0 && (state.beatableContracts ?? []).length > 0) {
    const names = state.beatableContracts.map((c) => CONTRACT_COMMANDS[c] ?? String(c));
    lines.push(`you may declare: ${names.join(', ')}`);
  }

  if (state.callableSuits.length > 0) {
    lines.push(`callable suits: ${state.callableSuits.map((s) => SUIT_NAMES[s] ?? s).join(', ')}`);
  }

  if (state.phase === 4 || state.phase === 5) {
    lines.push(
      `round result: picker points=${state.roundPickerPoints} multiplier=x${state.roundMultiplier} ` +
        `pickerWon=${state.roundPickerWon ? 'yes' : 'no'}`,
    );
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
