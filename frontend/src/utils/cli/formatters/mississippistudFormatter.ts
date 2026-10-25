import type { MississippiStudResponse } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { MississippiStudPhase, MississippiStudResult } from '../../../types/phases';
import {
  anteMax,
  canPlay,
  isBettingStreet,
  MISSISSIPPI_STUD_ANTE_MIN,
  MISSISSIPPI_STUD_ANTE_STEP,
  MISSISSIPPI_STUD_PLAY_MULTIPLIERS,
} from '../../mississippiStudBet';
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
    const max = anteMax(state.chips);
    if (max < MISSISSIPPI_STUD_ANTE_MIN) lines.push('Ante cannot be placed with the current chips');
    else lines.push(`Ante range: min ${MISSISSIPPI_STUD_ANTE_MIN}, step ${MISSISSIPPI_STUD_ANTE_STEP}, max ${max}`);
    lines.push('Action: bet <amount>');
  } else if (isBettingStreet(state.phase)) {
    const multipliers = MISSISSIPPI_STUD_PLAY_MULTIPLIERS.filter((multiplier) =>
      canPlay(multiplier, state.chips, state.anteAmount),
    );
    lines.push(
      multipliers.length ? `Action: play ${multipliers.join(', ')}x or fold` : 'Action: fold (no affordable play)',
    );
  } else if (state.phase === MississippiStudPhase.END) lines.push('Action: reset');
  if (state.folded) lines.push('Status: folded');
  if (state.phase === MississippiStudPhase.END) {
    const result =
      state.result === MississippiStudResult.WIN
        ? 'win'
        : state.result === MississippiStudResult.DRAW
          ? 'push'
          : 'lose';
    lines.push(`result: ${result}  payout: ${state.totalPayout}  net: ${state.netChange}`);
    lines.push(`hand rank: ${state.handRank}  payout multiplier: ${state.payoutMultiplier}`);
  }
  if (state.message) lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
