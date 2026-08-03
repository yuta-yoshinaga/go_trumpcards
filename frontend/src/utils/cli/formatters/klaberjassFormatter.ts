import type { KlaberjassResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['BidTurnUp', 'BidFree', 'Schmeiss', 'Play', 'HandEnd', 'GameEnd'];

const SUIT_GLYPHS: Record<number, string> = { 1: 'S', 2: 'C', 3: 'H', 4: 'D' };

/** Format a Klaberjass game state as terminal text. */
export function formatKlaberjassState(state: KlaberjassResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Klaberjass'));
  lines.push(`deal: ${state.dealNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`trump: ${SUIT_GLYPHS[state.trumpSuit] ?? '-'}  target: ${state.targetScore}`);
  if (state.turnUpCard && state.trumpSuit === 0) {
    lines.push(`turn-up: ${formatCard(state.turnUpCard)}`);
  }
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const hand =
      p.cards.length > 0 ? p.cards.map((c, j) => `[${j}]${formatCard(c)}`).join(' ') : `hidden (${p.cardCount})`;
    const dealer = p.isDealer ? ' [dealer]' : '';
    const maker = p.isMaker ? ' [maker]' : '';
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    lines.push(`${name}${dealer}${maker}: game ${p.score} hand ${p.handPoints}${turn}`);
    lines.push(`  ${hand}`);
  });
  lines.push('----------');

  if (state.trick.length > 0) {
    lines.push(`trick: ${state.trick.map(formatCard).join(' ')}`);
  }

  if (state.phase === 3 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    // 追随・切札・上乗せが全部強制なので、出せる札を出さないと操作できない。
    lines.push(`playable: ${state.validPlays.join(' ') || '-'}`);
    lines.push('(your turn — play with "p <i>")');
  } else if (state.phase === 0 && state.bidPlayerIdx === 0) {
    lines.push('(your bid — "a" to take the turn-up suit, "ps" to pass, "sm" to throw it in)');
  } else if (state.phase === 1 && state.bidPlayerIdx === 0) {
    lines.push('(your bid — "c <1-4>" to name a suit, "ps" to pass)');
  } else if (state.phase === 2 && state.bidPlayerIdx === 0) {
    lines.push('(a schmeiss is on the table — "y" to agree, "no" to refuse and make them the maker)');
  } else if (state.phase === 4) {
    lines.push(state.bete ? '(bete — the maker handed the whole hand over)' : '(hand over — "n" to deal again)');
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(`Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
