import { describe, expect, it } from 'vitest';
import type { BaseballPokerResponse, Card } from '../../../types/card';
import { BaseballPhase } from '../../../types/phases';
import { formatBaseballPokerState } from './baseballpokerFormatter';

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (over: Partial<BaseballPokerResponse['seats'][number]> = {}) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 0,
    cards: [card(1), card(2), card(5)],
    faceUp: [false, false, true],
    bonusCards: 0,
    folded: false,
    allIn: false,
    isTurn: false,
    isBuying: false,
    handRank: 0,
    usedWild: false,
    bestHand: [],
    wonAmount: 0,
    ...over,
  }) as BaseballPokerResponse['seats'][number];

const base = {
  phase: BaseballPhase.BETTING,
  seats: [seat({ isTurn: true }), seat({ name: 'CPU1', isHuman: false, cards: [null, null, card(7)] })],
  street: 1,
  streetTotal: 4,
  wildValues: [3, 9],
  bonusValue: 4,
  buyInValue: 3,
  pot: 40,
  currentBet: 0,
  toCall: 0,
  raiseCount: 0,
  canRaise: false,
  turnSeat: 0,
  humanSeat: 0,
  isHumanTurn: true,
  buyerSeat: -1,
  buyCost: 0,
  isBuying: false,
  handNumber: 2,
  remainingCards: 30,
  winnerSeat: 0,
  gameEndFlag: false,
  message: '',
} as BaseballPokerResponse;

const withState = (over: Partial<BaseballPokerResponse>): BaseballPokerResponse => ({ ...base, ...over });

describe('formatBaseballPokerState', () => {
  it('フェーズ・ハンド・ポット・ストリートを出す', () => {
    const out = formatBaseballPokerState(base);
    expect(out).toContain('BETTING');
    expect(out).toContain('Hand: 2');
    expect(out).toContain('pot: 40');
    expect(out).toContain('Street: 1 of 4');
  });

  // **ワイルドとイベントの値はサーバから来る。** 画面が 3 と 9 を持たない。
  it('サーバが送った値でワイルドとイベントを説明する', () => {
    expect(formatBaseballPokerState(base)).toContain('Wild: 3, 9');
    const other = formatBaseballPokerState(withState({ wildValues: [2, 7], bonusValue: 8, buyInValue: 2 }));
    expect(other).toContain('Wild: 2, 7');
    expect(other).toContain('face-up 8 pays a card');
    expect(other).toContain('face-up 2 buys the pot');
  });

  // **届いていない札だけを伏せる。** 表札は全席ぶん届いている。
  it('届いた表札を出し、届いていない札を伏せる', () => {
    const out = formatBaseballPokerState(base);
    expect(out.split('[??]').length - 1).toBe(2);

    const shown = formatBaseballPokerState(
      withState({
        seats: [seat(), seat({ name: 'CPU1', isHuman: false, cards: [card(11), card(12), card(7)] })],
      }),
    );
    expect(shown).not.toContain('[??]');
  });

  it('追加札と買い増しの印を出す', () => {
    const out = formatBaseballPokerState(
      withState({ seats: [seat({ bonusCards: 2, isBuying: true }), seat({ name: 'CPU1', isHuman: false })] }),
    );
    expect(out).toContain('+2 bonus');
    expect(out).toContain('$YOU');
  });

  it('買い増しの場面では額を出し、ベットの案内を出さない', () => {
    const out = formatBaseballPokerState(
      withState({ phase: BaseballPhase.BUY_IN, isBuying: true, buyerSeat: 0, buyCost: 80 }),
    );
    expect(out).toContain('Pay 80 to stay in');
    expect(out).not.toContain('You may check');
  });

  it('賭けの案内を場況で切り替える', () => {
    expect(formatBaseballPokerState(base)).toContain('You may check');
    expect(formatBaseballPokerState(withState({ toCall: 20 }))).toContain('20 to call');
  });

  it('獲得額とワイルド使用と勝者を出す', () => {
    const out = formatBaseballPokerState(
      withState({
        phase: BaseballPhase.GAME_END,
        gameEndFlag: true,
        winnerSeat: 1,
        seats: [seat(), seat({ name: 'CPU1', isHuman: false, wonAmount: 80, usedWild: true })],
      }),
    );
    expect(out).toContain('won 80 with wilds');
    expect(out).toContain('Winner: CPU1');
  });
});
