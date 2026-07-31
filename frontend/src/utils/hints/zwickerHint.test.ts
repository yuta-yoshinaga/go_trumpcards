import { describe, expect, it } from 'vitest';
import type { ZwickerResponse } from '../../types/card';
import { getZwickerHint } from './zwickerHint';

const base = {
  players: [],
  phase: 0,
  currentPlayerIdx: 0,
  stockCount: 20,
  tableCards: [],
  builds: [],
  teamScores: [0, 0],
  targetScore: 61,
  gameEndFlag: false,
  winnerTeam: -1,
  message: '',
};

const state = (reason?: string, extra?: Record<string, unknown>): ZwickerResponse =>
  ({ ...base, hint: reason ? { reason, take: false, ...extra } : undefined }) as ZwickerResponse;

describe('getZwickerHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getZwickerHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'round_end', 'not_your_turn', 'none']) {
      expect(getZwickerHint(state(`zwicker.hint.${r}`))).toBeNull();
    }
  });

  it('sends each suggestion to its own control', () => {
    // 取ると置くは別のボタン。同じ action に潰すと、押せない場所が光る。
    expect(getZwickerHint(state('zwicker.hint.take', { take: true, cardIndex: 1 }))?.targetAction).toBe('take');
    expect(getZwickerHint(state('zwicker.hint.trail', { cardIndex: 2 }))?.targetAction).toBe('discard');
  });

  it('carries the reason key', () => {
    expect(getZwickerHint(state('zwicker.hint.take', { take: true, cardIndex: 1 }))).toEqual({
      targetAction: 'take',
      reason: 'hint.take',
      confidence: 'moderate',
    });
  });
});
