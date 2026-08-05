import { describe, expect, it } from 'vitest';
import type { CardDesign, PopeJoanPlayer, PopeJoanResponse } from '../../../types/card';
import { formatPopeJoanState } from './popejoanFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

const COMPS = ['ace', 'king', 'queen', 'jack', 'game', 'pope', 'matrimony', 'intrigue'];
const DRESS = [1, 1, 1, 1, 1, 6, 2, 2];

function seat(id: number, isHuman: boolean, overrides?: Partial<PopeJoanPlayer>): PopeJoanPlayer {
  return {
    id,
    isHuman,
    cardCount: 10,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 9)] : [],
    chips: -15,
    holdsPope: false,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<PopeJoanResponse>): PopeJoanResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    phase: 0,
    validPlays: [],
    currentPlayerIdx: 0,
    compartments: COMPS.map((name, i) => ({ name, chips: DRESS[i] })),
    trumpSuit: 1,
    turnUp: card('SPADE', 5),
    awards: [],
    playedPile: [],
    runSuit: -1,
    runRank: 0,
    dealNo: 0,
    targetDeals: 5,
    dealWinner: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatPopeJoanState', () => {
  it('prints both rules every frame', () => {
    const out = formatPopeJoanState(makeState());
    expect(out).toContain('only on a trump');
    expect(out).toContain('8D is out');
  });

  it("prints all eight compartments with the dealer's fixed dress", () => {
    const out = formatPopeJoanState(makeState());
    for (const name of COMPS) {
      expect(out).toContain(`${name}:`);
    }
    // ポープは 6 から始まる。
    expect(out).toContain('pope:6');
    expect(out).toContain('matrimony:2');
  });

  it('marks a turn-up award apart from an ordinary one', () => {
    const byTurnUp = formatPopeJoanState(
      makeState({ awards: [{ compartment: 'pope', player: 0, chips: 6, byTurnUp: true }] }),
    );
    expect(byTurnUp).toContain('from the turn-up');

    const plain = formatPopeJoanState(
      makeState({ awards: [{ compartment: 'king', player: 1, chips: 1, byTurnUp: false }] }),
    );
    expect(plain).toContain('takes king (1)');
    expect(plain).not.toContain('from the turn-up');
  });

  it('says when the run is stopped', () => {
    expect(formatPopeJoanState(makeState())).toContain('the run is stopped');
    const live = formatPopeJoanState(makeState({ runSuit: 1, runRank: 5, playedPile: [card('SPADE', 5)] }));
    expect(live).not.toContain('the run is stopped');
    expect(live).toContain('played:');
  });

  // 支払い免除の有無は精算を読むのに要る。
  it('marks the Pope holder', () => {
    const out = formatPopeJoanState(makeState({ players: [seat(0, true), seat(1, false, { holdsPope: true })] }));
    expect(out).toContain('holds the Pope');
  });

  it('prints each seat with its chips', () => {
    const out = formatPopeJoanState(makeState());
    expect(out).toContain('-15 chips');
    expect(out).toContain('10 cards'); // the hidden hand is a count
  });

  it('reports the deal winner and each ending', () => {
    expect(formatPopeJoanState(makeState({ dealWinner: 2 }))).toContain('seat 2 went out');
    expect(formatPopeJoanState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('most chips');
    expect(formatPopeJoanState(makeState({ gameEndFlag: true, winnerIdx: 2 }))).toContain('finish behind');
  });
});
