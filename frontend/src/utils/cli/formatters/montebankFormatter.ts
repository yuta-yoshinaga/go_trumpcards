import type { MonteBankResponse } from '../../../types/card';
import { MONTE_BANK_RESULT } from '../../../types/games/montebank';
import { MonteBankPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [MonteBankPhase.BET]: 'BET',
  [MonteBankPhase.RESULT]: 'RESULT',
  [MonteBankPhase.GAME_END]: 'GAME END',
};

const RESULT_NAMES: Record<number, string> = {
  [MONTE_BANK_RESULT.none]: 'undecided',
  [MONTE_BANK_RESULT.win]: 'match',
  [MONTE_BANK_RESULT.lose]: 'no match',
};

/** Format a Monte Bank game state as terminal text. */
export function formatMonteBankState(state: MonteBankResponse): string {
  const lines: string[] = [formatHeader('Monte Bank')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Round: ${state.roundNumber} (chips: ${state.chips} / cards left: ${state.remainingCards})`);

  lines.push(formatSeparator());
  state.layout.forEach((entry, i) => {
    const mark = entry.isPicked ? '*' : ' ';
    // **各札に「互角か」を必ず添える。** それが賭けの良し悪しを決める唯一の数字。
    const note = entry.isEven ? 'only one of this suit (even)' : `${entry.suitCount} of this suit (against you)`;
    lines.push(`${mark}[${i + 1}] ${formatCard(entry.card)}  <- ${note}`);
  });

  if (state.gate) {
    lines.push(`Gate: ${formatCard(state.gate)}`);
  }
  if (state.result !== MONTE_BANK_RESULT.none) {
    lines.push(`${RESULT_NAMES[state.result] ?? '?'} (net ${state.payout - state.bet})`);
  }
  if (state.gameEndFlag) lines.push(`Finished with ${state.chips} chips.`);

  return lines.join('\n');
}
