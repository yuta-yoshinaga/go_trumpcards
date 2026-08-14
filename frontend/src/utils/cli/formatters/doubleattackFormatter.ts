import type { DoubleAttackResponse } from '../../../types/card';
import { DOUBLE_ATTACK_RESULT } from '../../../types/games/doubleattack';
import { DoubleAttackPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [DoubleAttackPhase.BET]: 'BET',
  [DoubleAttackPhase.ATTACK]: 'ATTACK',
  [DoubleAttackPhase.PLAY]: 'PLAY',
  [DoubleAttackPhase.RESULT]: 'RESULT',
};

const RESULT_NAMES: Record<number, string> = {
  [DOUBLE_ATTACK_RESULT.none]: 'undecided',
  [DOUBLE_ATTACK_RESULT.win]: 'win',
  [DOUBLE_ATTACK_RESULT.lose]: 'lose',
  [DOUBLE_ATTACK_RESULT.push]: 'push',
  [DOUBLE_ATTACK_RESULT.blackjack]: 'blackjack (1:1)',
};

/** Format an Extra Bet Blackjack game state as terminal text. */
export function formatDoubleAttackState(state: DoubleAttackResponse): string {
  const lines: string[] = [formatHeader('Extra Bet Blackjack')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Hand: ${state.roundNumber} (chips: ${state.chips})`);

  if (state.anteBet > 0) {
    lines.push(`Ante ${state.anteBet} / Extra bet ${state.attackBet} / Bust It ${state.bustItBet}`);
  }

  if (state.dealerCards.length > 0) {
    lines.push(formatSeparator());
    if (!state.dealerHoleDealt) {
      // **アップカードだけ。** 2 枚目はまだ存在しないので点数も出さない。
      lines.push(`Dealer: ${formatCard(state.dealerCards[0])} (second card comes after the extra bet)`);
    } else {
      lines.push(`Dealer: ${state.dealerCards.map(formatCard).join(' ')} = ${state.dealerScore}`);
    }
  }

  state.hands.forEach((h, i) => {
    const mark = i === state.activeHand && state.phase === DoubleAttackPhase.PLAY ? '*' : ' ';
    const flags = [h.busted && 'bust', h.doubled && 'doubled', h.blackjack && 'blackjack'].filter(Boolean).join(' ');
    lines.push(
      `${mark}[${i + 1}] ${h.cards.map(formatCard).join(' ')} = ${h.score} (bet ${h.bet})${flags ? ` ${flags}` : ''}`,
    );
  });

  if (state.phase === DoubleAttackPhase.ATTACK) {
    lines.push(`Extra bet limit: ${state.maxAttackBet}`);
  }

  if (state.phase === DoubleAttackPhase.RESULT) {
    lines.push(formatSeparator());
    state.hands.forEach((h, i) => {
      lines.push(`Hand ${i + 1}: ${RESULT_NAMES[h.result] ?? '?'}`);
    });
    if (state.bustItPayout > 0) lines.push(`Bust It pays: ${state.bustItPayout}`);
  }
  if (state.gameEndFlag) lines.push('Out of chips.');

  return lines.join('\n');
}
