import { describe, expect, it } from 'vitest';
import type { Card, KingoResponse } from '../../../types/card';
import { KingoPhase } from '../../../types/phases';
import { formatKingoState } from './kingoFormatter';

const card = (value: number): Card => ({ design: 'SPADE', value });

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

const base = {
  phase: KingoPhase.BET,
  seats: [seat({ bet: 20 }), seat({ name: 'CPU1', isHuman: false, isBanker: true })],
  bankerSeat: 1,
  roundNumber: 3,
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
} as KingoResponse;

const withState = (over: Partial<KingoResponse>): KingoResponse => ({ ...base, ...over });

describe('formatKingoState', () => {
  it('フェーズ・ラウンド・親を出す', () => {
    const out = formatKingoState(base);
    expect(out).toContain('Phase: BET');
    expect(out).toContain('Round: 3 of 10');
    expect(out).toContain('Banker: CPU1');
  });

  // **配当はサーバから来る。** 画面が 3 倍を持たない。
  it('サーバが送った配当倍率を出す', () => {
    expect(formatKingoState(base)).toContain('Arashi pays 3x, a pair pays 1x');
    expect(formatKingoState(withState({ payoutArashi: 7, payoutPair: 2 }))).toContain('Arashi pays 7x, a pair pays 2x');
  });

  // **親は張らない。** 額の欄は空欄。
  it('親の張り欄を空欄にする', () => {
    const out = formatKingoState(base);
    // 見出しの "Banker: CPU1" ではなく、席の行を掴む。
    const bankerLine = out.split('\n').find((l) => l.includes('CPU1') && l.includes('chips'));
    expect(bankerLine).toBeDefined();
    expect(bankerLine).toContain('bet -');
    // 子の行には額が出ている ── 空欄は親だけ。
    const childLine = out.split('\n').find((l) => l.includes('YOU') && l.includes('chips'));
    expect(childLine).toContain('bet 20');
  });

  it('配る前は手札を出さず、決着後は出す', () => {
    expect(formatKingoState(base)).not.toContain('no combination');

    const out = formatKingoState(
      withState({
        phase: KingoPhase.RESULT,
        seats: [
          seat({ bet: 20, cards: [card(3), card(3), card(8)], rank: 1, wonAmount: 20 }),
          seat({ name: 'CPU1', isHuman: false, isBanker: true, cards: [card(1), card(5), card(9)] }),
        ],
      }),
    );
    expect(out).toContain('pair');
    expect(out).toContain('no combination');
    expect(out).toContain('YOU: 20');
  });

  it('親と子で促す操作を変える', () => {
    expect(formatKingoState(base)).toContain('Place a bet');
    expect(formatKingoState(withState({ isHumanBanker: true }))).toContain('deal to continue');
  });

  it('勝者を出す', () => {
    const out = formatKingoState(withState({ phase: KingoPhase.GAME_END, gameEndFlag: true, winnerSeat: 1 }));
    expect(out).toContain('Winner: CPU1');
  });
});
