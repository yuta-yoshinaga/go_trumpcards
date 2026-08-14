import type { FreeBetResponse } from '../../../types/card';
import { FREE_BET_RESULT } from '../../../types/games/freebet';
import { FreeBetPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [FreeBetPhase.BET]: 'BET',
  [FreeBetPhase.PLAY]: 'PLAY',
  [FreeBetPhase.RESULT]: 'RESULT',
};

const RESULT_NAMES: Record<number, string> = {
  [FREE_BET_RESULT.none]: 'undecided',
  [FREE_BET_RESULT.win]: 'win',
  [FREE_BET_RESULT.lose]: 'lose',
  [FREE_BET_RESULT.push]: 'push',
  [FREE_BET_RESULT.blackjack]: 'blackjack (3:2)',
  [FREE_BET_RESULT.dealer22Push]: 'push (dealer 22)',
};

/** Format a Free Bet Blackjack game state as terminal text. */
export function formatFreeBetState(state: FreeBetResponse): string {
  const lines: string[] = [formatHeader('Free Bet Blackjack')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Hand: ${state.roundNumber} (chips: ${state.chips})`);

  if (state.anteBet > 0) lines.push(`Ante ${state.anteBet}`);

  if (state.dealerCards.length > 0) {
    lines.push(formatSeparator());
    lines.push(`Dealer: ${state.dealerCards.map(formatCard).join(' ')} = ${state.dealerScore}`);
  }

  state.hands.forEach((h, i) => {
    const mark = i === state.activeHand && state.phase === FreeBetPhase.PLAY ? '*' : ' ';
    const flags = [h.busted && 'bust', h.doubled && 'doubled', h.blackjack && 'blackjack'].filter(Boolean).join(' ');
    // **ハウス持ちのぶんは別建てで書く。** 合算すると自分がいくら失うのかが読めない。
    const house = h.freeBet > 0 ? ` + house ${h.freeBet}` : '';
    lines.push(
      `${mark}[${i + 1}] ${h.cards.map(formatCard).join(' ')} = ${h.score} (bet ${h.bet}${house})${flags ? ` ${flags}` : ''}`,
    );
  });

  if (state.phase === FreeBetPhase.PLAY) {
    const free = [state.canFreeDouble && 'freedouble', state.canFreeSplit && 'freesplit'].filter(Boolean);
    if (free.length > 0) lines.push(`Free now: ${free.join(' / ')}`);
  }

  if (state.phase === FreeBetPhase.RESULT) {
    lines.push(formatSeparator());
    if (state.dealerPushed22) lines.push('Dealer busted with 22 — surviving hands push.');
    state.hands.forEach((h, i) => {
      lines.push(`Hand ${i + 1}: ${RESULT_NAMES[h.result] ?? '?'}`);
    });
  }
  if (state.gameEndFlag) lines.push('Out of chips.');

  return lines.join('\n');
}
