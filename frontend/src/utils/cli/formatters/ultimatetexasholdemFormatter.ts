import type { UltimateTexasHoldemResponse } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { UltimateTexasHoldemPhase } from '../../../types/phases';
import {
  anteMax,
  canPlay,
  UTH_ANTE_MIN,
  UTH_ANTE_STEP,
  UTH_HAND_RANK_NAMES,
  UTH_TRIPS_MIN,
  UTH_TRIPS_STEP,
} from '../../utHoldemBet';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASES: Record<number, string> = {
  [UltimateTexasHoldemPhase.BET]: 'Bet',
  [UltimateTexasHoldemPhase.PRE_FLOP]: 'PreFlop',
  [UltimateTexasHoldemPhase.FLOP]: 'Flop',
  [UltimateTexasHoldemPhase.RIVER]: 'River',
  [UltimateTexasHoldemPhase.END]: 'End',
};

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
  if (state.phase === UltimateTexasHoldemPhase.BET) {
    const maxAnte = anteMax(state.chips);
    lines.push(
      `bet range: ante ${UTH_ANTE_MIN}–${maxAnte} (step ${UTH_ANTE_STEP}); trips ${UTH_TRIPS_MIN}–(chips − 2×ante) (step ${UTH_TRIPS_STEP}; depends on the ante you enter)`,
    );
    lines.push('action: bet <ante> [trips]');
  } else if (state.phase === UltimateTexasHoldemPhase.PRE_FLOP) {
    const available = [
      canPlay(4, state.chips, state.anteBet) ? '4' : '',
      canPlay(3, state.chips, state.anteBet) ? '3' : '',
      'check',
    ].filter(Boolean);
    lines.push(`action: ${available.join(' / ')}`);
  } else if (state.phase === UltimateTexasHoldemPhase.FLOP) {
    lines.push(`action: ${canPlay(2, state.chips, state.anteBet) ? '2 / check' : 'check'}`);
  } else if (state.phase === UltimateTexasHoldemPhase.RIVER) {
    lines.push(`action: ${canPlay(1, state.chips, state.anteBet) ? '1 / fold' : 'fold'}`);
  } else if (state.phase === UltimateTexasHoldemPhase.END) lines.push('action: reset');
  if (state.message) lines.push(state.message);
  if (state.phase === UltimateTexasHoldemPhase.END) {
    const result = state.result === 1 ? 'win' : state.result === 0 ? 'push' : 'lose';
    lines.push(`result: ${result} net=${state.netChange}`);
    lines.push(
      `payout: ante=${state.antePayout} blind=${state.blindPayout} play=${state.playPayout} trips=${state.tripsPayout} total=${state.totalPayout}`,
    );
    lines.push(
      `hand: player=${UTH_HAND_RANK_NAMES[state.playerHandRank] ?? 'Unknown'} dealer=${UTH_HAND_RANK_NAMES[state.dealerHandRank] ?? 'Unknown'}`,
    );
  }
  lines.push(formatSeparator());
  return lines.join('\n');
}
