import type { UltimateTexasHoldemResponse } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASES: Record<number, string> = { 1: 'Bet', 2: 'PreFlop', 3: 'Flop', 4: 'River', 5: 'End' };

/** Format an Ultimate Texas Hold'em response for the terminal. */
export function formatUltimateTexasHoldemState(state: UltimateTexasHoldemResponse): string {
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
  if (state.phase === 1) lines.push('action: bet <ante> [trips]');
  else if (state.phase === 2) lines.push('action: 4 / 3 / check');
  else if (state.phase === 3) lines.push('action: 2 / check');
  else if (state.phase === 4) lines.push('action: 1 / fold');
  else lines.push('action: reset');
  if (state.message) lines.push(state.message);
  if (state.phase === 5) lines.push(`result: ${state.result} net=${state.netChange}`);
  lines.push(formatSeparator());
  return lines.join('\n');
}
