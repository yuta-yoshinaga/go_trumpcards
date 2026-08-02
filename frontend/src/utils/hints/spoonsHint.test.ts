import { describe, expect, it } from 'vitest';
import type { Card, SpoonsResponse } from '../../types/card';
import { SpoonsPhase } from '../../types/phases';
import { getSpoonsHint } from './spoonsHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[]; hasSpoon?: boolean };

function base({ hand = [card('SPADE', 5)], hasSpoon = false, ...overrides }: Partial<SpoonsResponse> & Extra = {}) {
  return {
    phase: SpoonsPhase.PASS,
    gameEndFlag: false,
    winnerIdx: -1,
    currentPlayerIdx: 0,
    feederIdx: 0,
    isHumanTurn: true,
    spoonsRemaining: 3,
    grabWindowOpen: false,
    firstGrabberIdx: -1,
    roundLoserIdx: -1,
    roundNumber: 1,
    drawPileSize: 20,
    players: [
      { name: 'あなた', isHuman: true, handSize: hand.length, hand, letters: 0, eliminated: false, hasSpoon },
      { name: 'CPU1', isHuman: false, handSize: 4, hand: [], letters: 0, eliminated: false, hasSpoon: false },
    ],
    cpuDifficulty: 1,
    message: '',
    ...overrides,
  } as SpoonsResponse;
}

describe('getSpoonsHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getSpoonsHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getSpoonsHint(base({ phase: SpoonsPhase.ROUND_END }))).toBeNull();
  });

  // **掴む窓が開いたら理屈より速さ。**遅れた 1 人が文字を負う。
  it('grabs while the window is open', () => {
    const s = base({ phase: SpoonsPhase.GRAB, grabWindowOpen: true });
    expect(getSpoonsHint(s)).toEqual({
      targetAction: 'grab',
      reason: 'frontendHint.spoonsGrabNow',
      confidence: 'strong',
    });
  });

  it('says nothing once the player already holds a spoon', () => {
    const s = base({ phase: SpoonsPhase.GRAB, grabWindowOpen: true, hasSpoon: true });
    expect(getSpoonsHint(s)).toBeNull();
  });

  // 窓が閉じている GRAB フェーズ（全部取られた後）は掴めない。
  it('says nothing when the window is closed', () => {
    expect(getSpoonsHint(base({ phase: SpoonsPhase.GRAB }))).toBeNull();
  });

  // **4 枚揃ったら自分から始める。**待つ理由がない。
  it('starts the race on four of a kind', () => {
    const hand = [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('DIAMOND', 7)];
    expect(getSpoonsHint(base({ hand }))).toEqual({
      targetAction: 'grab',
      reason: 'frontendHint.spoonsFourOfAKind',
      confidence: 'strong',
    });
  });

  it('names the odd card out to pass on', () => {
    const hand = [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5), card('DIAMOND', 9)];
    expect(getSpoonsHint(base({ hand }))).toEqual({
      targetAction: 'hand-3',
      reason: 'frontendHint.spoonsPassOdd',
      confidence: 'moderate',
    });
  });

  // 全部同じランクだが 4 枚に満たない（配り途中）。取りに行ける状態ではない。
  it('does not start the race before the hand is full', () => {
    const hand = [card('SPADE', 7), card('HEART', 7)];
    expect(getSpoonsHint(base({ hand }))).toBeNull();
  });

  it('stays quiet when it is not the human turn', () => {
    expect(getSpoonsHint(base({ isHumanTurn: false }))).toBeNull();
  });

  it('stays quiet without a visible hand', () => {
    expect(getSpoonsHint(base({ hand: [] }))).toBeNull();
  });

  it('stays quiet for an eliminated player', () => {
    const s = base();
    s.players[0].eliminated = true;
    expect(getSpoonsHint(s)).toBeNull();
  });
});
