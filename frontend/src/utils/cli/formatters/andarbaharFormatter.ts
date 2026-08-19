import type { AndarBaharResponse } from '../../../types/card';
import { ANDAR_BAHAR_SIDE_BANDS } from '../../../types/games/andarbahar';
import { AndarBaharColumn, AndarBaharPhase, AndarBaharSideBand } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [AndarBaharPhase.BET]: 'BET',
  [AndarBaharPhase.END]: 'END',
};

const COLUMN_NAMES: Record<number, string> = {
  [AndarBaharColumn.ANDAR]: 'Andar',
  [AndarBaharColumn.BAHAR]: 'Bahar',
};

/** Renders a side-bet band as its inclusive card-count range. */
function bandLabel(band: number): string {
  const found = ANDAR_BAHAR_SIDE_BANDS.find((b) => b.band === band);
  if (!found) return '?';
  return found.lo === found.hi ? `${found.lo}` : `${found.lo}-${found.hi}`;
}

/** Format an Andar Bahar game state as terminal text. */
export function formatAndarBaharState(state: AndarBaharResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Andar Bahar'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`joker: ${state.joker ? formatCard(state.joker) : '??'}`);
  // **先に配る列は賭ける前に見えている必要がある。** 配当が 0.9:1 に下がる側です。
  lines.push(`dealt first: ${COLUMN_NAMES[state.firstColumn] ?? '?'} (pays 0.9:1)`);

  if (state.betAmount > 0) {
    lines.push(`bet: ${state.betAmount} on ${COLUMN_NAMES[state.betTarget] ?? '?'}`);
  }
  if (state.sideBand !== AndarBaharSideBand.NONE) {
    lines.push(`side bet: ${state.sideAmount} on ${bandLabel(state.sideBand)} cards`);
  }

  if (state.andarCards.length > 0 || state.baharCards.length > 0) {
    lines.push(`Andar: ${state.andarCards.map(formatCard).join(' ')}`);
    lines.push(`Bahar: ${state.baharCards.map(formatCard).join(' ')}`);
    lines.push(`cards dealt: ${state.dealtCount}`);
  }

  if (state.phase === AndarBaharPhase.END) {
    lines.push(`winner: ${COLUMN_NAMES[state.winner] ?? '?'}  payout: ${state.payout}`);
    // **サイドベットは別の賭け** (#5770)。張った回だけ内訳を出す。
    if (state.sideBand !== AndarBaharSideBand.NONE) {
      lines.push(`  breakdown: main ${state.mainPayout} / side ${state.sidePayout}`);
    }
  }

  if (state.history.length > 0) {
    lines.push(`history: ${state.history.map((h) => COLUMN_NAMES[h] ?? '?').join(' ')}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
