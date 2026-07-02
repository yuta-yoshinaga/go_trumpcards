import type { TrenteEtQuaranteResponse } from '../../../types/card';
import { TrenteEtQuarantePhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [TrenteEtQuarantePhase.BET]: 'BET',
  [TrenteEtQuarantePhase.RESULT]: 'RESULT',
};

const BET_NAMES: Record<number, string> = { 0: 'NOIR', 1: 'ROUGE', 2: 'COULEUR', 3: 'INVERSE' };

const RESULT_NAMES: Record<number, string> = { 1: 'WIN', 0: 'PUSH', [-1]: 'LOSE' };

/** Format a Trente et Quarante (Rouge et Noir) game state as terminal text. */
export function formatTrenteEtQuaranteState(state: TrenteEtQuaranteResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Trente et Quarante'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  if (state.stake > 0) lines.push(`stake: ${state.stake}  bet: ${BET_NAMES[state.currentBet] ?? 'UNKNOWN'}`);
  if (state.noirRow.length > 0) {
    lines.push(`Noir (${state.noirTotal}): ${state.noirRow.map(formatCard).join(', ')}`);
  }
  if (state.rougeRow.length > 0) {
    lines.push(`Rouge (${state.rougeTotal}): ${state.rougeRow.map(formatCard).join(', ')}`);
  }
  if (state.phase === TrenteEtQuarantePhase.RESULT) {
    if (state.refait) {
      lines.push('REFAIT (tie at 31 — half stake lost)');
    } else if (state.winningRow >= 0) {
      lines.push(`winner: ${state.winningRow === 0 ? 'NOIR' : 'ROUGE'}`);
    }
    lines.push(`result: ${RESULT_NAMES[state.result] ?? 'UNKNOWN'}  payout: ${state.payout}`);
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
