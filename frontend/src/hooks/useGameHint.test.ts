import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { BJ_SUGGEST_HIT, BJ_SUGGEST_NONE } from '../components/blackjack/bjConstants';
import type { BaccaratResponse, BlackJackResponse, ThreeCardResponse } from '../types/card';
import { useGameHint } from './useGameHint';

function makeBjState(overrides: Partial<BlackJackResponse> = {}): BlackJackResponse {
  return {
    dealer: { score: 0, cards: [], chips: 1000 },
    player: { score: 0, cards: [], chips: 1000 },
    hands: [],
    currentHandIdx: 0,
    phase: 4,
    insuranceBet: 0,
    insuranceAvailable: false,
    message: '',
    hintEnabled: true,
    suggestedAction: BJ_SUGGEST_NONE,
    deckCount: 6,
    dealerHitsSoft17: false,
    countingEnabled: false,
    cpuPlayerCount: 0,
    runningCount: 0,
    trueCount: 0,
    perfectPairsBet: 0,
    twentyOnePlus3Bet: 0,
    doubleAfterSplit: false,
    countingSystem: 0,
    deckPenetration: 75,
    multiHandCount: 1,
    surrenderRule: 0,
    ...overrides,
  };
}

describe('useGameHint', () => {
  afterEach(() => {
    localStorage.clear();
  });

  it('returns hintEnabled false by default', () => {
    const { result } = renderHook(() => useGameHint('blackjack', null));
    expect(result.current.hintEnabled).toBe(false);
  });

  it('returns null hint when disabled', () => {
    const state = makeBjState({ suggestedAction: BJ_SUGGEST_HIT });
    const { result } = renderHook(() => useGameHint('blackjack', state));
    expect(result.current.hint).toBeNull();
  });

  it('returns hint when enabled and state has suggestion', () => {
    localStorage.setItem('hint_enabled_blackjack', 'true');
    const state = makeBjState({ suggestedAction: BJ_SUGGEST_HIT });
    const { result } = renderHook(() => useGameHint('blackjack', state));
    expect(result.current.hintEnabled).toBe(true);
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('hit');
  });

  it('toggles hint enabled', () => {
    const { result } = renderHook(() => useGameHint('blackjack', null));
    act(() => result.current.setHintEnabled(true));
    expect(result.current.hintEnabled).toBe(true);
    expect(localStorage.getItem('hint_enabled_blackjack')).toBe('true');
  });

  it('returns null hint when state is null', () => {
    localStorage.setItem('hint_enabled_blackjack', 'true');
    const { result } = renderHook(() => useGameHint('blackjack', null));
    expect(result.current.hint).toBeNull();
  });

  it('returns null for unknown game name', () => {
    localStorage.setItem('hint_enabled_unknown', 'true');
    // @ts-expect-error testing invalid game name
    const { result } = renderHook(() => useGameHint('unknown', {}));
    expect(result.current.hint).toBeNull();
  });

  it('returns baccarat hint when enabled', () => {
    localStorage.setItem('hint_enabled_baccarat', 'true');
    const state: Partial<BaccaratResponse> = {
      phase: 1,
      betType: 0,
      playerHand: [],
      bankerHand: [],
      playerHandValue: 0,
      bankerHandValue: 0,
      chips: 1000,
      betAmount: 100,
      result: 0,
      payout: 0,
      history: [],
      playerPairBet: 0,
      bankerPairBet: 0,
      sideBetResults: [],
      message: '',
    };
    const { result } = renderHook(() => useGameHint('baccarat', state as BaccaratResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('banker');
  });

  it('returns threecard hint when enabled', () => {
    localStorage.setItem('hint_enabled_threecard', 'true');
    const state: Partial<ThreeCardResponse> = {
      phase: 2,
      playerHandRank: 2,
      playerHand: [
        { design: 'HEART', value: 10 },
        { design: 'DIAMOND', value: 10 },
        { design: 'SPADE', value: 5 },
      ],
      dealerHand: [],
      chips: 1000,
      anteBet: 100,
      pairPlusBet: 0,
      playBet: 0,
      result: 0,
      antePayout: 0,
      playPayout: 0,
      anteBonusPayout: 0,
      pairPlusPayout: 0,
      totalPayout: 0,
      dealerQualified: false,
      dealerHandRank: 0,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('threecard', state as ThreeCardResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('play');
  });
});
