import type { ChinesePokerResponse } from '../../../types/card';
import { ChinesePokerPhase } from '../../../types/phases';
import { formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [ChinesePokerPhase.BET]: 'BET',
  [ChinesePokerPhase.SET_HANDS]: 'SET HANDS',
  [ChinesePokerPhase.END]: 'END',
};

/** Describe a single row's showdown result (from the player's perspective). */
function rowResult(res: number): string {
  if (res > 0) return 'WIN';
  if (res < 0) return 'LOSE';
  return 'PUSH';
}

/** Format a Chinese Poker game state as terminal text. */
export function formatChinesePokerState(state: ChinesePokerResponse | null): string {
  if (!state) return 'Loading...';
  const lines: string[] = [];

  lines.push(formatHeader('Chinese Poker'));
  lines.push(`chips: ${state.chips} | bet: ${state.bet} | phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push('----------');

  if (state.phase === ChinesePokerPhase.SET_HANDS) {
    // Show the 13-card hand with indices so the player can build `set` arguments.
    lines.push('your cards:');
    lines.push(`  ${formatIndexedCards(state.playerCards)}`);
    lines.push('Set with: s <f0 f1 f2 m0 m1 m2 m3 m4> (back = the remaining 5)');
  } else if (state.phase === ChinesePokerPhase.END) {
    lines.push(`you    front:  ${formatCardList(state.playerFront) || '-'}  [${rowResult(state.frontResult)}]`);
    lines.push(`       middle: ${formatCardList(state.playerMiddle) || '-'}  [${rowResult(state.middleResult)}]`);
    lines.push(`       back:   ${formatCardList(state.playerBack) || '-'}  [${rowResult(state.backResult)}]`);
    lines.push('----------');
    lines.push(`dealer front:  ${formatCardList(state.dealerFront) || '-'}`);
    lines.push(`       middle: ${formatCardList(state.dealerMiddle) || '-'}`);
    lines.push(`       back:   ${formatCardList(state.dealerBack) || '-'}`);
    lines.push('----------');
    if (state.scoop) lines.push('SCOOP! (won all three rows)');
    if (state.playerRoyalty > 0) lines.push(`royalty bonus: +${state.playerRoyalty}`);
    lines.push(`payout: ${state.payout}`);
  } else {
    lines.push('Place a bet: b <amount>');
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
