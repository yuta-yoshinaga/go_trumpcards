import type { BaccaratResponse } from '../../../types/card';
import { formatCardList, formatHeader, formatSeparator } from '../formatterBase';

const BET_TYPE_NAMES: Record<number, string> = { 0: 'Player', 1: 'Banker', 2: 'Tie' };
const RESULT_NAMES: Record<number, string> = { 0: 'Player Win', 1: 'Banker Win', 2: 'Tie' };

/** Format a Baccarat game state as terminal text. */
export function formatBaccaratState(state: BaccaratResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Baccarat'));
  lines.push(`chips: ${state.chips}  phase: ${state.phase === 1 ? 'BET' : 'END'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Player: ${formatCardList(state.playerHand)} (${state.playerHandValue})`);
    lines.push(`Banker: ${formatCardList(state.bankerHand)} (${state.bankerHandValue})`);
  }
  lines.push('----------');

  if (state.betAmount > 0) {
    lines.push(`bet: ${state.betAmount} on ${BET_TYPE_NAMES[state.betType] ?? 'Unknown'}`);
  }
  if (state.playerPairBet > 0) lines.push(`player pair bet: ${state.playerPairBet}`);
  if (state.bankerPairBet > 0) lines.push(`banker pair bet: ${state.bankerPairBet}`);

  if (state.phase === 2) {
    lines.push(`result: ${RESULT_NAMES[state.result] ?? 'Unknown'} | payout: ${state.payout}`);
    if (state.sideBetResults.length > 0) {
      for (const sb of state.sideBetResults) {
        lines.push(`  ${sb.resultName}: ${sb.payout > 0 ? `+${sb.payout}` : String(sb.payout)}`);
      }
    }
  }

  if (state.history.length > 0) {
    const hist = state.history.slice(-10).map((r) => (r === 0 ? 'P' : r === 1 ? 'B' : 'T'));
    lines.push(`history: ${hist.join(' ')}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
