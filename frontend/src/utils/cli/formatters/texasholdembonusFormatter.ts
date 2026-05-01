import type { Card, MaskedCard, TexasHoldemBonusResponse } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { TexasHoldemBonusPhase } from '../../../types/phases';
import { formatCard, formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [TexasHoldemBonusPhase.BET]: 'BET',
  [TexasHoldemBonusPhase.PRE_FLOP]: 'PRE-FLOP',
  [TexasHoldemBonusPhase.FLOP]: 'FLOP',
  [TexasHoldemBonusPhase.TURN]: 'TURN',
  [TexasHoldemBonusPhase.END]: 'END',
};

/** Format the dealer's hidden hand (all masked) for the pre-showdown phases. */
function formatDealerHidden(dealerHand: (Card | MaskedCard)[]): string {
  return dealerHand.map((c) => (isMaskedCard(c) ? '??' : formatCard(c))).join(', ');
}

/** Format a Texas Hold'em Bonus Poker game state as terminal text. */
export function formatTexasholdembonusState(state: TexasHoldemBonusResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader("Texas Hold'em Bonus Poker"));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.community.length > 0) {
    lines.push(`Board: ${formatCardList(state.community)}`);
  }

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
  }

  if (state.dealerHand.length > 0) {
    if (state.phase === TexasHoldemBonusPhase.END) {
      lines.push(`Dealer: ${formatCardList(state.dealerHand as Card[])}`);
    } else {
      lines.push(`Dealer: ${formatDealerHidden(state.dealerHand)}`);
    }
  }
  lines.push('----------');

  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.bonusBet > 0) lines.push(`bonus: ${state.bonusBet}`);
  if (state.totalPlayBet > 0) lines.push(`play bets: ${state.totalPlayBet}`);

  if (state.phase === TexasHoldemBonusPhase.END) {
    lines.push(`payout: ante=${state.antePayout} play=${state.playPayout} bonus=${state.bonusPayout}`);
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
