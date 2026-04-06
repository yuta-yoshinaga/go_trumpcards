import type { BlackJackResponse } from '../../../types/card';
import { BjPhase } from '../../../types/phases';
import { formatCard, formatCardList, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [BjPhase.BET]: 'BET',
  [BjPhase.DEAL]: 'DEAL',
  [BjPhase.INSURANCE]: 'INSURANCE',
  [BjPhase.ACTION]: 'ACTION',
  [BjPhase.END]: 'END',
  [BjPhase.EARLY_SURRENDER]: 'EARLY SURRENDER',
};

/** Format a BlackJack game state as terminal text. */
export function formatBlackjackState(state: BlackJackResponse): string {
  const lines: string[] = [];
  const sep = '----------';

  // Header info
  lines.push(sep);
  lines.push(`chips: player=${state.player.chips} dealer=${state.dealer.chips} decks=${state.deckCount}`);
  if (state.dealerHitsSoft17) lines.push('rule: H17 (Dealer hits soft 17)');
  if (state.countingEnabled) {
    lines.push(`count: RC=${state.runningCount} TC=${state.trueCount.toFixed(1)}`);
  }
  if (state.multiHandCount > 1) lines.push(`multi-hand: ${state.multiHandCount} hands`);
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  // Dealer
  const dealerCards = state.dealer.cards ?? [];
  if (state.phase === BjPhase.END || state.phase === BjPhase.BET) {
    const score = state.dealer.score ?? 0;
    lines.push(`dealer score ${score > 0 ? score : ''}`);
    lines.push(`  ${dealerCards.length > 0 ? formatCardList(dealerCards) : '(no cards)'}`);
  } else {
    lines.push('dealer score [?]');
    if (dealerCards.length > 0) {
      lines.push(`  ${formatCard(dealerCards[0])}, [?]`);
    }
  }
  lines.push(sep);

  // Player hands
  const hands = state.hands ?? [];
  for (let i = 0; i < hands.length; i++) {
    const hand = hands[i];
    const tags: string[] = [];
    if (hand.doubled) tags.push('[DD]');
    if (hand.busted) tags.push('[BUST]');
    if (hand.stood) tags.push('[STAND]');
    if (hand.isBlackJack) tags.push('[BJ]');
    if (hand.surrendered) tags.push('[SURRENDER]');
    const current = i === state.currentHandIdx ? '>' : ' ';
    lines.push(
      `${current}hand${hands.length > 1 ? ` ${i}` : ''} score ${hand.score} bet=${hand.bet} ${tags.join(' ')}`.trimEnd(),
    );
    lines.push(`  ${formatIndexedCards(hand.cards)}`);
  }
  lines.push(sep);

  // Insurance
  if (state.insuranceBet > 0) lines.push(`insurance bet: ${state.insuranceBet}`);
  if (state.insuranceAvailable) lines.push('Insurance available!');

  // Hint
  if (state.hintEnabled && state.suggestedAction > 0) {
    const hints: Record<number, string> = {
      1: 'Hit',
      2: 'Stand',
      3: 'Double',
      4: 'Split',
      5: 'Surrender',
      6: 'Decline Insurance',
      7: 'Double/Stand',
    };
    lines.push(`HINT: ${hints[state.suggestedAction] ?? 'Unknown'}`);
  }

  // Message
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
