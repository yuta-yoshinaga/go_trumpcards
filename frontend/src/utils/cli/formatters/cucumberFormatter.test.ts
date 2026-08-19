import { describe, expect, it } from 'vitest';
import type { Card, CucumberResponse } from '../../../types/card';
import { formatCucumberState } from './cucumberFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 7,
  cards: id === 0 ? [card('SPADE', 3), card('HEART', 10)] : [],
  penalty: 0,
  ...over,
});

const state = (over: Partial<CucumberResponse> = {}): CucumberResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    validPlays: [1],
    highestInTrick: 9,
    forced: false,
    currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 9) }],
    currentPlayerIdx: 0,
    leadPlayerIdx: 1,
    totalTricks: 7,
    trickNumber: 2,
    roundNumber: 3,
    lastTrickWinnerIdx: -1,
    lastPenalty: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4, targetScore: 30 },
    message: '',
    ...over,
  }) as unknown as CucumberResponse;

describe('formatCucumberState', () => {
  it('reports loading for a null state', () => {
    expect(formatCucumberState(null)).toBe('Loading...');
  });

  it('shows the round, trick, target and the rule', () => {
    const out = formatCucumberState(state());
    expect(out).toContain('round 3');
    // **失点が出るのは最終トリックだけ** (#5768)。あと何回かが読めること。
    expect(out).toContain('trick 3/7');
    expect(out).toContain('ends at 30');
    expect(out).toMatch(/suits are irrelevant/);
  });

  // **超えるべきランクは盤面から数えさせない。**
  it('states the rank to beat, or that you lead', () => {
    expect(formatCucumberState(state())).toMatch(/rank to beat: 9/);
    expect(formatCucumberState(state({ highestInTrick: 0, currentTrick: [] }))).toMatch(/you lead/);
  });

  it('lists every seat with its hand size and penalty', () => {
    const out = formatCucumberState(state({ players: [seat(0, { penalty: 12 }), seat(1)] }));
    expect(out).toMatch(/>あなた: 7 cards, 12 penalty/);
    expect(out).toMatch(/ CPU 1: 7 cards, 0 penalty/);
  });

  it('marks who took the last trick and for how much', () => {
    expect(formatCucumberState(state({ lastTrickWinnerIdx: 2, lastPenalty: 11 }))).toContain('CPU 2[last trick +11]');
    expect(formatCucumberState(state())).not.toContain('[last trick');
  });

  it('stars the legal plays in your hand', () => {
    expect(formatCucumberState(state())).toMatch(/\[1\]\S+\*/);
  });

  // **合法手が1つ = 更新できない、ではない。** サーバの forced をそのまま使う。
  it('says when the play is forced, and only then', () => {
    expect(formatCucumberState(state({ forced: true, validPlays: [0] }))).toMatch(/cannot beat it/);
    // 合法手は1つだが更新できる場面では出さない。
    expect(formatCucumberState(state({ forced: false, validPlays: [1] }))).not.toMatch(/cannot beat it/);
  });

  // **失点はラウンドに1回だけの出来事。**
  it('reports the round penalty and offers the next deal', () => {
    const out = formatCucumberState(state({ phase: 1, lastTrickWinnerIdx: 1, lastPenalty: 13 }));
    expect(out).toMatch(/CPU 1 took the last trick — 13 penalty/);
    expect(out).toMatch(/next — deal the next round/);
  });

  it('names the winner once the game ends', () => {
    const out = formatCucumberState(
      state({ phase: 2, gameEndFlag: true, winnerIdx: 0, players: [seat(0, { penalty: 8 }), seat(1)] }),
    );
    expect(out).toMatch(/game over — あなた finished with the fewest \(8\)/);
  });

  it('appends the server message', () => {
    expect(formatCucumberState(state({ message: 'boom' }))).toContain('boom');
  });
});
