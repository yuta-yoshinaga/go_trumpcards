import i18n from '../../../i18n';
import type { MaoResponse } from '../../../types/card';
import { ruleHintText } from '../../maoRuleHint';
import { formatCard, formatHeader, formatIndexedCards, formatPlayerName, formatSeparator } from '../formatterBase';

const SUIT_NAMES: Record<number, string> = { 1: 'Spade', 2: 'Clover', 3: 'Heart', 4: 'Diamond' };

/** Number of correct compliances required before a rule hint is unlocked. */
const HINT_THRESHOLD = 3;

/** Format a Mao game state as terminal text. */
export function formatMaoState(state: MaoResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Mao'));
  lines.push(
    `round: ${state.roundNumber}  draw pile: ${state.drawPileCount}  direction: ${state.direction < 0 ? '<-' : '->'}`,
  );
  if (state.discardTop) lines.push(`discard: ${formatCard(state.discardTop)}`);
  if (state.chosenSuit > 0) lines.push(`chosen suit: ${SUIT_NAMES[state.chosenSuit] ?? '?'}`);
  if (state.penaltyDrawCount > 0) lines.push(`draw penalty: ${state.penaltyDrawCount}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  // Hidden-rule signals: never reveal the rule, only the indirect hints.
  lines.push(`compliance: ${state.correctCount}/${HINT_THRESHOLD}`);
  if (state.rulePenalty) lines.push('!! a hidden-rule penalty was applied');
  if (state.awaitingWord) lines.push('?? you may need to say a word (dw <word>)');
  // CLI パネルも同じ応答を読むので、ここも訳す。素の ruleHint はサーバの
  // 言語で届く (#4917)。
  if (state.hintUnlocked && state.ruleHint) lines.push(`hint: ${ruleHintText(state, (key) => i18n.t(`mao:${key}`))}`);

  if (state.phase === 1) lines.push('Choose a suit (suit <spade|clover|heart|diamond>)');
  if (state.phase === 2) lines.push('Declare "Mao!" (dc) or skip (sk)');

  if (!state.gameEndFlag) {
    const current = formatPlayerName(state.currentPlayerIdx, state.players[state.currentPlayerIdx]?.isHuman ?? false);
    lines.push(`turn: ${current}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
