import { describe, expect, it } from 'vitest';
import type { PochResponse } from '../../types/card';
import { getPochHint } from './pochHint';

const base = {
  players: [],
  phase: 1,
  validPlays: [],
  currentPlayerIdx: 0,
  pools: [],
  paySuit: 1,
  stakingAwards: [],
  betTarget: 0,
  pochenWinner: -1,
  pochenPot: 0,
  playedPile: [],
  stopsSuit: -1,
  stopsRank: 0,
  dealNo: 0,
  targetDeals: 5,
  dealWinner: -1,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  yourBestComboSize: 0,
  yourBestComboRank: 0,
};

const state = (reason?: string, extra?: Record<string, unknown>): PochResponse =>
  ({ ...base, hint: reason ? { reason, action: '', ...extra } : undefined }) as PochResponse;

describe('getPochHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getPochHint(state())).toBeNull();
  });

  it('returns null for the non-suggestions', () => {
    for (const r of ['game_end', 'deal_end', 'not_your_turn', 'none']) {
      expect(getPochHint(state(`poch.hint.${r}`))).toBeNull();
    }
  });

  it('sends each suggestion to its own control', () => {
    // 賭ける・降りる・出すは別のボタン。同じ action に潰すと押せない場所が光る。
    expect(getPochHint(state('poch.hint.bet', { action: 'bet' }))?.targetAction).toBe('bet');
    expect(getPochHint(state('poch.hint.fold', { action: 'fold' }))?.targetAction).toBe('fold');
    expect(getPochHint(state('poch.hint.play', { action: 'play', cardIndex: 2 }))?.targetAction).toBe('play');
  });

  it('carries the reason key', () => {
    expect(getPochHint(state('poch.hint.play', { action: 'play', cardIndex: 2 }))).toEqual({
      targetAction: 'play',
      reason: 'hint.play',
      confidence: 'moderate',
    });
  });
});
