import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { BJ_SUGGEST_HIT, BJ_SUGGEST_NONE } from '../components/blackjack/bjConstants';
import type {
  AccordionResponse,
  BaccaratResponse,
  BlackJackResponse,
  CanastaResponse,
  CaribbeanStudResponse,
  CrazyEightsResponse,
  CribbageResponse,
  DaifugoResponse,
  DoubtResponse,
  DurakResponse,
  EuchreResponse,
  GinRummyResponse,
  GoFishResponse,
  NapoleonResponse,
  OhHellResponse,
  OldMaidResponse,
  PinochleResponse,
  ScorpionResponse,
  SevensResponse,
  SpeedResponse,
  ThreeCardResponse,
  TrashResponse,
  TwoTenJackResponse,
} from '../types/card';
import { CanastaPhase, CaribbeanStudPhase, GoFishPhase } from '../types/phases';
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

  it('returns null hint for bakersdozen (server-side hints only)', () => {
    localStorage.setItem('hint_enabled_bakersdozen', 'true');
    const { result } = renderHook(() => useGameHint('bakersdozen', { phase: 0 } as never));
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

  it('returns ginrummy hint when enabled', () => {
    localStorage.setItem('hint_enabled_ginrummy', 'true');
    const state: Partial<GinRummyResponse> = {
      phase: 1, // DISCARD
      currentPlayerIdx: 0,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 11,
          cards: [
            { design: 'HEART', value: 2 },
            { design: 'SPADE', value: 5 },
            { design: 'CLOVER', value: 8 },
            { design: 'DIAMOND', value: 11 },
            { design: 'HEART', value: 13 },
            { design: 'SPADE', value: 9 },
            { design: 'CLOVER', value: 4 },
            { design: 'DIAMOND', value: 6 },
            { design: 'HEART', value: 1 },
            { design: 'SPADE', value: 12 },
            { design: 'DIAMOND', value: 10 },
          ],
          roundScore: 0,
          cumulativeScore: 0,
        },
        { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
      discardTop: null,
      drawPileCount: 20,
      knockerMelds: [],
      knockerDeadwood: [],
      isGin: false,
      gameEndFlag: false,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 100 },
    };
    const { result } = renderHook(() => useGameHint('ginrummy', state as GinRummyResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('discard');
  });

  it('returns gofish hint when enabled', () => {
    localStorage.setItem('hint_enabled_gofish', 'true');
    const state: GoFishResponse = {
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 2,
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'SPADE', value: 5 },
          ],
          bookCount: 0,
          books: [],
        },
        { id: 1, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
      ],
      phase: GoFishPhase.PLAY,
      currentTurn: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      turnNumber: 0,
      deckRemaining: 20,
      lastAsk: null,
      cpuActions: [],
      humanAction: null,
      message: '',
      config: { cpuDifficulty: 0 },
    };
    const { result } = renderHook(() => useGameHint('gofish', state));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('ask');
  });

  it('returns caribbeanstud hint when enabled', () => {
    localStorage.setItem('hint_enabled_caribbeanstud', 'true');
    const state: Partial<CaribbeanStudResponse> = {
      phase: CaribbeanStudPhase.ACTION,
      playerHand: [
        { design: 'HEART', value: 1 },
        { design: 'SPADE', value: 13 },
        { design: 'DIAMOND', value: 7 },
        { design: 'CLOVER', value: 4 },
        { design: 'HEART', value: 3 },
      ],
      playerHandRank: 0,
      dealerHand: [],
      chips: 1000,
      anteBet: 100,
      jackpotBet: 0,
      playBet: 0,
      result: 0,
      antePayout: 0,
      playPayout: 0,
      jackpotPayout: 0,
      totalPayout: 0,
      dealerQualified: false,
      dealerHandRank: 0,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('caribbeanstud', state as CaribbeanStudResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('play');
  });

  it('returns durak hint when enabled', () => {
    localStorage.setItem('hint_enabled_durak', 'true');
    const state: Partial<DurakResponse> = {
      players: [
        {
          id: 0,
          isHuman: true,
          isFinished: false,
          cardCount: 2,
          cards: [
            { design: 'SPADE', value: 6 },
            { design: 'CLOVER', value: 9 },
          ],
        },
        { id: 1, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
      ],
      phase: 0,
      attackerIdx: 0,
      defenderIdx: 1,
      tablePairs: [],
      trumpSuit: 'H',
      trumpCard: { design: 'HEART', value: 10 },
      stockCount: 12,
      loserIdx: -1,
      gameEndFlag: false,
      cpuActions: [],
      humanAction: null,
      boutNumber: 1,
      sortMode: 0,
      currentTurn: 0,
      message: '',
      config: { playerCount: 2, cpuDifficulty: 0, transferEnabled: false },
    };
    const { result } = renderHook(() => useGameHint('durak', state as DurakResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('attack');
  });

  it('returns canasta hint when enabled', () => {
    localStorage.setItem('hint_enabled_canasta', 'true');
    const humanPlayer = {
      id: 0,
      isHuman: true,
      cardCount: 11,
      cards: [],
      melds: [],
      red3Count: 0,
      red3s: [],
      roundScore: 0,
      cumulativeScore: 0,
      hasCanasta: false,
      hasInitMeld: false,
    };
    const cpuPlayer = { ...humanPlayer, id: 1, isHuman: false };
    const state: Partial<CanastaResponse> = {
      players: [humanPlayer, cpuPlayer],
      phase: CanastaPhase.DRAW,
      roundNumber: 1,
      currentPlayerIdx: 0,
      discardTop: null,
      drawPileCount: 40,
      discardPileCount: 0,
      isFrozen: false,
      gameEndFlag: false,
      winnerIdx: -1,
      message: '',
      config: { cpuDifficulty: 0, pointLimit: 5000 },
    };
    const { result } = renderHook(() => useGameHint('canasta', state as CanastaResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('draw');
  });

  it('returns pinochle hint when enabled', () => {
    localStorage.setItem('hint_enabled_pinochle', 'true');
    const state: Partial<PinochleResponse> = {
      gameEndFlag: false,
      hint: { reason: 'hint_play', cardIndex: 0 },
      players: [],
      phase: 0,
      roundNumber: 1,
      trickNumber: 1,
      currentPlayerIdx: 0,
      bidPlayerIdx: 0,
      dealerIdx: 0,
      trumpSuit: 0,
      highestBid: 0,
      highestBidder: 0,
      currentTrick: [],
      teamScores: [0, 0],
      winnerTeam: -1,
      leadPlayerIdx: 0,
      playerMelds: [],
      message: '',
      config: { cpuDifficulty: 0, pointLimit: 1500 },
    };
    const { result } = renderHook(() => useGameHint('pinochle', state as PinochleResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('play');
  });

  it('returns twotenjack hint when enabled', () => {
    localStorage.setItem('hint_enabled_twotenjack', 'true');
    const state: Partial<TwoTenJackResponse> = {
      gameEndFlag: false,
      hint: { reason: 'lead', cardIndex: 0 },
      players: [],
      phase: 0,
      roundNumber: 1,
      trickNumber: 1,
      currentPlayerIdx: 0,
      declarerIdx: 0,
      trumpSuit: 0,
      currentTrick: [],
      winnerTeam: -1,
      leadPlayerIdx: 0,
      message: '',
      config: { cpuDifficulty: 0, pointLimit: 500 },
    };
    const { result } = renderHook(() => useGameHint('twotenjack', state as TwoTenJackResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.targetAction).toBe('play');
  });

  it('returns cribbage hint when enabled', () => {
    localStorage.setItem('hint_enabled_cribbage', 'true');
    const state: Partial<CribbageResponse> = {
      phase: 2, // PEGGING
      currentPlayerIdx: 0,
      dealerIdx: 1,
      pegCount: 10,
      pegPlayedCards: [],
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 4,
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'SPADE', value: 3 },
            { design: 'CLOVER', value: 7 },
            { design: 'DIAMOND', value: 9 },
          ],
          roundScore: 0,
          cumulativeScore: 0,
        },
        { id: 1, isHuman: false, cardCount: 4, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
      crib: [],
      starter: null,
      showPhaseStep: 0,
      handScoreDetails: [null, null],
      gameEndFlag: false,
      winnerIdx: -1,
      message: '',
      config: { cpuDifficulty: 1, pointLimit: 121 },
    };
    const { result } = renderHook(() => useGameHint('cribbage', state as CribbageResponse));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.reason).toBe('hint.pegFifteen');
  });

  it('routes scorpion through getScorpionHint (tableau move)', () => {
    localStorage.setItem('hint_enabled_scorpion', 'true');
    const state: ScorpionResponse = {
      tableau: [],
      stockCount: 0,
      completedSuits: 0,
      phase: 0,
      moveCount: 0,
      canUndo: false,
      isStalemate: false,
      message: '',
      hint: { fromCol: 1, cardIndex: 2, toCol: 3 },
    };
    const { result } = renderHook(() => useGameHint('scorpion', state));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.reason).toBe('frontendHint.moveToTableau');
  });

  it('routes scorpion through getScorpionHint (deal stock)', () => {
    localStorage.setItem('hint_enabled_scorpion', 'true');
    const state: ScorpionResponse = {
      tableau: [],
      stockCount: 3,
      completedSuits: 0,
      phase: 0,
      moveCount: 0,
      canUndo: false,
      isStalemate: false,
      message: '',
      hint: { fromCol: -1, cardIndex: -1, toCol: -1 },
    };
    const { result } = renderHook(() => useGameHint('scorpion', state));
    expect(result.current.hint?.reason).toBe('frontendHint.dealStock');
  });

  it('routes accordion through getAccordionHint', () => {
    localStorage.setItem('hint_enabled_accordion', 'true');
    const state: AccordionResponse = {
      piles: [],
      pileCount: 0,
      phase: 0,
      moveCount: 0,
      canUndo: false,
      isStalemate: false,
      message: '',
      hint: { fromIdx: 3, toIdx: 0 },
    };
    const { result } = renderHook(() => useGameHint('accordion', state));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.hint?.reason).toBe('frontendHint.accordionOffset3');
  });

  it('routes trash through getTrashHint (always null — decisional hints n/a)', () => {
    localStorage.setItem('hint_enabled_trash', 'true');
    const state: TrashResponse = {
      phase: 0,
      current: 0,
      players: [
        { slots: Array.from({ length: 10 }, () => ({ faceUp: false })), isCpu: false },
        { slots: Array.from({ length: 10 }, () => ({ faceUp: false })), isCpu: true },
      ],
      stockSize: 34,
      discardSize: 0,
      moveCount: 0,
      winner: -1,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('trash', state));
    expect(result.current.hint).toBeNull();
  });

  it('routes slapjack through getSlapjackHint (slap when Jack is on top)', () => {
    localStorage.setItem('hint_enabled_slapjack', 'true');
    const state = {
      phase: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      currentTurnIdx: 0,
      isHumanTurn: true,
      isTopJack: true,
      centerPileSize: 1,
      topCard: { design: 'SPADE' as const, value: 11 },
      players: [
        { name: 'You', isHuman: true, stockSize: 25 },
        { name: 'CPU', isHuman: false, stockSize: 26 },
      ],
      cpuDifficulty: 1,
      pendingKind: 0,
      pendingDeadlineMs: 0,
      lastEventKind: 0,
      lastEventPlayerIdx: 0,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('slapjack', state));
    expect(result.current.hint?.targetAction).toBe('slap');
    expect(result.current.hint?.reason).toBe('hint.slapJack');
  });

  it('routes slapjack through getSlapjackHint (step on human turn, non-Jack)', () => {
    localStorage.setItem('hint_enabled_slapjack', 'true');
    const state = {
      phase: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      currentTurnIdx: 0,
      isHumanTurn: true,
      isTopJack: false,
      centerPileSize: 0,
      topCard: null,
      players: [
        { name: 'You', isHuman: true, stockSize: 26 },
        { name: 'CPU', isHuman: false, stockSize: 26 },
      ],
      cpuDifficulty: 1,
      pendingKind: 0,
      pendingDeadlineMs: 0,
      lastEventKind: 0,
      lastEventPlayerIdx: 0,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('slapjack', state));
    expect(result.current.hint?.targetAction).toBe('step');
    expect(result.current.hint?.reason).toBe('hint.flipCard');
  });

  it('returns null slapjack hint when game has ended', () => {
    localStorage.setItem('hint_enabled_slapjack', 'true');
    const state = {
      phase: 1,
      gameEndFlag: true,
      winnerIdx: 0,
      currentTurnIdx: 0,
      isHumanTurn: false,
      isTopJack: false,
      centerPileSize: 0,
      topCard: null,
      players: [
        { name: 'You', isHuman: true, stockSize: 52 },
        { name: 'CPU', isHuman: false, stockSize: 0 },
      ],
      cpuDifficulty: 1,
      pendingKind: 0,
      pendingDeadlineMs: 0,
      lastEventKind: 0,
      lastEventPlayerIdx: 0,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('slapjack', state));
    expect(result.current.hint).toBeNull();
  });

  it('routes egyptianratscrew through getEgyptianRatscrewHint (slap when slappable)', () => {
    localStorage.setItem('hint_enabled_egyptianratscrew', 'true');
    const state = {
      phase: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      currentTurnIdx: 1,
      isHumanTurn: false,
      isTopFaceCard: false,
      isSlappable: true,
      centerPileSize: 2,
      topCard: { design: 'SPADE' as const, value: 7 },
      players: [
        { name: 'You', isHuman: true, stockSize: 25 },
        { name: 'CPU', isHuman: false, stockSize: 25 },
      ],
      cpuDifficulty: 1,
      chanceRemaining: 0,
      chanceFromIdx: -1,
      pendingKind: 0,
      pendingDeadlineMs: 0,
      lastEventKind: 0,
      lastEventPlayerIdx: 0,
      lastSlapReason: 0,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('egyptianratscrew', state));
    expect(result.current.hint?.targetAction).toBe('slap');
    expect(result.current.hint?.reason).toBe('hint.slap');
  });

  it('routes egyptianratscrew through getEgyptianRatscrewHint (step on human turn, no slap)', () => {
    localStorage.setItem('hint_enabled_egyptianratscrew', 'true');
    const state = {
      phase: 0,
      gameEndFlag: false,
      winnerIdx: -1,
      currentTurnIdx: 0,
      isHumanTurn: true,
      isTopFaceCard: false,
      isSlappable: false,
      centerPileSize: 0,
      topCard: null,
      players: [
        { name: 'You', isHuman: true, stockSize: 26 },
        { name: 'CPU', isHuman: false, stockSize: 26 },
      ],
      cpuDifficulty: 1,
      chanceRemaining: 0,
      chanceFromIdx: -1,
      pendingKind: 0,
      pendingDeadlineMs: 0,
      lastEventKind: 0,
      lastEventPlayerIdx: 0,
      lastSlapReason: 0,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('egyptianratscrew', state));
    expect(result.current.hint?.targetAction).toBe('step');
    expect(result.current.hint?.reason).toBe('hint.flipCard');
  });

  it('returns null egyptianratscrew hint when game has ended', () => {
    localStorage.setItem('hint_enabled_egyptianratscrew', 'true');
    const state = {
      phase: 1,
      gameEndFlag: true,
      winnerIdx: 0,
      currentTurnIdx: 0,
      isHumanTurn: false,
      isTopFaceCard: false,
      isSlappable: false,
      centerPileSize: 0,
      topCard: null,
      players: [
        { name: 'You', isHuman: true, stockSize: 52 },
        { name: 'CPU', isHuman: false, stockSize: 0 },
      ],
      cpuDifficulty: 1,
      chanceRemaining: 0,
      chanceFromIdx: -1,
      pendingKind: 0,
      pendingDeadlineMs: 0,
      lastEventKind: 0,
      lastEventPlayerIdx: 0,
      lastSlapReason: 0,
      message: '',
    };
    const { result } = renderHook(() => useGameHint('egyptianratscrew', state));
    expect(result.current.hint).toBeNull();
  });
});
