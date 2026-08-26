import type { ContinentalRummyResponse } from '../../../types/card';
import { CONTINENTAL_RUMMY_PHASE } from '../../../types/games/continentalrummy';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<string, string> = {
  [CONTINENTAL_RUMMY_PHASE.DRAW]: 'DRAW',
  [CONTINENTAL_RUMMY_PHASE.DISCARD]: 'DISCARD',
  [CONTINENTAL_RUMMY_PHASE.ROUND_END]: 'ROUND END',
  [CONTINENTAL_RUMMY_PHASE.GAME_END]: 'GAME OVER',
};

const BONUS_NAMES: Record<string, string> = {
  win: 'going out',
  joker: 'jokers melded',
  noJoker: 'no joker used',
  firstTurn: 'out on the first turn',
  dealt: 'out on the dealt fifteen',
  oneSuit: 'all one suit',
};

/** Format a Continental Rummy game state as terminal text. */
export function formatContinentalRummyState(state: ContinentalRummyResponse): string {
  const lines: string[] = [formatHeader('Continental Rummy')];

  lines.push(`Phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}`);
  lines.push(`Round: ${state.roundNumber} / ${state.totalRounds}`);
  lines.push(`Stock: ${state.stockCount}${state.discardTop ? `  discard: ${formatCard(state.discardTop)}` : ''}`);
  // **上がれる形はサーバが返したものだけを並べる。** 15 の分割から組み直すと
  // 5+5+5 を勝手に足してしまう。
  lines.push(`Legal go-outs: ${state.layouts.map((l) => l.join('+')).join(' / ')} (sets are never melds)`);

  lines.push(formatSeparator());
  for (const p of state.players) {
    const who = p.isHuman ? 'You' : `CPU ${p.id}`;
    const marks = `${p.isDealer ? ' (dealer)' : ''}${p.id === state.currentPlayerIdx && !state.gameEndFlag ? ' <- to play' : ''}`;
    lines.push(`${who}${marks}: ${p.cardCount} card(s), score ${p.score}`);
    for (const run of p.melds) lines.push(`    ${run.map(formatCard).join(' ')}`);
  }

  const me = state.players.find((p) => p.isHuman);
  if (me && me.cards.length > 0) {
    lines.push(me.cards.map((c, i) => `[${i}]${formatCard(c)}`).join(' '));
  }
  // **上がれるときは黙っていない。** 15 枚の分割は目で追いきれない。
  if (state.canGoOutOnDeal) {
    lines.push('The dealt fifteen already go out: gooutdeal (worth more than going out after a draw)');
  } else if (state.goOutIdx >= 0) {
    lines.push(`You can go out: goout ${state.goOutIdx}`);
  }

  if (state.lastResult) {
    lines.push(formatSeparator());
    if (state.lastResult.winnerIdx < 0) {
      lines.push('Nobody went out - the round is washed out.');
    } else {
      const who = state.lastResult.winnerIdx === 0 ? 'You' : `CPU ${state.lastResult.winnerIdx}`;
      lines.push(`${who} went out:`);
      for (const b of state.lastResult.bonuses) {
        lines.push(`  ${BONUS_NAMES[b.key] ?? b.key}: ${b.points}`);
      }
      // **取り立てるのは相手 1 人あたり。** 合計だけだと何倍したのか読めない。
      lines.push(`  ${state.lastResult.perOpponent} from each opponent, ${state.lastResult.total} in all`);
    }
  }

  if (state.gameEndFlag) {
    lines.push(state.winnerIdx === 0 ? 'You win!' : state.winnerIdx < 0 ? 'A draw.' : `CPU ${state.winnerIdx} wins.`);
  }

  return lines.join('\n');
}
