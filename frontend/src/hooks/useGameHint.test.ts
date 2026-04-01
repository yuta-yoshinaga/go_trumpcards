import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { BJ_SUGGEST_HIT, BJ_SUGGEST_NONE } from '../components/blackjack/bjConstants';
import type {
  BaccaratResponse,
  BlackJackResponse,
  CrazyEightsResponse,
  DaifugoResponse,
  DoubtResponse,
  EuchreResponse,
  NapoleonResponse,
  OhHellResponse,
  OldMaidResponse,
  SevensResponse,
  SpeedResponse,
  ThreeCardResponse,
} from '../types/card';
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

  it('returns euchre hint when enabled', () => {
    localStorage.setItem('hint_enabled_euchre', 'true');
    const state: Partial<EuchreResponse> = {
      phase: 3, // EuchrePhase.PLAY
      currentPlayerIdx: 0,
      players: [
        { id: 0, isHuman: true, cardCount: 1, cards: [{ design: 'SPADE', value: 14 }], team: 0, trickCount: 0 },
        { id: 1, isHuman: false, cardCount: 1, cards: [], team: 1, trickCount: 0 },
      ],
      trumpSuit: 1,
      currentTrick: [],
      message: '',
    };
    const { result } = renderHook(() => useGameHint('euchre', state as EuchreResponse));
    expect(result.current.hint).not.toBeNull();
  });

  it('returns napoleon hint when enabled', () => {
    localStorage.setItem('hint_enabled_napoleon', 'true');
    const state: Partial<NapoleonResponse> = {
      phase: 3, // NapoleonPhase.PLAY
      currentPlayerIdx: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 14 }],
          bid: -1,
          isNapoleon: false,
          isAdjutant: false,
          adjutantRevealed: false,
          pictureCards: 0,
          roundScore: 0,
          cumulativeScore: 0,
          trickCount: 0,
        },
        {
          id: 1,
          isHuman: false,
          cardCount: 1,
          cards: [],
          bid: -1,
          isNapoleon: false,
          isAdjutant: false,
          adjutantRevealed: false,
          pictureCards: 0,
          roundScore: 0,
          cumulativeScore: 0,
          trickCount: 0,
        },
      ],
      trumpSuit: 1,
      currentTrick: [],
      message: '',
    };
    const { result } = renderHook(() => useGameHint('napoleon', state as NapoleonResponse));
    expect(result.current.hint).not.toBeNull();
  });

  it('returns ohhell hint when enabled', () => {
    localStorage.setItem('hint_enabled_ohhell', 'true');
    const state: Partial<OhHellResponse> = {
      phase: 0, // OhHellPhase.BID
      bidPlayerIdx: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 14 }],
          bid: -1,
          roundScore: 0,
          cumulativeScore: 0,
          trickCount: 0,
        },
        {
          id: 1,
          isHuman: false,
          cardCount: 1,
          cards: [],
          bid: -1,
          roundScore: 0,
          cumulativeScore: 0,
          trickCount: 0,
        },
      ],
      trumpSuit: 1,
      restrictedBid: -1,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('ohhell', state as OhHellResponse));
    expect(result.current.hint).not.toBeNull();
  });

  it('returns oldmaid hint when enabled', () => {
    localStorage.setItem('hint_enabled_oldmaid', 'true');
    const state: Partial<OldMaidResponse> = {
      players: [
        { id: 0, isHuman: true, isFinished: false, cardCount: 3, cards: [] },
        { id: 1, isHuman: false, isFinished: false, cardCount: 5, cards: [] },
      ],
      currentTurn: 0,
      nextDrawTargetIdx: 1,
      gameEndFlag: false,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('oldmaid', state as OldMaidResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('draw');
  });

  it('returns doubt hint when enabled', () => {
    localStorage.setItem('hint_enabled_doubt', 'true');
    const state: Partial<DoubtResponse> = {
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 3,
          cards: [
            { design: 'HEART', value: 3 },
            { design: 'SPADE', value: 3 },
            { design: 'DIAMOND', value: 7 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 3, cards: [] },
      ],
      currentTurn: 0,
      phase: 0,
      gameEndFlag: false,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('doubt', state as DoubtResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('play');
  });

  it('returns crazyeights hint when enabled', () => {
    localStorage.setItem('hint_enabled_crazyeights', 'true');
    const state: Partial<CrazyEightsResponse> = {
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'SPADE', value: 10 },
          ],
          roundScore: 0,
          cumulativeScore: 0,
        },
      ],
      phase: 0,
      currentPlayerIdx: 0,
      discardTop: { design: 'HEART', value: 7 },
      chosenSuit: 0,
      gameEndFlag: false,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('crazyeights', state as CrazyEightsResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.reason).toBe('hint.playMatchingSuit');
  });

  it('returns sevens hint when enabled', () => {
    localStorage.setItem('hint_enabled_sevens', 'true');
    const state: Partial<SevensResponse> = {
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          rank: 0,
          cardCount: 1,
          passesUsed: 0,
          maxPasses: 3,
          cards: [{ design: 'HEART', value: 6 }],
          lastPlayedJoker: false,
        },
      ],
      currentTurn: 0,
      tableMinVals: [0, 7, 7, 7, 7],
      tableMaxVals: [0, 7, 7, 7, 7],
      gameEndFlag: false,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('sevens', state as SevensResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.reason).toBe('hint.playExtend');
  });

  it('returns daifugo hint when enabled', () => {
    localStorage.setItem('hint_enabled_daifugo', 'true');
    const state: Partial<DaifugoResponse> = {
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          rank: 0,
          cardCount: 2,
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'SPADE', value: 8 },
          ],
        },
      ],
      currentTurn: 0,
      tableCards: [],
      gameEndFlag: false,
      pendingAction: 'none',
      revolutionActive: false,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('daifugo', state as DaifugoResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.reason).toBe('hint.playLowest');
  });

  it('returns speed hint when enabled', () => {
    localStorage.setItem('hint_enabled_speed', 'true');
    const state: Partial<SpeedResponse> = {
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'SPADE', value: 8 },
          ],
          drawPileSize: 10,
        },
      ],
      centerPiles: [
        { design: 'HEART', value: 6 },
        { design: 'SPADE', value: 3 },
      ],
      phase: 0,
      gameEndFlag: false,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('speed', state as SpeedResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.reason).toBe('hint.hasPlayable');
  });
});
