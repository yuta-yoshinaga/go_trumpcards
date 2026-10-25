import type { MississippiStudResponse } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { MississippiStudPhase } from '../../../types/phases';
import { getMississippiStudAnteMax } from '../commands/mississippistudCommands';
import { formatCard, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [MississippiStudPhase.ANTE]: 'Ante',
  [MississippiStudPhase.THIRD_STREET]: 'ThirdStreet',
  [MississippiStudPhase.FOURTH_STREET]: 'FourthStreet',
  [MississippiStudPhase.FIFTH_STREET]: 'FifthStreet',
  [MississippiStudPhase.END]: 'End',
};

/** Format Mississippi Stud response state as terminal text. */
export function formatMississippiStudState(state: MississippiStudResponse): string {
  const lines = [formatHeader('Mississippi Stud'), `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`];
  lines.push(`chips: ${state.chips}  ante: ${state.anteAmount}  total bet: ${state.totalBet}`);
  if (state.playerHand.length > 0) {
    lines.push(`hand: ${formatIndexedCards(state.playerHand)}`);
  }
  if (state.communityCards.length > 0) {
    lines.push(
      `community: ${state.communityCards.map((card) => (isMaskedCard(card) ? '[hidden]' : formatCard(card))).join(' ')}`,
    );
  }
  if (state.phase === MississippiStudPhase.ANTE) {
    lines.push(`Ante range: min 10, step 10, max ${getMississippiStudAnteMax(state.chips)}`);
    lines.push('Action: bet <amount>');
  } else if (
    state.phase === MississippiStudPhase.THIRD_STREET ||
    state.phase === MississippiStudPhase.FOURTH_STREET ||
    state.phase === MississippiStudPhase.FIFTH_STREET
  ) {
    lines.push('Action: play <1-3> or fold');
  } else if (state.phase === MississippiStudPhase.END) lines.push('Action: reset');
  if (state.folded) lines.push('Status: folded');
  if (state.phase === MississippiStudPhase.END) {
    const result = state.result > 0 ? 'win' : state.result < 0 ? 'push' : 'lose';
    lines.push(`result: ${result}  payout: ${state.totalPayout}  net: ${state.netChange}`);
    lines.push(`hand rank: ${state.handRank}  payout multiplier: ${state.payoutMultiplier}`);
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
