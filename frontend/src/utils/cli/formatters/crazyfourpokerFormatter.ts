import type { CrazyFourPokerResponse } from '../../../types/card';
import { CRAZY_FOUR_POKER_RESULT } from '../../../types/games/crazyfourpoker';
import { CrazyFourPokerPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [CrazyFourPokerPhase.BET]: 'BET',
  [CrazyFourPokerPhase.DECIDE]: 'DECIDE',
  [CrazyFourPokerPhase.RESULT]: 'RESULT',
};

const RESULT_NAMES: Record<number, string> = {
  [CRAZY_FOUR_POKER_RESULT.none]: 'undecided',
  [CRAZY_FOUR_POKER_RESULT.fold]: 'folded',
  [CRAZY_FOUR_POKER_RESULT.win]: 'you win',
  [CRAZY_FOUR_POKER_RESULT.lose]: 'dealer wins',
  [CRAZY_FOUR_POKER_RESULT.push]: 'a push',
  [CRAZY_FOUR_POKER_RESULT.dealerNotQualified]: 'dealer does not qualify',
};

const RANK_NAMES = [
  '',
  'High Card',
  'Pair',
  'Two Pair',
  'Straight',
  'Flush',
  'Three of a Kind',
  'Straight Flush',
  'Four of a Kind',
];

/** Renders a rank number, or a dash before any hand exists. */
function rankName(rank: number): string {
  return RANK_NAMES[rank] ?? '?';
}

/** Format a Crazy 4 Poker game state as terminal text. */
export function formatCrazyFourPokerState(state: CrazyFourPokerResponse): string {
  const lines: string[] = [formatHeader('Crazy 4 Poker')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Hand: ${state.roundNumber} (chips: ${state.chips})`);

  if (state.anteBet > 0) {
    lines.push(`Ante ${state.anteBet} / Super Bonus ${state.superBet} / Queens Up ${state.queensUpBet}`);
  }

  if (state.playerHand.length > 0) {
    lines.push(formatSeparator());
    lines.push(`You: ${state.playerHand.map(formatCard).join(' ')}`);
    lines.push(`  best four: ${state.playerBest.map(formatCard).join(' ')} (${rankName(state.playerHandRank)})`);
    if (state.phase === CrazyFourPokerPhase.DECIDE) {
      lines.push(`  multipliers available: 1-${state.maxMultiplier}`);
    }
  }

  // **決着するまでディーラーの手は出さない。** サーバも送っていない。
  if (state.phase === CrazyFourPokerPhase.RESULT && state.dealerHand.length > 0) {
    lines.push(`Dealer: ${state.dealerHand.map(formatCard).join(' ')}`);
    lines.push(`  best four: ${state.dealerBest.map(formatCard).join(' ')} (${rankName(state.dealerHandRank)})`);
    if (!state.dealerQualifies) {
      lines.push('  (does not qualify: ante pays 1:1, play pushes)');
    }
  }

  if (state.result !== CRAZY_FOUR_POKER_RESULT.none) {
    const staked = state.anteBet + state.superBet + state.queensUpBet + state.playBet;
    lines.push(formatSeparator());
    lines.push(`Result: ${RESULT_NAMES[state.result] ?? '?'} (net ${state.payout - staked})`);
  }
  if (state.gameEndFlag) lines.push('Out of chips.');

  return lines.join('\n');
}
