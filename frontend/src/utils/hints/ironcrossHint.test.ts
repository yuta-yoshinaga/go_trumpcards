import { describe, expect, it } from 'vitest';
import type { Card, IronCrossResponse } from '../../types/card';
import { IronCrossPhase } from '../../types/phases';
import { getIroncrossHint, ironcrossLineName } from './ironcrossHint';

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<IronCrossResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [card(1), card(2), card(3), card(4)],
    folded: false,
    allIn: false,
    isTurn: true,
    line: 0,
    handRank: 0,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as IronCrossResponse['seats'][number];

const state = (over: Partial<IronCrossResponse> = {}) =>
  ({
    phase: IronCrossPhase.BETTING,
    seats: [seat()],
    cross: [null, null, null, null, null],
    revealedCount: 0,
    crossTotal: 5,
    verticalIndexes: [1, 0, 2],
    horizontalIndexes: [3, 0, 4],
    pot: 40,
    currentBet: 0,
    toCall: 0,
    raiseCount: 0,
    canRaise: true,
    turnSeat: 0,
    humanSeat: 0,
    isHumanTurn: true,
    isChoosing: false,
    handNumber: 1,
    remainingCards: 30,
    winnerSeat: 0,
    gameEndFlag: false,
    message: '',
    config: { seats: 4, initialChips: 1000, ante: 10 },
    ...over,
  }) as IronCrossResponse;

describe('getIroncrossHint', () => {
  it('終局では助言しない', () => {
    expect(getIroncrossHint(state({ gameEndFlag: true }))).toBeNull();
  });

  it('他人の手番では助言しない', () => {
    expect(getIroncrossHint(state({ isHumanTurn: false }))).toBeNull();
  });

  // **選ぶ場面が最優先。** ここだけは取り返しがつかない。
  it('選ぶ場面では列を選ぶよう促す', () => {
    const hint = getIroncrossHint(state({ phase: IronCrossPhase.CHOOSE_LINE, isChoosing: true, isHumanTurn: false }));
    expect(hint?.targetAction).toBe('line');
    expect(hint?.confidence).toBe('strong');
  });

  it('もう選んでいれば助言しない', () => {
    expect(
      getIroncrossHint(
        state({
          phase: IronCrossPhase.CHOOSE_LINE,
          isChoosing: true,
          isHumanTurn: false,
          seats: [seat({ line: 1 })],
        }),
      ),
    ).toBeNull();
  });

  it('賭けが無ければチェックを薦める', () => {
    expect(getIroncrossHint(state())?.targetAction).toBe('check');
  });

  it('役が強ければ賭けを薦める', () => {
    expect(getIroncrossHint(state({ seats: [seat({ handRank: 3 })] }))?.targetAction).toBe('bet');
  });

  it('レイズ上限に達していたら賭けを薦めない', () => {
    expect(getIroncrossHint(state({ seats: [seat({ handRank: 3 })], canRaise: false }))?.targetAction).toBe('check');
  });

  it('役がスリーカード以上ならレイズを薦める', () => {
    expect(getIroncrossHint(state({ toCall: 20, seats: [seat({ handRank: 4 })] }))?.targetAction).toBe('raise');
  });

  it('役があればコールを薦める', () => {
    expect(getIroncrossHint(state({ toCall: 20, seats: [seat({ handRank: 2 })] }))?.targetAction).toBe('call');
  });

  it('安ければ役が無くてもコールを薦める', () => {
    expect(getIroncrossHint(state({ toCall: 5 }))?.targetAction).toBe('call');
  });

  it('役が無く高ければ降りるよう薦める', () => {
    expect(getIroncrossHint(state({ toCall: 200 }))?.targetAction).toBe('fold');
  });

  it('席が見つからなければ助言しない', () => {
    expect(getIroncrossHint(state({ humanSeat: 9 }))).toBeNull();
  });
});

describe('ironcrossLineName', () => {
  it('列の値を名前にする', () => {
    expect(ironcrossLineName(1)).toBe('vertical');
    expect(ironcrossLineName(2)).toBe('horizontal');
    expect(ironcrossLineName(0)).toBeNull();
  });
});
