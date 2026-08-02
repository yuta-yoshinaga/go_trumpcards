import { describe, expect, it } from 'vitest';
import type { Card, CassinoResponse } from '../../types/card';
import { getCassinoHint } from './cassinoHint';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<CassinoResponse> = {}): CassinoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [card('HEART', 5)],
        capturedCount: 0,
        sweepCount: 0,
        totalScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 2,
        cards: [],
        capturedCount: 0,
        sweepCount: 0,
        totalScore: 0,
      },
    ],
    currentTurn: 0,
    tableCards: [],
    builds: [],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 21, multiBuildEnabled: true, sweepBonusEnabled: true, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 0,
    packsDealt: 1,
    roundWinners: [],
    lastRoundDetail: null,
    message: '',
    ...overrides,
  };
}

describe('getCassinoHint', () => {
  it('returns null if game ended', () => {
    expect(getCassinoHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null if no human found', () => {
    expect(getCassinoHint(makeState({ players: [] }))).toBeNull();
  });

  it('returns null if human has no cards', () => {
    const state = makeState({
      players: [{ id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 }],
    });
    expect(getCassinoHint(state)).toBeNull();
  });

  it('recommends take when point cards are on the table', () => {
    const state = makeState({
      tableCards: [card('SPADE', 5)], // ♠5 — point (spade)
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 5)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    const hint = getCassinoHint(state);
    expect(hint?.targetAction).toBe('take');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends take (moderate) for most-cards race', () => {
    const state = makeState({
      // 4♥ + 5♥ = 9。**合計で取れる**ので 2 枚取り。どちらも得点札ではない。
      // 以前ここは 7♥ 8♥ で、9 では取れないのに take を期待していた
      // ——「場が 2 枚以上なら take」という実装の誤りをそのまま固定していた。
      // 13♥ は 9 では取れないので場が残る。**残さないとスイープ判定が先に立ち**、
      // moderate ではなく strong が返って「最多カード争い」の枝を検証できない。
      tableCards: [card('HEART', 4), card('HEART', 5), card('HEART', 13)],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('DIAMOND', 9)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    const hint = getCassinoHint(state);
    expect(hint?.targetAction).toBe('take');
    expect(hint?.confidence).toBe('moderate');
  });

  it('recommends trail when there is nothing useful to take', () => {
    const state = makeState({
      tableCards: [],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 5)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    const hint = getCassinoHint(state);
    expect(hint?.targetAction).toBe('trail');
  });

  it('captures a sum, not only a rank match', () => {
    // 7 は 3+4 を取れる (`partitionIntoSumGroups`)。同ランクしか見ないと見落とす。
    const state = makeState({
      tableCards: [card('HEART', 3), card('HEART', 4)],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('DIAMOND', 7)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    expect(getCassinoHint(state)?.targetAction).toBe('take');
  });

  it('captures with a face card, which the old version skipped outright', () => {
    // 絵札は同ランクの絵札を取る。K♠ は得点札 (スペード) でもある。
    const state = makeState({
      // 6♥ は K では取れないので場が残る（残さないとスイープ扱いになる）。
      tableCards: [card('SPADE', 13), card('HEART', 6)],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 13)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    expect(getCassinoHint(state)?.reason).toBe('hint.take.points');
  });

  it('does not let a face card take a numeric card of the same value', () => {
    // 絵札は絵札としか組めない。値だけで比べると J(11) が数札を取れることになる。
    const state = makeState({
      tableCards: [card('HEART', 5), card('CLOVER', 6)],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 11)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    expect(getCassinoHint(state)?.targetAction).toBe('trail');
  });

  it('calls out a sweep when the capture empties the table', () => {
    const state = makeState({
      tableCards: [card('HEART', 4), card('CLOVER', 5)],
      builds: [],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 9)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    expect(getCassinoHint(state)?.reason).toBe('hint.take.sweep');
  });

  it('is not a sweep while a build is still standing', () => {
    // スイープは場**と**ビルドが空になること (`Cassino.go:330`)。
    const state = makeState({
      tableCards: [card('HEART', 4), card('CLOVER', 5)],
      builds: [{ ownerIdx: 1, value: 8, groups: [[card('SPADE', 8)]], isMulti: false }],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 9)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    expect(getCassinoHint(state)?.reason).not.toBe('hint.take.sweep');
  });

  it('does not promise a sweep bonus when the local rule is off', () => {
    // `sweepBonusEnabled` が false なら場を空にしても加点は無い
    // (`Cassino.go:331`)。同じ盤面でも文言を変える。
    const state = makeState({
      tableCards: [card('HEART', 4), card('CLOVER', 5)],
      builds: [],
      config: { targetScore: 21, multiBuildEnabled: false, sweepBonusEnabled: false, cpuDifficulty: 1 },
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [card('HEART', 9)],
          capturedCount: 0,
          sweepCount: 0,
          totalScore: 0,
        },
      ],
    });
    expect(getCassinoHint(state)?.reason).not.toBe('hint.take.sweep');
  });
});
