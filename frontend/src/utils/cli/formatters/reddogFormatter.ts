import type { RedDogResponse } from '../../../types/card';
import { RedDogPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [RedDogPhase.BET]: 'BET',
  [RedDogPhase.INITIAL_DEALT]: 'INITIAL DEALT',
  [RedDogPhase.SPREAD_DECISION]: 'SPREAD DECISION',
  [RedDogPhase.PAIR_THIRD]: 'PAIR THIRD',
  [RedDogPhase.END]: 'END',
};

const RESULT_NAMES: Record<number, string> = { 1: 'WIN', 0: 'PUSH', [-1]: 'LOSE' };

/** Format a Red Dog game state as terminal text. */
export function formatReddogState(state: RedDogResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Red Dog'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  if (state.ante > 0) lines.push(`ante: ${state.ante}${state.raise > 0 ? `  raise: ${state.raise}` : ''}`);
  if (state.phase === RedDogPhase.SPREAD_DECISION || state.phase === RedDogPhase.END)
    lines.push(`spread: ${state.spread}`);
  if (state.initialCards.length > 0) {
    lines.push(`Initial: ${state.initialCards.map(formatCard).join(', ')}`);
  }
  if (state.thirdCard) {
    lines.push(`Third: ${formatCard(state.thirdCard)}`);
  }
  if (state.phase === RedDogPhase.END) {
    lines.push(`result: ${RESULT_NAMES[state.result] ?? 'UNKNOWN'}  payout: ${state.totalPayout}`);
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
