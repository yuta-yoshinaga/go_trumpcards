import { describe, expect, it } from 'vitest';
import type { Card } from '../../../types/card';
import type { BlackHoleResponse } from '../../../types/games/blackhole';
import { BlackHolePhase } from '../../../types/phases';
import { formatBlackHoleState } from './blackholeFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeMockState(overrides: Partial<BlackHoleResponse> = {}): BlackHoleResponse {
  const fans: Card[][] = Array.from({ length: 17 }, () => []);
  // fan 0: top is 2 (adjacent to ♠A which is 1 -> legal)
  fans[0] = [card('HEART', 9), card('CLOVER', 2)];
  // fan 1: top is 7 (not adjacent to ♠A -> not legal)
  fans[1] = [card('DIAMOND', 7)];
  // fan 2: empty
  fans[2] = [];
  return {
    fans,
    blackHole: [card('SPADE', 1)],
    phase: BlackHolePhase.PLAYING,
    moveCount: 3,
    canUndo: false,
    isStalemate: false,
    messageCode: 'blackhole.playing',
    message: '',
    ...overrides,
  };
}

describe('formatBlackHoleState', () => {
  it('formats initial/playing board layout (representative position 1)', () => {
    const state = makeMockState();
    const text = formatBlackHoleState(state);

    expect(text).toContain('Black Hole');
    expect(text).toContain('moves: 3  progress: 1/52  phase: Playing');
    expect(text).toContain('black hole: ♠A (accepts: 2)');
    expect(text).toContain('[ 0] ♥9 ♣2 [*]');
    expect(text).toContain('[ 1] ♦7 [x]');
    expect(text).toContain('[ 2] [empty]');
    expect(text).toContain('undo: none');
    expect(text).toContain('Available commands: p <fan> (0), h (hint), g (giveup), r (reset)');
  });

  it('suppresses hint line during normal play even if state.hint is present', () => {
    // Passive hint is returned by the server on Output(), but must NOT be shown in CLI without asking
    const state = makeMockState({
      hint: { fan: 0, movesAfter: 1, canClear: true, continuationMoves: 49 },
      messageCode: 'blackhole.playing',
    });
    const text = formatBlackHoleState(state);
    expect(text).not.toContain('HINT');
    expect(text).not.toContain('No hint available');
  });

  it('renders hint line when player explicitly requested it (blackhole.hintAvailable)', () => {
    const state = makeMockState({
      hint: { fan: 0, movesAfter: 2, canClear: true, continuationMoves: 49 },
      messageCode: 'blackhole.hintAvailable',
    });
    const text = formatBlackHoleState(state);
    expect(text).toContain('HINT: Fan 0 (♣2) [clear in 49 moves]');
  });

  it('renders moves-after in hint when canClear is false', () => {
    const state = makeMockState({
      hint: { fan: 0, movesAfter: 5, canClear: false, continuationMoves: 0 },
      messageCode: 'blackhole.hintAvailable',
    });
    const text = formatBlackHoleState(state);
    expect(text).toContain('HINT: Fan 0 (♣2) [moves after: 5]');
  });

  it('renders hint line when recommended fan is out of bounds or empty', () => {
    const state = makeMockState({
      hint: { fan: 2, movesAfter: 0, canClear: false, continuationMoves: 0 }, // fan 2 is empty
      messageCode: 'blackhole.hintAvailable',
    });
    const text = formatBlackHoleState(state);
    expect(text).toContain('HINT: Fan 2 [moves after: 0]');
  });

  it('renders "No hint available." when hint requested but no legal move (blackhole.noHint)', () => {
    const state = makeMockState({
      hint: undefined,
      messageCode: 'blackhole.noHint',
    });
    const text = formatBlackHoleState(state);
    expect(text).toContain('No hint available.');
  });

  it('formats game clear state (representative position 2)', () => {
    const fans: Card[][] = Array.from({ length: 17 }, () => []);
    const blackHole = Array.from({ length: 52 }, (_, i) => card('SPADE', (i % 13) + 1));
    const state = makeMockState({
      fans,
      blackHole,
      phase: BlackHolePhase.GAME_CLEAR,
      moveCount: 51,
      message: 'ゲームクリア！ 手数: 51',
    });
    const text = formatBlackHoleState(state);

    expect(text).toContain('phase: Game Clear');
    expect(text).toContain('Game Clear! Total moves: 51');
    expect(text).not.toContain('ゲームクリア！ 手数: 51');
    expect(text).toContain('Available commands: r (reset)');
    expect(text).not.toContain('p <fan>');
  });

  it('formats game over and stalemate states', () => {
    const overState = makeMockState({
      phase: BlackHolePhase.GAME_OVER,
      message: 'ゲームオーバー',
    });
    const overText = formatBlackHoleState(overState);
    expect(overText).toContain('phase: Game Over');
    expect(overText).toContain('Game Over!');
    expect(overText).not.toContain('ゲームオーバー');

    const staleState = makeMockState({
      phase: BlackHolePhase.PLAYING,
      isStalemate: true,
      fans: [[card('DIAMOND', 7)]], // not adjacent to ♠A
    });
    const staleText = formatBlackHoleState(staleState);
    expect(staleText).toContain('Stalemate - no legal moves remaining');
    expect(staleText).toContain('Available commands: h (hint), g (giveup), r (reset)');
  });

  it('surfaces undo availability and falls back for unknown phase or empty hole', () => {
    const state = makeMockState({
      canUndo: true,
      blackHole: [],
      phase: 99,
    });
    const text = formatBlackHoleState(state);
    expect(text).toContain('undo: available');
    expect(text).toContain('phase: Phase 99');
    expect(text).toContain('black hole: [empty] (accepts: none)');
  });

  it('handles undo command inclusion in available commands when canUndo is true and playing', () => {
    const state = makeMockState({
      canUndo: true,
      phase: BlackHolePhase.PLAYING,
    });
    const text = formatBlackHoleState(state);
    expect(text).toContain('u (undo)');
  });
});
