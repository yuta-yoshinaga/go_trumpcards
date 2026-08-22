import type { Card, CaribbeanDrawResponse, MaskedCard } from '../../../types/card';
import { isMaskedCard } from '../../../types/card';
import { CaribbeanDrawPhase } from '../../../types/phases';
import { formatCard, formatCardList, formatHeader, formatSeparator } from '../formatterBase';

/** Phase labels (sync: internal/domain/CaribbeanDraw.go — DRAW sits between BET and ACTION). */
const PHASE_NAMES: Record<number, string> = {
  [CaribbeanDrawPhase.BET]: 'BET',
  [CaribbeanDrawPhase.DRAW]: 'DRAW',
  [CaribbeanDrawPhase.ACTION]: 'ACTION',
  [CaribbeanDrawPhase.END]: 'END',
};

/**
 * Number the player's cards from 1, matching what `d <n...>` expects.
 *
 * The shared `formatIndexedCards` labels from 0; printing those numbers next to
 * a parser that subtracts one would discard the card *before* the one the
 * player pointed at, without any error to show for it.
 */
function formatNumberedCards(cards: Card[]): string {
  return cards.map((c, i) => `[${i + 1}]${formatCard(c)}`).join('  ');
}

/** Format the dealer's partial hand before the showdown (1 face-up + hidden). */
function formatDealerActionHand(dealerHand: (Card | MaskedCard)[]): string {
  return dealerHand.map((c) => (isMaskedCard(c) ? '??' : formatCard(c))).join(', ');
}

/** Format a Caribbean Draw Poker game state as terminal text. */
export function formatCaribbeandrawState(state: CaribbeanDrawResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Caribbean Draw Poker'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatNumberedCards(state.playerHand)}`);
  }

  if (state.phase === CaribbeanDrawPhase.DRAW) {
    // The fee is charged on confirm, so it has to be on screen before the
    // player types `d` — not only in the payout block once it is too late.
    lines.push(`Exchange up to 2 cards for ${state.anteBet} (d 1 3), or stand pat (d).`);
  }

  if (
    (state.phase === CaribbeanDrawPhase.DRAW || state.phase === CaribbeanDrawPhase.ACTION) &&
    state.dealerHand.length > 0
  ) {
    lines.push(`Dealer: ${formatDealerActionHand(state.dealerHand)}`);
  }

  if (state.phase === CaribbeanDrawPhase.END && state.dealerHand.length > 0) {
    // At end phase the backend reveals all dealer cards, so the slice is guaranteed Card[].
    lines.push(`Dealer: ${formatCardList(state.dealerHand as Card[])}`);
    lines.push(`Dealer qualified: ${state.dealerQualified ? 'yes' : 'no'}`);
  }
  lines.push('----------');

  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.jackpotBet > 0) lines.push(`jackpot: ${state.jackpotBet}`);
  if (state.playBet > 0) lines.push(`play bet: ${state.playBet}`);
  // A cost, not a payout: it is missing from totalPayout by design, which is
  // exactly why it has to be printed on its own line.
  if (state.drawCost > 0) lines.push(`draw fee: ${state.drawCost}`);

  if (state.phase === CaribbeanDrawPhase.END) {
    lines.push(`payout: ante=${state.antePayout} play=${state.playPayout} jackpot=${state.jackpotPayout}`);
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
