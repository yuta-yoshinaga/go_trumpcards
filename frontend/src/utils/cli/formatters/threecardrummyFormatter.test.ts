import { describe, expect, it } from 'vitest';
import type { Card } from '../../../types/common';
import type { ThreeCardRummyResponse } from '../../../types/games/threecardrummy';
import { ThreeCardRummyPhase } from '../../../types/phases';
import { formatThreecardrummyState } from './threecardrummyFormatter';

/** ♠K + ♥5 + ♦2 = 10 + 5 + 2 = 17 点。役ではない普通の手。 */
const PLAIN_HAND: Card[] = [
  { design: 'SPADE', value: 13 },
  { design: 'HEART', value: 5 },
  { design: 'DIAMOND', value: 2 },
];

/** 同ランク 3 枚 = 役 = 0 点。**このゲームの最強手。** */
const MELD_HAND: Card[] = [
  { design: 'SPADE', value: 7 },
  { design: 'HEART', value: 7 },
  { design: 'CLOVER', value: 7 },
];

// ♣3 ♦9 ♠A = 3 + 9 + 1 = 13。混色かつ非連番なので役ではない。20 以下なので
// クオリファイする側の手。
const DEALER_HAND: Card[] = [
  { design: 'CLOVER', value: 3 },
  { design: 'DIAMOND', value: 9 },
  { design: 'SPADE', value: 1 },
];

// ♣K ♦Q ♠5 = 10 + 10 + 5 = 25。クオリファイ上限 20 を超える手。
const UNQUALIFIED_DEALER_HAND: Card[] = [
  { design: 'CLOVER', value: 13 },
  { design: 'DIAMOND', value: 12 },
  { design: 'SPADE', value: 5 },
];

function makeState(overrides: Partial<ThreeCardRummyResponse> = {}): ThreeCardRummyResponse {
  return {
    message: '',
    playerHand: PLAIN_HAND,
    dealerHand: [],
    phase: ThreeCardRummyPhase.ACTION,
    chips: 950,
    anteBet: 0,
    lowBonusBet: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    anteBonusPayout: 0,
    lowBonusPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerScore: 17,
    dealerScore: 13,
    ...overrides,
  };
}

describe('formatThreecardrummyState', () => {
  it('renders the header, chips and phase name', () => {
    const out = formatThreecardrummyState(makeState({ phase: ThreeCardRummyPhase.BET, playerHand: [] }));
    expect(out).toContain('Three Card Rummy');
    expect(out).toContain('chips: 950');
    expect(out).toContain('phase: BET');
  });

  it.each([
    [ThreeCardRummyPhase.BET, 'BET'],
    [ThreeCardRummyPhase.ACTION, 'ACTION'],
    [ThreeCardRummyPhase.END, 'END'],
  ])('names phase %i as %s', (phase, name) => {
    expect(formatThreecardrummyState(makeState({ phase }))).toContain(`phase: ${name}`);
  });

  it('renders UNKNOWN for an out-of-range phase', () => {
    expect(formatThreecardrummyState(makeState({ phase: 99 }))).toContain('phase: UNKNOWN');
  });

  it('renders the indexed player hand', () => {
    const out = formatThreecardrummyState(makeState());
    expect(out).toContain('Your hand: [0]♠K  [1]♥5  [2]♦2');
  });

  it('omits the hand and score lines before the cards are dealt', () => {
    const out = formatThreecardrummyState(makeState({ phase: ThreeCardRummyPhase.BET, playerHand: [] }));
    expect(out).not.toContain('Your hand');
    expect(out).not.toContain('Your score');
  });

  // **点数が play/fold の判断材料そのもの。** アクションフェーズで出ていなければ
  // 決めようがない。
  it('renders the player score from the action phase, noting lower is better', () => {
    const out = formatThreecardrummyState(makeState({ phase: ThreeCardRummyPhase.ACTION }));
    expect(out).toContain('Your score: 17 (lower is better)');
  });

  // 素の "0" は「手が無い/未計算」に見える。役だと言い切らないと最強手が
  // 最弱手のように読める。
  it('renders a meld as "0 (meld)" rather than a bare 0', () => {
    // 点数はサーバが返す値をそのまま出す (手札から数え直さない)。
    const out = formatThreecardrummyState(makeState({ playerHand: MELD_HAND, playerScore: 0 }));
    expect(out).toContain('Your score: 0 (meld) (lower is better)');
    expect(out).not.toMatch(/Your score: 0 \(lower/);
  });

  it('hides the dealer entirely outside the end phase', () => {
    const out = formatThreecardrummyState(
      makeState({ phase: ThreeCardRummyPhase.ACTION, dealerHand: DEALER_HAND, dealerScore: 13 }),
    );
    expect(out).not.toContain('Dealer:');
    expect(out).not.toContain('Dealer score');
    expect(out).not.toContain('Dealer qualified');
  });

  it('shows the dealer cards, score and qualification in the end phase', () => {
    const out = formatThreecardrummyState(
      makeState({
        phase: ThreeCardRummyPhase.END,
        dealerHand: DEALER_HAND,
        dealerScore: 13,
        dealerQualified: true,
      }),
    );
    expect(out).toContain('Dealer: ♣3, ♦9, ♠A');
    expect(out).toContain('Dealer score: 13');
    expect(out).toContain('Dealer qualified: yes');
  });

  // ディーラーは 20 以下でクオリファイ。しなかったことは払い戻しの前提なので
  // 明示されないと結果が読めない。
  it('reports a non-qualifying dealer as no', () => {
    const out = formatThreecardrummyState(
      makeState({
        phase: ThreeCardRummyPhase.END,
        dealerHand: UNQUALIFIED_DEALER_HAND,
        dealerScore: 25,
        dealerQualified: false,
      }),
    );
    expect(out).toContain('Dealer score: 25');
    expect(out).toContain('Dealer qualified: no');
  });

  it('renders a dealer meld as "0 (meld)"', () => {
    const out = formatThreecardrummyState(
      makeState({ phase: ThreeCardRummyPhase.END, dealerHand: MELD_HAND, dealerScore: 0 }),
    );
    expect(out).toContain('Dealer score: 0 (meld)');
  });

  it('omits every bet line while the bets are zero', () => {
    const out = formatThreecardrummyState(makeState());
    expect(out).not.toContain('ante:');
    expect(out).not.toContain('low bonus:');
    expect(out).not.toContain('play bet:');
  });

  it('renders each bet line once it is non-zero', () => {
    const out = formatThreecardrummyState(makeState({ anteBet: 10, lowBonusBet: 5, playBet: 10 }));
    expect(out).toContain('ante: 10');
    expect(out).toContain('low bonus: 5');
    expect(out).toContain('play bet: 10');
  });

  it('shows the ante alone when the low bonus was not taken', () => {
    const out = formatThreecardrummyState(makeState({ anteBet: 10, lowBonusBet: 0 }));
    expect(out).toContain('ante: 10');
    expect(out).not.toContain('low bonus:');
  });

  it('renders the payout breakdown and total in the end phase', () => {
    const out = formatThreecardrummyState(
      makeState({
        phase: ThreeCardRummyPhase.END,
        antePayout: 10,
        playPayout: 10,
        anteBonusPayout: 5,
        lowBonusPayout: 20,
        totalPayout: 45,
      }),
    );
    expect(out).toContain('payout: ante=10 play=10 anteBonus=5 lowBonus=20');
    expect(out).toContain('total: 45');
  });

  it('omits the payout breakdown outside the end phase', () => {
    const out = formatThreecardrummyState(makeState({ phase: ThreeCardRummyPhase.ACTION, totalPayout: 45 }));
    expect(out).not.toContain('payout:');
    expect(out).not.toContain('total:');
  });

  it('appends the server message', () => {
    const out = formatThreecardrummyState(makeState({ message: 'You win with a meld!' }));
    expect(out).toContain('You win with a meld!');
  });

  it('adds no blank message line when the server sent none', () => {
    // `formatSeparator()` は無条件に最後へ積まれるので「末尾が ===== か」を
    // 見ても何も証明しない。区切りの**直前**に空行が挟まっていないかを見る。
    const withMessage = formatThreecardrummyState(makeState({ message: 'You win!' }));
    const without = formatThreecardrummyState(makeState({ message: '' }));
    expect(withMessage.split('\n').at(-2)).toBe('You win!');
    expect(without.split('\n').at(-2)).not.toBe('');
    expect(without.split('\n').length).toBe(withMessage.split('\n').length - 1);
  });
});
