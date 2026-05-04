import type { CasinoWarResponse } from '../../../types/card';
import { CasinoWarPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [CasinoWarPhase.BET]: 'BET',
  [CasinoWarPhase.INITIAL_DEALT]: 'INITIAL DEALT',
  [CasinoWarPhase.TIE_DECISION]: 'TIE DECISION',
  [CasinoWarPhase.WAR_DEALT]: 'WAR DEALT',
  [CasinoWarPhase.END]: 'END',
};

const RESULT_NAMES: Record<number, string> = { 1: 'WIN', 0: 'PUSH', [-1]: 'LOSE' };

/** Format a Casino War game state as terminal text. */
export function formatCasinowarState(state: CasinoWarResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Casino War'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  if (state.ante > 0) lines.push(`ante: ${state.ante}${state.warBet > 0 ? `  warBet: ${state.warBet}` : ''}`);
  if (state.playerCard || state.dealerCard) {
    lines.push('Initial:');
    if (state.playerCard) lines.push(`  player: ${formatCard(state.playerCard)}`);
    if (state.dealerCard) lines.push(`  dealer: ${formatCard(state.dealerCard)}`);
  }
  if (state.burnCards.length > 0) {
    lines.push(`Burn: ${state.burnCards.map(formatCard).join(', ')}`);
  }
  if (state.playerWarCard || state.dealerWarCard) {
    lines.push('War:');
    if (state.playerWarCard) lines.push(`  player: ${formatCard(state.playerWarCard)}`);
    if (state.dealerWarCard) lines.push(`  dealer: ${formatCard(state.dealerWarCard)}`);
  }
  if (state.phase === CasinoWarPhase.END) {
    lines.push(`result: ${RESULT_NAMES[state.result] ?? 'UNKNOWN'}  payout: ${state.totalPayout}`);
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
