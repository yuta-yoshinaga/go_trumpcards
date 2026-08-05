import { describe, expect, it } from 'vitest';
import type { CardDesign, NainJaunePlayer, NainJauneResponse } from '../../../types/card';
import { formatNainJauneState } from './nainjauneFormatter';

const card = (design: CardDesign, value: number) => ({ design, value });

const BOXES = [
  { name: 'ten', chips: 4, card: card('DIAMOND', 10) },
  { name: 'jack', chips: 8, card: card('CLOVER', 11) },
  { name: 'queen', chips: 12, card: card('SPADE', 12) },
  { name: 'king', chips: 16, card: card('HEART', 13) },
  { name: 'dwarf', chips: 20, card: card('DIAMOND', 7) },
];

function seat(id: number, isHuman: boolean, overrides?: Partial<NainJaunePlayer>): NainJaunePlayer {
  return {
    id,
    isHuman,
    cardCount: 12,
    cards: isHuman ? [card('SPADE', 1), card('HEART', 13)] : [],
    chips: -15,
    points: 63,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<NainJauneResponse>): NainJauneResponse {
  return {
    players: [seat(0, true), seat(1, false)],
    phase: 0,
    validPlays: [],
    currentPlayerIdx: 0,
    boxes: BOXES,
    talonCount: 4,
    awards: [],
    playedPile: [],
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

describe('formatNainJauneState', () => {
  it('prints both rules every frame', () => {
    const out = formatNainJauneState(makeState());
    expect(out).toContain('IGNORES SUIT');
    expect(out).toContain('POINTS, not cards');
  });

  // **区画はスートまで一致した1枚でしか取れない。**札を出さないと判断できない。
  it('prints each box with the exact card that claims it', () => {
    const out = formatNainJauneState(makeState());
    expect(out).toContain('dwarf(');
    expect(out).toContain('):20');
    for (const b of BOXES) {
      expect(out).toContain(`${b.name}(`);
    }
  });

  it('says whether the run is stopped or what comes next', () => {
    expect(formatNainJauneState(makeState())).toContain('the run is stopped');
    const live = formatNainJauneState(makeState({ runRank: 5, playedPile: [card('SPADE', 5)] }));
    expect(live).toContain('next up: a 6 of any suit');
    expect(live).not.toContain('the run is stopped');
  });

  // 支払いは点数なので、枚数だけでは負債額が読めない。
  it('prints what each hand is worth, not just its size', () => {
    const out = formatNainJauneState(makeState());
    expect(out).toContain('12 card(s) worth 63');
    expect(out).toContain('-15 chips');
  });

  it('reports awards, the deal winner and each ending', () => {
    expect(formatNainJauneState(makeState({ awards: [{ box: 'dwarf', player: 1, chips: 20 }] }))).toContain(
      'seat 1 takes dwarf (20)',
    );
    expect(formatNainJauneState(makeState({ dealWinner: 2 }))).toContain('seat 2 went out');
    expect(formatNainJauneState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('most chips');
    expect(formatNainJauneState(makeState({ gameEndFlag: true, winnerIdx: 2 }))).toContain('finish behind');
  });
});
