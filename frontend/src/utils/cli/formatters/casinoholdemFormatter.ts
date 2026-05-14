import type { Card, CasinoHoldemResponse, MaskedCard } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { CasinoHoldemPhase } from '../../../types/phases';
import { formatCard, formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [CasinoHoldemPhase.BET]: 'BET',
  [CasinoHoldemPhase.FLOP]: 'FLOP',
  [CasinoHoldemPhase.END]: 'END',
};

/** Format the dealer's hidden hand (all masked) prior to showdown. */
function formatDealerHidden(dealerHand: (Card | MaskedCard)[]): string {
  return dealerHand.map((c) => (isMaskedCard(c) ? '??' : formatCard(c))).join(', ');
}

/** Format a Casino Hold'em game state as terminal text. */
export function formatCasinoholdemState(state: CasinoHoldemResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader("Casino Hold'em"));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.community.length > 0) {
    lines.push(`Board: ${formatCardList(state.community)}`);
  }

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
  }

  if (state.dealerHand.length > 0) {
    if (state.phase === CasinoHoldemPhase.END && state.callBet > 0) {
      lines.push(`Dealer: ${formatCardList(state.dealerHand as Card[])}`);
    } else {
      lines.push(`Dealer: ${formatDealerHidden(state.dealerHand)}`);
    }
  }
  lines.push('----------');

  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.bonusBet > 0) lines.push(`AA bonus: ${state.bonusBet}`);
  if (state.callBet > 0) lines.push(`call: ${state.callBet}`);

  if (state.phase === CasinoHoldemPhase.END) {
    if (state.callBet > 0) {
      lines.push(state.dealerQualify ? 'Dealer qualifies' : 'Dealer does not qualify');
    }
    lines.push(`payout: ante=${state.antePayout} call=${state.callPayout} bonus=${state.bonusPayout}`);
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
