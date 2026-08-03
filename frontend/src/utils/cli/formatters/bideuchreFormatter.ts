import type { BidEuchreResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'ChooseTrump', 'Play', 'HandEnd', 'GameEnd'];

/** Declarations, in menu order. **There are two no-trump forms.** */
const TRUMP_NAMES = ['S', 'C', 'D', 'H', 'NT-high', 'NT-low'];

/** Format a Bid Euchre game state as terminal text. */
export function formatBidEuchreState(state: BidEuchreResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Bid Euchre'));
  lines.push(`hand: ${state.handNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`team0: ${state.scores[0]} | team1: ${state.scores[1]}  (game is ${state.gameTarget})`);
  if (state.highBid) {
    const trump = state.trumpChosen ? (TRUMP_NAMES[state.trump] ?? '?') : 'not yet named';
    lines.push(`contract: ${state.highBid.value} tricks, trump ${trump}`);
  }
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    // **キティが無く、伏せられているのは他家の手札だけ。**
    const hand =
      p.cards.length > 0 ? p.cards.map((c, j) => `[${j}]${formatCard(c)}`).join(' ') : `hidden (${p.cardCount})`;
    const dealer = p.isDealer ? ' [dealer]' : '';
    const declarer = p.isDeclarer ? ' [declarer]' : '';
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    lines.push(`${name}(T${p.team})${dealer}${declarer}: tricks ${p.tricksWon}${turn}`);
    lines.push(`  ${hand}`);
  });
  lines.push('----------');

  if (state.trick.length > 0) lines.push(`trick: ${state.trick.map(formatCard).join(' ')}`);

  if (state.phase === 0 && state.bidPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push(`(your bid — "b <${state.minBid}-${state.maxBid}>"; the DEALER alone may EQUAL the standing bid)`);
  } else if (state.phase === 1 && state.declarerIdx === 0 && !state.gameEndFlag) {
    // **ノートランプが 2 種類ある。**ローは序列が逆転する。
    lines.push(`trump: ${TRUMP_NAMES.map((n, i) => `${i}:${n}`).join(' / ')}`);
    lines.push('(name trump with "t <0-5>"; at NT-low the ranking REVERSES and the nine is highest)');
  } else if (state.phase === 2 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push(`playable: ${state.validPlays.join(' ') || '-'}`);
    lines.push('(your turn — play with "p <i>"; the left bower counts as a trump)');
  } else if (state.phase === 3 && state.lastResult) {
    const r = state.lastResult;
    lines.push(r.made ? `(contract made: bid ${r.bid})` : `(contract FAILED: bid ${r.bid})`);
    // **未達側は宣言額を失い、守備側は取ったトリックを得点する。**
    lines.push(`points: team0 ${r.points[0]} / team1 ${r.points[1]}`);
    lines.push(`tricks: team0 ${r.tricks[0]} / team1 ${r.tricks[1]}`);
    if (!r.made) {
      lines.push('(a set costs the BID, not the tricks taken; the defenders still score theirs)');
    }
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game over! Winning team: ${state.winnerTeam}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
