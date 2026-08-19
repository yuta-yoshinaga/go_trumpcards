import { describe, expect, it } from 'vitest';
import type { CardDesign, PochPlayer, PochResponse } from '../../../types/card';
import { formatPochState } from './pochFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<PochPlayer>): PochPlayer {
  return {
    id,
    isHuman,
    cardCount: 8,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 9)] : [],
    chips: -5,
    bet: 1,
    folded: false,
    hidden: !isHuman,
    ...overrides,
  };
}

const POOLS = ['ace', 'king', 'queen', 'jack', 'ten', 'marriage', 'sequence', 'pocher', 'centre'];

function makeState(overrides?: Partial<PochResponse>): PochResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    phase: 1,
    validPlays: [],
    currentPlayerIdx: 0,
    pools: POOLS.map((name, i) => ({ name, chips: i === 5 ? 12 : 4 })),
    paySuit: 1,
    turnUp: card('SPADE', 9),
    stakingAwards: [{ pool: 'marriage', player: 1, chips: 12 }],
    betTarget: 1,
    yourBestComboSize: 0,
    yourBestComboRank: 0,
    pochenWinner: -1,
    pochenPot: 0,
    playedPile: [],
    stopsSuit: -1,
    stopsRank: 0,
    dealNo: 0,
    targetDeals: 5,
    dealWinner: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatPochState', () => {
  it('prints both rules every frame', () => {
    // 「pay suit の札でしか取れない」と「pochen は宣言ではない」が誤解の的。
    const out = formatPochState(makeState());
    expect(out).toContain('turn-up suit');
    expect(out).toContain('no declaration');
  });

  it('prints all nine pools with their chips', () => {
    const out = formatPochState(makeState());
    for (const name of POOLS) {
      expect(out).toContain(`${name}:`);
    }
    // **持ち越しで膨らんだ区画**が読めること。
    expect(out).toContain('marriage:12');
  });

  it('reports what stage one paid out', () => {
    // 自動で解決するので、これが無いと何が起きたのか判らない。
    expect(formatPochState(makeState())).toContain('seat 1 takes marriage (12)');
  });

  it('says when the run is stopped', () => {
    expect(formatPochState(makeState())).toContain('the run is stopped');
    const live = formatPochState(makeState({ stopsSuit: 1, stopsRank: 9, playedPile: [card('SPADE', 9)] }));
    expect(live).not.toContain('the run is stopped');
    expect(live).toContain('played:');
  });

  it('prints each seat with its chips and bet', () => {
    const out = formatPochState(makeState());
    expect(out).toContain('-5 chips, bet 1');
    expect(out).toContain('8 cards'); // the hidden hand is a count
  });

  it('marks a folded seat', () => {
    const out = formatPochState(makeState({ players: [seat(0, true), seat(1, false, { folded: true })] }));
    expect(out).toContain('-- folded');
  });

  it('reports the pochen and deal winners', () => {
    const out = formatPochState(makeState({ pochenWinner: 1, pochenPot: 16, dealWinner: 2 }));
    expect(out).toContain('seat 1 won the pochen (16)');
    expect(out).toContain('seat 2 went out');
  });

  it('reports each ending', () => {
    expect(formatPochState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('most chips');
    expect(formatPochState(makeState({ gameEndFlag: true, winnerIdx: 2 }))).toContain('finish behind');
  });
});
