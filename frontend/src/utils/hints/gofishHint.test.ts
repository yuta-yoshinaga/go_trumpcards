import { describe, expect, it } from 'vitest';
import type { Card, GoFishConfig, GoFishResponse } from '../../types/card';
import { GoFishPhase } from '../../types/phases';
import { getGoFishHint } from './gofishHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: GoFishConfig = { cpuDifficulty: 0 };

function makeState(overrides: Partial<GoFishResponse> = {}): GoFishResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 4,
        cards: [card('HEART', 5), card('SPADE', 5), card('DIAMOND', 8), card('CLOVER', 3)],
        bookCount: 0,
        books: [],
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
      { id: 2, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
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
    config: defaultConfig,
    ...overrides,
  };
}

describe('getGoFishHint', () => {
  it('returns null when game has ended', () => {
    expect(getGoFishHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when phase is not PLAY', () => {
    expect(getGoFishHint(makeState({ phase: GoFishPhase.GAME_END, gameEndFlag: false }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getGoFishHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getGoFishHint(state)).toBeNull();
  });

  it('recommends asking with moderate confidence when no opponent info is known', () => {
    const hint = getGoFishHint(makeState());
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('ask');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.askMostCopies');
  });

  // #5518: 文言は「最も多く持っているランク」と言うのに、どのランクかは返って
  // いなかった。プレイヤーは助言を読んだあと自分で手札を数え直すことになる。
  it('names the rank it holds most copies of, and points at those cards', () => {
    const hint = getGoFishHint(makeState()); // H5 S5 D8 C3 -> 5 が2枚
    expect(hint?.reasonParams).toEqual({ rank: '5' });
    expect(hint?.targetIndices).toEqual([0, 1]);
  });

  it('labels a face rank by its letter, not its number', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 13), card('SPADE', 13), card('DIAMOND', 8)];
    const hint = getGoFishHint(state);
    expect(hint?.reasonParams).toEqual({ rank: 'K' });
    expect(hint?.targetIndices).toEqual([0, 1]);
  });

  // **同数のときの選び方を固定する。**手札の並び順に任せると、同じ盤面でも
  // 引き直すたびに違うランクを勧めることになる。
  it('breaks a tie by the lowest rank, whatever the hand order is', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 9), card('SPADE', 9), card('DIAMOND', 4), card('CLOVER', 4)];
    expect(getGoFishHint(state)?.reasonParams).toEqual({ rank: '4' });
    expect(getGoFishHint(state)?.targetIndices).toEqual([2, 3]);

    state.players[0].cards = [card('DIAMOND', 4), card('CLOVER', 4), card('HEART', 9), card('SPADE', 9)];
    expect(getGoFishHint(state)?.reasonParams).toEqual({ rank: '4' });
    expect(getGoFishHint(state)?.targetIndices).toEqual([0, 1]);
  });

  // A は 1 として持っているが、表示は "A"。数の小さい順のタイブレークでは
  // 一番若いランクなので先に選ばれる。
  it('treats the ace as the lowest rank', () => {
    const state = makeState();
    state.players[0].cards = [card('HEART', 7), card('SPADE', 7), card('DIAMOND', 1), card('CLOVER', 1)];
    expect(getGoFishHint(state)?.reasonParams).toEqual({ rank: 'A' });
    expect(getGoFishHint(state)?.targetIndices).toEqual([2, 3]);
  });

  it('upgrades confidence when a cpu has recently taken a rank the human holds', () => {
    const state = makeState({
      cpuActions: [
        {
          askPlayerIdx: 1,
          askTargetIdx: 2,
          askRank: 5,
          success: true,
          cardsReceived: 1,
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      ],
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.askKnownRank');
    // 既知のランクも同じで、見つけた値を捨てずに返す。
    expect(hint?.reasonParams).toEqual({ rank: '5' });
    expect(hint?.targetIndices).toEqual([0, 1]);
  });

  // 既知のランクが複数あるときも、手札の並び順ではなく若い順に決める。
  it('picks the lowest known rank when an opponent revealed several', () => {
    const state = makeState({
      cpuActions: [8, 3].map((askRank) => ({
        askPlayerIdx: 1,
        askTargetIdx: 2,
        askRank,
        success: true,
        cardsReceived: 1,
        drawnCard: null,
        bookFormed: false,
        bookRank: 0,
      })),
    });
    const hint = getGoFishHint(state); // 手札は H5 S5 D8 C3
    expect(hint?.reason).toBe('hint.askKnownRank');
    expect(hint?.reasonParams).toEqual({ rank: '3' });
    expect(hint?.targetIndices).toEqual([3]);
  });

  it('uses lastAsk when available', () => {
    const state = makeState({
      lastAsk: {
        playerIdx: 1,
        targetIdx: 2,
        rank: 5,
        success: true,
        cardsReceived: [card('HEART', 5)],
        drawnCard: null,
        bookFormed: false,
        bookRank: 0,
      },
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('strong');
  });

  it('ignores failed asks', () => {
    const state = makeState({
      cpuActions: [
        {
          askPlayerIdx: 1,
          askTargetIdx: 2,
          askRank: 5,
          success: false,
          cardsReceived: 0,
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      ],
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('moderate');
  });

  it('ignores lastAsk where the human was the asker', () => {
    const state = makeState({
      lastAsk: {
        playerIdx: 0, // human was the asker — collectFromLastAsk returns early
        targetIdx: 1,
        rank: 5,
        success: true,
        cardsReceived: [card('HEART', 5)],
        drawnCard: null,
        bookFormed: false,
        bookRank: 0,
      },
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('moderate');
  });

  it('ignores asks targeting the human itself', () => {
    const state = makeState({
      cpuActions: [
        {
          askPlayerIdx: 1,
          askTargetIdx: 0,
          askRank: 5,
          success: true,
          cardsReceived: 1,
          drawnCard: null,
          bookFormed: false,
          bookRank: 0,
        },
      ],
    });
    const hint = getGoFishHint(state);
    expect(hint?.confidence).toBe('moderate');
  });
});
