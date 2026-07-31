import type { VintResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Play', 'HandEnd', 'GameEnd'];

/** Denominations in bidding order — spades LOWEST, no trump highest. */
const DENOM_NAMES = ['S', 'C', 'D', 'H', 'NT'];

/** Format a Vint game state as terminal text. */
export function formatVintState(state: VintResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Vint'));
  lines.push(`hand: ${state.handNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(
    `team0: below ${state.below[0]} above ${state.above[0]} games ${state.gamesWon[0]} | ` +
      `team1: below ${state.below[1]} above ${state.above[1]} games ${state.gamesWon[1]}`,
  );
  if (state.highBid) {
    lines.push(
      `contract: ${state.highBid.level} ${DENOM_NAMES[state.highBid.denom] ?? '?'} (trick value ${state.highBid.trickValue})`,
    );
  }
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    // **ダミーが無いので、プレイ中は誰の手札も見えない。**
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
    // **♠ が最弱で NT が最強。**ブリッジと逆なので必ず見せる。
    const ladder = DENOM_NAMES.map((n, d) => `${d}:${n}(${state.trickValues[d]})`).join(' < ');
    lines.push(`ranking: ${ladder}`);
    lines.push('(your bid — "b <1-7> <0-4>"; spades are LOWEST, the reverse of bridge)');
  } else if (state.phase === 1 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push(`playable: ${state.validPlays.join(' ') || '-'}`);
    lines.push('(your turn — play with "p <i>")');
  } else if (state.phase === 2 && state.lastResult) {
    const r = state.lastResult;
    lines.push(
      r.made ? `(contract made with ${r.declarerTricks} tricks)` : `(contract FAILED with ${r.declarerTricks} tricks)`,
    );
    // **両チームが線下に得点する。**守備側の分も出す。
    lines.push(`below: team0 +${r.trickPoints[0]} / team1 +${r.trickPoints[1]} (BOTH sides score their tricks)`);
    if (r.honourPoints[0] > 0 || r.honourPoints[1] > 0) {
      lines.push(`honours: team0 +${r.honourPoints[0]} / team1 +${r.honourPoints[1]}`);
    }
    if (r.acePoints[0] > 0 || r.acePoints[1] > 0) {
      lines.push(`aces: team0 +${r.acePoints[0]} / team1 +${r.acePoints[1]}`);
    }
    if (r.penalty[0] > 0 || r.penalty[1] > 0) {
      lines.push(`penalty: team0 +${r.penalty[0]} / team1 +${r.penalty[1]}`);
    }
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Rubber over! Winning team: ${state.winnerTeam}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
