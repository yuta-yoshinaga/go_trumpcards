import { describe, expect, it } from 'vitest';
import type { BanLuckResponse, Card } from '../../types/card';
import { BanLuckPhase } from '../../types/phases';
import { getBanluckHint } from './banluckHint';

const card = (value: number): Card => ({ design: 'SPADE', value });

const seat = (cards: Card[], score: number) =>
  ({
    name: 'YOU',
    isHuman: true,
    chips: 1000,
    bet: 50,
    cards,
    score,
    rank: 1,
    outcome: 1,
    roundBet: 50,
    delta: 0,
    busted: false,
    stood: false,
    isBanker: false,
    isTurn: true,
  }) as BanLuckResponse['seats'][number];

const state = (over: Partial<BanLuckResponse> = {}) =>
  ({
    phase: BanLuckPhase.PLAY,
    seats: [seat([card(10), card(5)], 15)],
    bankerSeat: 1,
    turnSeat: 0,
    humanSeat: 0,
    isHumanTurn: true,
    mustHit: false,
    roundNumber: 1,
    remainingCards: 40,
    winnerSeat: 0,
    gameEndFlag: false,
    message: '',
    ...over,
  }) as BanLuckResponse;

describe('getBanluckHint', () => {
  it('終局・他フェーズ・他人の手番では助言しない', () => {
    expect(getBanluckHint(state({ gameEndFlag: true }))).toBeNull();
    expect(getBanluckHint(state({ phase: BanLuckPhase.BET }))).toBeNull();
    expect(getBanluckHint(state({ phase: BanLuckPhase.ROUND_END }))).toBeNull();
    expect(getBanluckHint(state({ isHumanTurn: false }))).toBeNull();
  });

  // **義務は戦略より先。** 押せないボタンを薦めないため。
  it('親の義務があるときは必ず引くと言う', () => {
    const hint = getBanluckHint(state({ mustHit: true, seats: [seat([card(10), card(4)], 14)] }));
    expect(hint?.targetAction).toBe('hit');
    expect(hint?.reason).toBe('frontendHint.banLuckBankerMustHit');
  });

  // **17 でも引くほうが得な場面がある。** ファイブドラゴンは合計に関係なく勝つ。
  it('4枚21以下ならファイブドラゴンを狙う', () => {
    const hint = getBanluckHint(state({ seats: [seat([card(2), card(3), card(4), card(6)], 15)] }));
    expect(hint?.targetAction).toBe('hit');
    expect(hint?.reason).toBe('frontendHint.banLuckChaseFiveDragon');
  });

  it('4枚でも16以上なら狙わない', () => {
    const hint = getBanluckHint(state({ seats: [seat([card(5), card(5), card(5), card(2)], 17)] }));
    expect(hint?.reason).not.toBe('frontendHint.banLuckChaseFiveDragon');
    expect(hint?.targetAction).toBe('stand');
  });

  it('5枚あるならもう引けない', () => {
    const hint = getBanluckHint(state({ seats: [seat([card(2), card(2), card(2), card(2), card(3)], 11)] }));
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.reason).toBe('frontendHint.banLuckHandFull');
  });

  it('11以下は引く', () => {
    expect(getBanluckHint(state({ seats: [seat([card(5), card(6)], 11)] }))?.targetAction).toBe('hit');
  });

  it('17以上は立つ', () => {
    expect(getBanluckHint(state({ seats: [seat([card(10), card(7)], 17)] }))?.targetAction).toBe('stand');
  });

  it('12〜16は追いかける', () => {
    const hint = getBanluckHint(state({ seats: [seat([card(10), card(5)], 15)] }));
    expect(hint?.targetAction).toBe('hit');
    expect(hint?.reason).toBe('frontendHint.banLuckChaseBanker');
  });

  it('席が無ければ助言しない', () => {
    expect(getBanluckHint(state({ seats: [], turnSeat: 0 }))).toBeNull();
  });
});
