import { describe, expect, it } from 'vitest';
import type { Card, CincinnatiResponse } from '../../../types/card';
import { CincinnatiPhase } from '../../../types/phases';
import { formatCincinnatiState } from './cincinnatiFormatter';

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<CincinnatiResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [card(1), card(2), card(3), card(4), card(5)],
    folded: false,
    allIn: false,
    isTurn: false,
    handRank: 0,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as CincinnatiResponse['seats'][number];

const base = {
  phase: CincinnatiPhase.BETTING,
  seats: [seat({ isTurn: true }), seat({ name: 'CPU1', isHuman: false, cards: [] })],
  community: [],
  revealedCount: 0,
  communityTotal: 5,
  pot: 40,
  currentBet: 0,
  toCall: 0,
  raiseCount: 0,
  canRaise: false,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  handNumber: 2,
  remainingCards: 30,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
} as unknown as CincinnatiResponse;

const at = (over: Partial<CincinnatiResponse>) => ({ ...base, ...over }) as CincinnatiResponse;

describe('formatCincinnatiState', () => {
  it('フェーズとハンドとポットを出す', () => {
    const out = formatCincinnatiState(base);
    expect(out).toContain('BETTING');
    expect(out).toContain('Hand: 2');
    expect(out).toContain('pot: 40');
  });

  // **あと何枚めくれるかを出す。** 残りの回数だけベットラウンドがある。
  it('公開枚数と総数を出す', () => {
    expect(formatCincinnatiState(base)).toContain('0 of 5 shown');
    expect(formatCincinnatiState(at({ community: [card(7), card(8)], revealedCount: 2 }))).toContain('2 of 5 shown');
  });

  it('配る前の場はプレースホルダを出す', () => {
    expect(formatCincinnatiState(base)).toContain('Board: -');
  });

  // **CPU の手札はサーバが送っていない。** 空なら伏せ表示にする。
  it('届いていない手札は伏せ表示にする', () => {
    const out = formatCincinnatiState(base);
    expect(out).toContain('(face down)');
    // 人間の 5 枚は出ている。
    expect(out.split('\n').find((l) => l.includes('YOU'))).not.toContain('(face down)');
  });

  it('ショーダウンでは届いた手札を開く', () => {
    const out = formatCincinnatiState(
      at({
        phase: CincinnatiPhase.SHOWDOWN,
        seats: [seat(), seat({ name: 'CPU1', isHuman: false })],
      }),
    );
    expect(out).not.toContain('(face down)');
  });

  it('手番に印を付け、降りと全ツッパを書く', () => {
    const out = formatCincinnatiState(
      at({ seats: [seat({ isTurn: true }), seat({ name: 'CPU1', isHuman: false, folded: true, cards: [] })] }),
    );
    const lines = out.split('\n').filter((l) => l.includes('chips'));
    expect(lines[0]?.startsWith('*')).toBe(true);
    expect(out).toContain('(folded)');
    expect(
      formatCincinnatiState(at({ seats: [seat({ name: 'CPU1', isHuman: false, allIn: true, cards: [] })] })),
    ).toContain('(all in)');
  });

  it('コールに必要な額を出す', () => {
    expect(formatCincinnatiState(base)).toContain('You may check');
    expect(formatCincinnatiState(at({ toCall: 20 }))).toContain('20 to call');
  });

  it('獲得額と勝者を出す', () => {
    const out = formatCincinnatiState(at({ seats: [seat({ wonAmount: 80 })], gameEndFlag: true, winnerSeat: 0 }));
    expect(out).toContain('won 80');
    expect(out).toContain('Winner: YOU');
  });
});
