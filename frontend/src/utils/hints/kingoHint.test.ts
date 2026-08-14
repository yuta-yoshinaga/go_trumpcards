import { describe, expect, it } from 'vitest';
import type { KingoResponse } from '../../types/card';
import { KingoPhase } from '../../types/phases';
import { getKingoHint } from './kingoHint';

const seat = (over: Partial<KingoResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [],
    rank: 0,
    matchedValue: 0,
    isBanker: false,
    wonAmount: 0,
    ...over,
  }) as KingoResponse['seats'][number];

const state = (over: Partial<KingoResponse> = {}) =>
  ({
    phase: KingoPhase.BET,
    seats: [seat()],
    bankerSeat: 1,
    roundNumber: 1,
    rounds: 10,
    humanSeat: 0,
    isHumanBanker: false,
    isHumanTurn: true,
    handSize: 3,
    payoutArashi: 3,
    payoutPair: 1,
    remainingCards: 34,
    winnerSeat: 0,
    gameEndFlag: false,
    message: '',
    config: { seats: 4, initialChips: 1000, minBet: 10, rounds: 10 },
    ...over,
  }) as KingoResponse;

describe('getKingoHint', () => {
  it('終局では助言しない', () => {
    expect(getKingoHint(state({ gameEndFlag: true }))).toBeNull();
  });

  it('決着後は次のラウンドを薦める', () => {
    expect(getKingoHint(state({ phase: KingoPhase.RESULT }))?.targetAction).toBe('next');
  });

  // **親と子で薦める操作が変わる。**
  it('親には配るよう薦める', () => {
    const hint = getKingoHint(state({ isHumanBanker: true }));
    expect(hint?.targetAction).toBe('deal');
  });

  it('子には張るよう薦める', () => {
    expect(getKingoHint(state())?.targetAction).toBe('bet');
  });

  // **手札は配る前なので、助言できるのは張りの重さだけ。**
  it('手持ちが薄いときは理由が変わる', () => {
    const rich = getKingoHint(state());
    const poor = getKingoHint(state({ seats: [seat({ chips: 30 })] }));
    expect(rich?.targetAction).toBe('bet');
    expect(poor?.targetAction).toBe('bet');
    expect(poor?.reason).not.toBe(rich?.reason);
  });

  it('席が見つからなければ助言しない', () => {
    expect(getKingoHint(state({ humanSeat: 9 }))).toBeNull();
  });
});
