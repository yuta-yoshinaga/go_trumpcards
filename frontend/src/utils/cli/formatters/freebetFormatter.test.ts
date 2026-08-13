import { describe, expect, it } from 'vitest';
import type { Card, FreeBetResponse } from '../../../types/card';
import { FREE_BET_RESULT } from '../../../types/games/freebet';
import { FreeBetPhase } from '../../../types/phases';
import { formatFreeBetState } from './freebetFormatter';

const card = (value: number): Card => ({ design: 'SPADE', value });

const hand = (over: Partial<FreeBetResponse['hands'][number]> = {}) =>
  ({
    cards: [card(13), card(7)],
    score: 17,
    bet: 50,
    freeBet: 0,
    isSoft: false,
    stood: false,
    doubled: false,
    busted: false,
    blackjack: false,
    result: 0,
    ...over,
  }) as FreeBetResponse['hands'][number];

const base = {
  phase: FreeBetPhase.BET,
  hands: [],
  activeHand: 0,
  dealerCards: [],
  dealerScore: 0,
  dealerPushed22: false,
  canFreeDouble: false,
  canFreeSplit: false,
  anteBet: 0,
  payout: 0,
  chips: 1000,
  roundNumber: 1,
  remainingCards: 312,
  gameEndFlag: false,
  message: '',
} as unknown as FreeBetResponse;

const at = (over: Partial<FreeBetResponse>) => ({ ...base, ...over }) as FreeBetResponse;

describe('formatFreeBetState', () => {
  it('フェーズとチップを出す', () => {
    const out = formatFreeBetState(base);
    expect(out).toContain('BET');
    expect(out).toContain('chips: 1000');
  });

  it('賭けた後はアンティを出す', () => {
    expect(formatFreeBetState(at({ anteBet: 50 }))).toContain('Ante 50');
  });

  it('ディーラーの札と点数を出す', () => {
    const out = formatFreeBetState(at({ dealerCards: [card(6), card(9)], dealerScore: 15 }));
    expect(out).toContain('Dealer:');
    expect(out).toContain('= 15');
  });

  // **ハウス持ちのぶんを合算しない。** 合算すると「いくら失うのか」が読めなくなる。
  it('自分の賭け金とハウスの出資を別々に出す', () => {
    const out = formatFreeBetState(
      at({ phase: FreeBetPhase.PLAY, hands: [hand({ bet: 50, freeBet: 50, doubled: true })] }),
    );
    expect(out).toContain('bet 50');
    expect(out).toContain('house 50');
    expect(out).not.toContain('bet 100');
  });

  it('ハウスの出資が無い手札には house を出さない', () => {
    const out = formatFreeBetState(at({ phase: FreeBetPhase.PLAY, hands: [hand()] }));
    expect(out).toContain('bet 50');
    expect(out).not.toContain('house');
  });

  it('操作中の手札に印を付ける', () => {
    const out = formatFreeBetState(at({ phase: FreeBetPhase.PLAY, hands: [hand(), hand()], activeHand: 1 }));
    const lines = out.split('\n').filter((l) => l.includes('] '));
    expect(lines[0]?.startsWith(' ')).toBe(true);
    expect(lines[1]?.startsWith('*')).toBe(true);
  });

  it('いま無料でできる操作を出す', () => {
    const out = formatFreeBetState(
      at({ phase: FreeBetPhase.PLAY, hands: [hand()], canFreeDouble: true, canFreeSplit: true }),
    );
    expect(out).toContain('Free now:');
    expect(out).toContain('freedouble');
    expect(out).toContain('freesplit');
    expect(formatFreeBetState(at({ phase: FreeBetPhase.PLAY, hands: [hand()] }))).not.toContain('Free now:');
  });

  it('決着で勝敗を出す', () => {
    const out = formatFreeBetState(
      at({
        phase: FreeBetPhase.RESULT,
        hands: [hand({ result: FREE_BET_RESULT.win }), hand({ result: FREE_BET_RESULT.blackjack })],
      }),
    );
    expect(out).toContain('Hand 1: win');
    expect(out).toContain('Hand 2: blackjack (3:2)');
  });

  // **22 は名指しする。** 無料ダブル / 無料スプリットの対価がこれなので、
  // 黙って引き分けにすると規則が読み取れない。
  it('ディーラーの22を名指しする', () => {
    const out = formatFreeBetState(
      at({
        phase: FreeBetPhase.RESULT,
        dealerPushed22: true,
        dealerScore: 22,
        hands: [hand({ result: FREE_BET_RESULT.dealer22Push })],
      }),
    );
    expect(out).toContain('22');
    expect(out).toContain('Hand 1: push (dealer 22)');
  });

  it('破産を出す', () => {
    expect(formatFreeBetState(at({ gameEndFlag: true }))).toContain('Out of chips.');
  });
});
