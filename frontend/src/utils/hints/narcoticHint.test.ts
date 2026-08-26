import { describe, expect, it } from 'vitest';
import type { NarcoticResponse } from '../../types/card';
import { getNarcoticHint } from './narcoticHint';

function makeState(overrides?: Partial<NarcoticResponse>): NarcoticResponse {
  return {
    columns: [[], [], [], []],
    stockCount: 40,
    discardCount: 0,
    redealCount: 0,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getNarcoticHint', () => {
  it('maps each backend action onto the button that performs it', () => {
    expect(getNarcoticHint(makeState({ hint: { type: 'remove', col: 2 } }))?.targetAction).toBe('remove');
    expect(getNarcoticHint(makeState({ hint: { type: 'move', col: 1 } }))?.targetAction).toBe('move');
    expect(getNarcoticHint(makeState({ hint: { type: 'draw', col: -1 } }))?.targetAction).toBe('deal');
  });

  it('rates a discard as strong and the others as moderate', () => {
    expect(getNarcoticHint(makeState({ hint: { type: 'remove', col: 0 } }))?.confidence).toBe('strong');
    expect(getNarcoticHint(makeState({ hint: { type: 'draw', col: -1 } }))?.confidence).toBe('moderate');
  });

  // **ヒントが無いときと、終局後は出さない。**どちらも null。
  it('returns null without a hint', () => {
    expect(getNarcoticHint(makeState())).toBeNull();
  });

  it('returns null once the game has ended', () => {
    expect(getNarcoticHint(makeState({ phase: 1, hint: { type: 'remove', col: 0 } }))).toBeNull();
  });

  // **知らない type は無視する。**バックエンドが新しい手を返しても、
  // 対応するボタンが無いうちは何も出さない。
  it('returns null for an action it does not know', () => {
    expect(getNarcoticHint(makeState({ hint: { type: 'teleport' as never, col: 0 } }))).toBeNull();
  });
});
