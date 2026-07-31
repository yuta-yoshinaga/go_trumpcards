import type { KaiserResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Bid', 'Discard', 'Play', 'HandEnd', 'GameEnd'];

const SUIT_GLYPHS: Record<number, string> = { 1: 'S', 2: 'C', 3: 'H', 4: 'D' };

const CONTRACT_NAMES = ['with trump', 'no trump', 'low no trump'];

/** Format a Kaiser game state as terminal text. */
export function formatKaiserState(state: KaiserResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Kaiser'));
  lines.push(`hand: ${state.handNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`target: ${state.targetScore}  team0: ${state.teamScores[0]}  team1: ${state.teamScores[1]}`);
  if (state.highBid) {
    const kind = CONTRACT_NAMES[state.contract] ?? '?';
    lines.push(`contract: ${state.highBid.value} ${kind} trump: ${SUIT_GLYPHS[state.trumpSuit] ?? '-'}`);
  }
  if (state.kittySize > 0) lines.push(`kitty: ${state.kittySize}`);
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const hand =
      p.cards.length > 0 ? p.cards.map((c, j) => `[${j}]${formatCard(c)}`).join(' ') : `hidden (${p.cardCount})`;
    const dealer = p.isDealer ? ' [dealer]' : '';
    const declarer = p.isDeclarer ? ' [declarer]' : '';
    const turn = p.isCurrentTurn && !state.gameEndFlag ? ' <- turn' : '';
    lines.push(`${name}(T${p.team})${dealer}${declarer}${turn}`);
    lines.push(`  ${hand}`);
  });
  lines.push('----------');

  if (state.trick.length > 0) lines.push(`trick: ${state.trick.map(formatCard).join(' ')}`);
  lines.push(`this hand: ${state.teamHandPoints[0]} to ${state.teamHandPoints[1]}`);
  // ♥5 と ♠3 の行方はトリック 8 点と同じ重みなので、必ず出す。
  if (state.heartFiveBy >= 0) {
    lines.push(`H5 (+5) taken by ${formatPlayerName(state.heartFiveBy, state.heartFiveBy === 0)}`);
  }
  if (state.spadeThreeBy >= 0) {
    lines.push(`S3 (-3) taken by ${formatPlayerName(state.spadeThreeBy, state.spadeThreeBy === 0)}`);
  }

  if (state.phase === 0 && state.bidPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push(`(your bid — "b <${state.minBid}-${state.maxBid}> [0-2]" in POINTS, or "ps" to pass)`);
  } else if (state.phase === 1 && state.declarerIdx === 0) {
    if (state.contract === 0 && state.trumpSuit === 0) {
      lines.push('(name trump with "t <1-4>")');
    } else {
      lines.push('(discard two with "d <i> <j>" — the H5 and S3 may not go)');
    }
  } else if (state.phase === 2 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    // 追随が強制なので、出せる札を出さないと操作できない。
    lines.push(`playable: ${state.validPlays.join(' ') || '-'}`);
    lines.push('(your turn — play with "p <i>")');
  } else if (state.phase === 3) {
    lines.push(
      state.bidMade ? '(the declaring side made it — "n" to deal again)' : '(SET — the bid comes off their score)',
    );
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerTeam >= 0) {
    lines.push(`Game Over! Winning team: ${state.winnerTeam}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
