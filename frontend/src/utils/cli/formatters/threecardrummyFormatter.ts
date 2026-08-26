import type { ThreeCardRummyResponse } from '../../../types/card';
import { type Card, isMaskedCard } from '../../../types/common';
import { ThreeCardRummyPhase } from '../../../types/phases';
import { formatCardList, formatHeader, formatIndexedCards, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 1: 'BET', 2: 'ACTION', 3: 'END' };

/** Renders a total, calling out that 0 is a meld rather than an empty hand. */
function scoreText(score: number): string {
  return score === 0 ? '0 (meld)' : String(score);
}

/** Format a Three Card Rummy game state as terminal text. */
export function formatThreecardrummyState(state: ThreeCardRummyResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Three Card Rummy'));
  lines.push(`chips: ${state.chips}  phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push('');

  if (state.playerHand.length > 0) {
    lines.push(`Your hand: ${formatIndexedCards(state.playerHand)}`);
    // **点数がそのまま判断材料。** 低いほど強いので、数字を出さないと
    // play/fold を決めようがない。
    lines.push(`Your score: ${scoreText(state.playerScore)} (lower is better)`);
  }

  if (state.phase === ThreeCardRummyPhase.END && state.dealerHand.length > 0) {
    // End フェーズではサーバが全枚数を開くので、マスク済みの札は残らない。
    // 万一残っても `??` に落として、伏せ札を実在のカードとして書かない。
    const revealed = state.dealerHand.filter((c): c is Card => !isMaskedCard(c));
    const masked = state.dealerHand.length - revealed.length;
    lines.push(`Dealer: ${[formatCardList(revealed), ...Array(masked).fill('??')].join(' ')}`.trimEnd());
    lines.push(`Dealer score: ${scoreText(state.dealerScore)}`);
    lines.push(`Dealer qualified: ${state.dealerQualified ? 'yes' : 'no'}`);
  }
  lines.push('----------');

  if (state.anteBet > 0) lines.push(`ante: ${state.anteBet}`);
  if (state.lowBonusBet > 0) lines.push(`low bonus: ${state.lowBonusBet}`);
  if (state.playBet > 0) lines.push(`play bet: ${state.playBet}`);

  if (state.phase === ThreeCardRummyPhase.END) {
    lines.push(
      `payout: ante=${state.antePayout} play=${state.playPayout} anteBonus=${state.anteBonusPayout} lowBonus=${state.lowBonusPayout}`,
    );
    lines.push(`total: ${state.totalPayout}`);
  }

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
