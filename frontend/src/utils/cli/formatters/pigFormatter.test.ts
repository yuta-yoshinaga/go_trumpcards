import { describe, expect, it } from 'vitest';
import type { Card, PigResponse } from '../../../types/card';
import { formatPigState } from './pigFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 4,
  cards: id === 0 ? [card('SPADE', 13), card('HEART', 13)] : [],
  letters: 0,
  letterWord: '',
  eliminated: false,
  hasSignalled: false,
  noticedOrder: 0,
  hasChosenPass: false,
  ...over,
});

const state = (over: Partial<PigResponse> = {}): PigResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    validPlays: [0, 1],
    signallerIdx: -1,
    noticedCnt: 0,
    roundLoserIdx: -1,
    roundNumber: 2,
    passCount: 3,
    deckSize: 16,
    currentPlayerIdx: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4, cpuDifficulty: 1 },
    message: '',
    ...over,
  }) as unknown as PigResponse;

describe('formatPigState', () => {
  it('reports loading for a null state', () => {
    expect(formatPigState(null)).toBe('Loading...');
  });

  it('shows the round, deck, passes and the rule', () => {
    const out = formatPigState(state());
    expect(out).toContain('round 2');
    expect(out).toContain('deck 16');
    expect(out).toContain('3 passes');
    expect(out).toContain('PASS');
    // **取り合うものが無いのが規則そのもの。**
    expect(out).toMatch(/signalled in silence/);
  });

  it('lists every seat with its hand size and letters', () => {
    const out = formatPigState(state({ players: [seat(0, { letters: 2, letterWord: 'PI' }), seat(1)] }));
    expect(out).toMatch(/>あなた: 4 cards, letters \[PI\]/);
    expect(out).toMatch(/ CPU 1: 4 cards, letters \[-\]/);
  });

  // **選び終えた席・気づいた席・脱落した席は盤面に痕跡が残らない。**
  it('marks who has chosen, who noticed and who is out', () => {
    expect(formatPigState(state({ players: [seat(0), seat(1, { hasChosenPass: true })] }))).toContain(
      'CPU 1[has chosen]',
    );
    expect(formatPigState(state({ players: [seat(0), seat(1, { hasSignalled: true, noticedOrder: 2 })] }))).toContain(
      'CPU 1[noticed 2]',
    );
    expect(formatPigState(state({ players: [seat(0), seat(1, { eliminated: true })] }))).toContain('CPU 1[out]');
    expect(formatPigState(state())).not.toContain('[has chosen]');
  });

  // **合図は声に出さない。** テキストで名乗らせるしかない。
  it('announces a live signal and the answer to it', () => {
    const out = formatPigState(state({ phase: 1, signallerIdx: 2, noticedCnt: 1 }));
    expect(out).toMatch(/signal now/);

    const done = formatPigState(
      state({ phase: 1, signallerIdx: 2, noticedCnt: 2, players: [seat(0, { hasSignalled: true }), seat(1)] }),
    );
    expect(done).toMatch(/you signalled/);
    expect(done).not.toMatch(/signal now/);
  });

  // **罰は1ラウンドに1回の出来事。** 配り直す前に読ませる。
  it('reports the round loser and offers the next deal', () => {
    const out = formatPigState(
      state({ phase: 2, roundLoserIdx: 1, players: [seat(0), seat(1, { letters: 1, letterWord: 'P' })] }),
    );
    expect(out).toMatch(/CPU 1 was last to notice/);
    expect(out).toContain('[P]');
    expect(out).toMatch(/next — deal the next round/);
  });

  it('says so when you have chosen and are waiting', () => {
    const out = formatPigState(state({ players: [seat(0, { hasChosenPass: true }), seat(1)] }));
    expect(out).toMatch(/waiting for everyone else/);
    // 負のコントロール: まだ選んでいなければ出さない。
    expect(formatPigState(state())).not.toMatch(/waiting for everyone else/);
  });

  // **脱落しても局は続く。**
  it('says so when you are out but the game continues', () => {
    const out = formatPigState(state({ players: [seat(0, { eliminated: true, cards: [] }), seat(1)] }));
    expect(out).toMatch(/you are out/);
    expect(formatPigState(state())).not.toMatch(/you are out/);
  });

  it('names the winner once the game ends', () => {
    expect(formatPigState(state({ phase: 3, gameEndFlag: true, winnerIdx: 0 }))).toMatch(
      /game over — あなた was the last one standing/,
    );
    expect(formatPigState(state({ phase: 3, gameEndFlag: true, winnerIdx: 2 }))).toMatch(
      /game over — CPU 2 was the last one standing/,
    );
  });

  it('appends the server message', () => {
    expect(formatPigState(state({ message: 'boom' }))).toContain('boom');
  });
});
