import { describe, expect, it } from 'vitest';
import type { Card, OasisPokerResponse } from '../../types/card';
import { OasisPokerPhase } from '../../types/phases';
import { getOasisPokerHint } from './oasispokerHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<OasisPokerResponse> = {}): OasisPokerResponse {
  return {
    playerHand: [card('SPADE', 2), card('HEART', 5), card('DIAMOND', 8), card('CLOVER', 10), card('SPADE', 12)],
    dealerHand: [],
    phase: OasisPokerPhase.ACTION,
    chips: 1000,
    anteBet: 100,
    jackpotBet: 0,
    exchangeCount: 0,
    exchangeFee: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    jackpotPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getOasisPokerHint', () => {
  it('returns null in bet and end phases', () => {
    expect(getOasisPokerHint(makeState({ phase: OasisPokerPhase.BET }))).toBeNull();
    expect(getOasisPokerHint(makeState({ phase: OasisPokerPhase.END }))).toBeNull();
  });

  it('returns null when player hand is empty', () => {
    expect(getOasisPokerHint(makeState({ playerHand: [] }))).toBeNull();
  });

  // --- Action phase ---
  it('recommends play with pair or better (strong)', () => {
    const hint = getOasisPokerHint(makeState({ phase: OasisPokerPhase.ACTION, playerHandRank: 1 }));
    expect(hint?.targetAction).toBe('play');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.pairOrBetter');
  });

  it('recommends play with Ace-King high (moderate)', () => {
    const state = makeState({
      phase: OasisPokerPhase.ACTION,
      playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 8), card('CLOVER', 5), card('SPADE', 3)],
      playerHandRank: 0,
    });
    const hint = getOasisPokerHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.aceKingHigh');
  });

  it('recommends fold with a weak high-card hand', () => {
    const hint = getOasisPokerHint(makeState({ phase: OasisPokerPhase.ACTION, playerHandRank: 0 }));
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.reason).toBe('hint.weakHand');
  });

  // --- Exchange phase ---
  // **ペアがあるだけでは "stand" ではない (#4711)。**この t.Run は以前
  // playerHandRank だけを見て stand を期待していたが、低いペア + くず札3枚は
  // 引き直したほうがよく、CUI はそう助言している。交換すべき札が無いときだけ
  // stand になる。
  it('stands on a hand where every card is worth keeping (strong)', () => {
    const hint = getOasisPokerHint(
      makeState({
        phase: OasisPokerPhase.EXCHANGE,
        playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 12), card('CLOVER', 11), card('SPADE', 11)],
        playerHandRank: 1,
      }),
    );
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.exchangeKeep');
  });

  it('swaps the kickers around a low pair rather than standing', () => {
    const hint = getOasisPokerHint(
      makeState({
        phase: OasisPokerPhase.EXCHANGE,
        playerHand: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 2), card('CLOVER', 8), card('SPADE', 3)],
        playerHandRank: 1,
      }),
    );
    expect(hint?.targetAction).toBe('exchange');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends exchanging a weak hand (moderate)', () => {
    const hint = getOasisPokerHint(makeState({ phase: OasisPokerPhase.EXCHANGE, playerHandRank: 0 }));
    expect(hint?.targetAction).toBe('exchange');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.exchangeImprove');
  });
});

// **CUI はどの札を交換すべきかインデックスで列挙しているのに、Web は
// 「交換すべき」としか言わなかった (#4711)。**5枚を個別にクリックして選ぶ UI が
// あるのに、どれを選ぶかの案内が無い。
describe('getOasisPokerHint — which cards to exchange (sync: oasisPokerExchangeIndices)', () => {
  it('names the weak cards to swap out', () => {
    const hint = getOasisPokerHint(
      makeState({
        phase: OasisPokerPhase.EXCHANGE,
        // ♠A ♥K は高札、♦5 ♣8 ♠3 は捨てる。
        playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 5), card('CLOVER', 8), card('SPADE', 3)],
        playerHandRank: 0,
      }),
    );
    expect(hint?.targetAction).toBe('exchange');
    expect(hint?.targetIndices).toEqual([2, 3, 4]);
  });

  // **ペアは残す。**残す札まで交換対象に入れると、できている役を壊す。
  it('keeps a pair and swaps only the rest', () => {
    const hint = getOasisPokerHint(
      makeState({
        phase: OasisPokerPhase.EXCHANGE,
        playerHand: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 2), card('CLOVER', 8), card('SPADE', 3)],
        playerHandRank: 1,
      }),
    );
    expect(hint?.targetIndices).toEqual([2, 3, 4]);
  });

  // **A/J/Q/K は残す。**エースは value 1 なので、素朴な大小比較だと捨ててしまう。
  it('keeps an ace even though its raw value is 1', () => {
    const hint = getOasisPokerHint(
      makeState({
        phase: OasisPokerPhase.EXCHANGE,
        playerHand: [card('SPADE', 1), card('HEART', 4), card('DIAMOND', 6), card('CLOVER', 8), card('SPADE', 9)],
        playerHandRank: 0,
      }),
    );
    expect(hint?.targetIndices).not.toContain(0);
  });

  // **交換するものが無いときだけ stand。**低いペアを持っているだけで
  // 「そのまま」と言うと、残り3枚を引き直す機会を捨てさせる。
  it('stands only when there is nothing worth swapping', () => {
    const hint = getOasisPokerHint(
      makeState({
        phase: OasisPokerPhase.EXCHANGE,
        playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 12), card('CLOVER', 11), card('SPADE', 11)],
        playerHandRank: 1,
      }),
    );
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.targetIndices ?? []).toEqual([]);
  });

  it('does not tag the action phase with exchange indices', () => {
    const hint = getOasisPokerHint(
      makeState({
        phase: OasisPokerPhase.ACTION,
        playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 5), card('CLOVER', 8), card('SPADE', 3)],
        playerHandRank: 0,
      }),
    );
    expect(hint?.targetIndices).toBeUndefined();
  });
});
