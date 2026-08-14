import { describe, expect, it } from 'vitest';
import type { Card, IronCrossResponse } from '../../../types/card';
import { IronCrossPhase } from '../../../types/phases';
import { formatIronCrossState } from './ironcrossFormatter';

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
    isTurn: false,
    line: 0,
    handRank: 0,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as IronCrossResponse['seats'][number];

const base = {
  phase: IronCrossPhase.BETTING,
  seats: [seat({ isTurn: true }), seat({ name: 'CPU1', isHuman: false, cards: [] })],
  cross: [null, null, null, null, null],
  revealedCount: 0,
  crossTotal: 5,
  verticalIndexes: [1, 0, 2],
  horizontalIndexes: [3, 0, 4],
  pot: 40,
  currentBet: 0,
  toCall: 0,
  raiseCount: 0,
  canRaise: false,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  isChoosing: false,
  handNumber: 2,
  remainingCards: 30,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
} as IronCrossResponse;

const withState = (over: Partial<IronCrossResponse>): IronCrossResponse => ({ ...base, ...over });

describe('formatIronCrossState', () => {
  it('フェーズ・ハンド・ポットを出す', () => {
    const out = formatIronCrossState(base);
    expect(out).toContain('BETTING');
    expect(out).toContain('Hand: 2');
    expect(out).toContain('pot: 40');
  });

  // **十字は十字の形で出す。** 1 行に並べると、どれが縦でどれが横か分からない。
  it('十字を3行に分けて出す', () => {
    // 中央=A, 上=2, 下=3, 左=4, 右=5
    const out = formatIronCrossState(
      withState({ cross: [card(1), card(2), card(3), card(4), card(5)], revealedCount: 5 }),
    );
    const lines = out.split('\n');
    const middle = lines.findIndex((l) => l.includes('4') && l.includes('5'));
    expect(middle).toBeGreaterThan(0);
    // 上下は中央行の直上・直下。
    expect(lines[middle - 1]).toContain('2');
    expect(lines[middle + 1]).toContain('3');
  });

  it('伏せている位置を印で埋める', () => {
    const out = formatIronCrossState(base);
    // 5 か所すべて伏せている。
    expect(out.split('[??]').length - 1).toBe(5);
    expect(out).toContain('Cross: 0 of 5 shown');
  });

  // **CPU の手札はサーバが送っていない。** 空なら伏せ表示、届けば開く。
  it('届いていない手札を伏せ、届いた手札を開く', () => {
    expect(formatIronCrossState(base)).toContain('(face down)');
    const shown = formatIronCrossState(withState({ seats: [seat(), seat({ name: 'CPU1', isHuman: false })] }));
    expect(shown).not.toContain('(face down)');
  });

  it('選んだ列を届いた席にだけ出す', () => {
    expect(formatIronCrossState(base)).not.toContain('[vertical]');
    const out = formatIronCrossState(
      withState({ seats: [seat({ line: 1 }), seat({ name: 'CPU1', isHuman: false, cards: [], line: 2 })] }),
    );
    expect(out).toContain('[vertical]');
    expect(out).toContain('[horizontal]');
  });

  it('選ぶ場面では選ぶよう促し、ベットの案内を出さない', () => {
    const out = formatIronCrossState(withState({ phase: IronCrossPhase.CHOOSE_LINE, isChoosing: true }));
    expect(out).toContain('Choose vertical or horizontal');
    expect(out).not.toContain('You may check');
  });

  it('賭けの案内を場況で切り替える', () => {
    expect(formatIronCrossState(base)).toContain('You may check');
    expect(formatIronCrossState(withState({ toCall: 20 }))).toContain('20 to call');
  });

  it('獲得額と勝者を出す', () => {
    const out = formatIronCrossState(
      withState({
        phase: IronCrossPhase.GAME_END,
        gameEndFlag: true,
        winnerSeat: 1,
        seats: [seat(), seat({ name: 'CPU1', isHuman: false, cards: [], wonAmount: 80 })],
      }),
    );
    expect(out).toContain('won 80');
    expect(out).toContain('Winner: CPU1');
  });
});
