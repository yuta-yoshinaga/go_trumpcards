import type { UltimateTexasHoldemResponse } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { utHoldemBetBounds } from '../../utHoldemBet';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASES: Record<number, string> = { 1: 'Bet', 2: 'PreFlop', 3: 'Flop', 4: 'River', 5: 'End' };

/** Format an Ultimate Texas Hold'em response for the terminal. */
export function formatUltimateTexasHoldemState(state: UltimateTexasHoldemResponse, anteAmount = 10): string {
  const lines = [
    formatHeader('Ultimate Texas Holdem'),
    `phase: ${PHASES[state.phase] ?? state.phase}`,
    `chips: ${state.chips}`,
  ];
  lines.push(`bets: ante=${state.anteBet} blind=${state.blindBet} trips=${state.tripsBet} play=${state.playBet}`);
  if (state.playerHand.length)
    lines.push(`player: ${state.playerHand.map((card, idx) => `[${idx}]${formatCard(card)}`).join(' ')}`);
  if (state.dealerHand.length)
    lines.push(`dealer: ${state.dealerHand.map((card) => (isMaskedCard(card) ? '??' : formatCard(card))).join(', ')}`);
  if (state.community.length) lines.push(`community: ${state.community.map(formatCard).join(', ')}`);
  if (state.phase === 1) {
    const maxAnte = Math.floor(state.chips / 2);
    lines.push(
      `bet range: ante 10–${maxAnte} (step 10); trips 0–${utHoldemBetBounds(anteAmount, 0, state.chips).maxTrips} (step 10)`,
    );
    lines.push('action: bet <ante> [trips]');
  } else if (state.phase === 2) {
    const available = [
      state.chips >= state.anteBet * 4 ? '4' : '',
      state.chips >= state.anteBet * 3 ? '3' : '',
      'check',
    ].filter(Boolean);
    lines.push(`action: ${available.join(' / ')}`);
  } else if (state.phase === 3) {
    lines.push(`action: ${state.chips >= state.anteBet * 2 ? '2 / check' : 'check'}`);
  } else if (state.phase === 4) {
    lines.push(`action: ${state.chips >= state.anteBet ? '1 / fold' : 'fold'}`);
  } else if (state.phase === 5) lines.push('action: reset');
  if (state.message) lines.push(state.message);
  if (state.phase === 5) {
    const result = state.result === 1 ? 'win' : state.result === 0 ? 'push' : 'lose';
    lines.push(`result: ${result} net=${state.netChange}`);
    lines.push(
      `payout: ante=${state.antePayout} blind=${state.blindPayout} play=${state.playPayout} trips=${state.tripsPayout} total=${state.totalPayout}`,
    );
    lines.push(`hand: player=${state.playerHandRank} dealer=${state.dealerHandRank}`);
  }
  lines.push(formatSeparator());
  return lines.join('\n');
}
