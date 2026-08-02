import { describe, expect, it } from 'vitest';
import type { Card, MaoResponse } from '../../types/card';
import { MaoPhase } from '../../types/phases';
import { getMaoHint } from './maoHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function base(overrides: Partial<MaoResponse> = {}, hand: Card[] = [card('HEART', 3)]): MaoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: hand.length,
        cards: hand,
        roundScore: 0,
        cumulativeScore: 0,
        hasDeclared: false,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
    ],
    phase: MaoPhase.PLAY,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 9),
    drawPileCount: 20,
    chosenSuit: -1,
    penaltyDrawCount: 0,
    direction: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    awaitingWord: false,
    correctCount: 0,
    hintUnlocked: false,
    ruleHint: '',
    rulePenalty: false,
    message: '',
    config: { cpuDifficulty: 1, playerCnt: 4 },
    ...overrides,
  } as MaoResponse;
}

describe('getMaoHint', () => {
  it('stays quiet when it is not the human turn', () => {
    expect(getMaoHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet once the game is over', () => {
    expect(getMaoHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('recommends a suit match', () => {
    expect(getMaoHint(base({}, [card('HEART', 3), card('SPADE', 2)]))).toEqual({
      targetAction: 'play',
      reason: 'frontendHint.maoMatchSuit',
      confidence: 'strong',
    });
  });

  it('recommends a rank match when no suit matches', () => {
    expect(getMaoHint(base({}, [card('SPADE', 9), card('CLOVER', 2)]))?.reason).toBe('frontendHint.maoMatchRank');
  });

  it('keeps the wild back while an ordinary card is playable', () => {
    expect(getMaoHint(base({}, [card('HEART', 3), card('SPADE', 8)]))?.reason).toBe('frontendHint.maoSaveWild');
  });

  it('falls back to the wild when nothing else is legal', () => {
    expect(getMaoHint(base({}, [card('SPADE', 2), card('CLOVER', 8)]))?.reason).toBe('frontendHint.maoPlayWild');
  });

  it('tells the player to draw when nothing is legal', () => {
    expect(getMaoHint(base({}, [card('SPADE', 2), card('CLOVER', 5)]))).toEqual({
      targetAction: 'draw',
      reason: 'frontendHint.maoDraw',
      confidence: 'moderate',
    });
  });

  // **8 の後は選ばれたスートだけが通る。**表の札のランクに合わせても出せない。
  // ここを見落とすと、ヒントが**規則違反の手を勧める**ことになる。
  it('ignores the rank of the top card once a suit has been called', () => {
    const state = base({ chosenSuit: 4, discardTop: card('HEART', 9) }, [card('SPADE', 9), card('CLOVER', 2)]);
    expect(getMaoHint(state)?.reason).toBe('frontendHint.maoDraw');
  });

  it('follows the called suit', () => {
    const state = base({ chosenSuit: 4, discardTop: card('HEART', 9) }, [card('DIAMOND', 2), card('SPADE', 3)]);
    expect(getMaoHint(state)?.reason).toBe('frontendHint.maoMatchSuit');
  });

  it('names the suit to call after playing a wild', () => {
    const state = base({ phase: MaoPhase.CHOOSE_SUIT }, [card('SPADE', 2), card('SPADE', 5), card('HEART', 4)]);
    expect(getMaoHint(state)).toEqual({
      targetAction: 'chooseSuit',
      reason: 'frontendHint.maoChooseSuit',
      confidence: 'strong',
    });
  });

  // **隠しルールには触れない。**Mao は規則を当てる遊びで、ヒントがそれを
  // 埋めてしまうとゲームそのものが消える。宣言の言葉は解放済みの ruleHint
  // だけが示すもので、ここが示すものではない。
  it('says nothing during the declaration phase', () => {
    expect(getMaoHint(base({ phase: MaoPhase.MUST_DECLARE }))).toBeNull();
    expect(getMaoHint(base({ awaitingWord: true, hintUnlocked: true, ruleHint: 'x' }))).toBeNull();
  });

  it('stays quiet with an empty hand', () => {
    expect(getMaoHint(base({}, []))).toBeNull();
  });
});
