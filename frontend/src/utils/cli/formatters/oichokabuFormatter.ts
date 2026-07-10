import type { OichoKabuResponse } from '../../../types/card';
import { OichoKabuPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [OichoKabuPhase.BET]: 'BET',
  [OichoKabuPhase.DRAW]: 'DRAW',
  [OichoKabuPhase.END]: 'END',
};

const RESULT_NAMES: Record<number, string> = { 1: 'WIN', 0: 'PUSH', [-1]: 'LOSE' };

/** Format an Oicho-Kabu game state as terminal text. */
export function formatOichokabuState(state: OichoKabuResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Oicho-Kabu'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  if (state.bet > 0) lines.push(`bet: ${state.bet}`);
  if (state.playerHand.length > 0) {
    lines.push(`Child:  ${state.playerHand.map(formatCard).join(', ')}  (kabu ${state.playerRank})`);
  }
  if (state.bankerHand.length > 0) {
    lines.push(`Banker: ${state.bankerHand.map(formatCard).join(', ')}  (kabu ${state.bankerRank})`);
  }
  if (state.phase === OichoKabuPhase.END) {
    lines.push(`result: ${RESULT_NAMES[state.result] ?? 'UNKNOWN'}  payout: ${state.totalPayout}`);
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
