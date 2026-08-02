import { describe, expect, it } from 'vitest';
import type { RookResponse } from '../../types/card';
import { getRookHint } from './rookHint';

const state = (hint?: RookResponse['hint']): RookResponse => ({ hint }) as RookResponse;

describe('getRookHint', () => {
  it('returns null without a hint', () => {
    expect(getRookHint(state())).toBeNull();
  });

  it('points at the card to play', () => {
    expect(getRookHint(state({ cardIndex: 4, reason: 'lead_high' }))).toEqual({
      targetAction: 'card-4',
      reason: 'hint.lead_high',
      confidence: 'moderate',
    });
  });

  // **札 0 は正当な手。**真偽値で見ると先頭だけ落ちる。
  it('keeps a hint that points at card index 0', () => {
    expect(getRookHint(state({ cardIndex: 0, reason: 'lead_high' }))?.targetAction).toBe('card-0');
  });

  it('points at the nest discard', () => {
    expect(getRookHint(state({ discardIndices: [1, 2], reason: 'discard_low' }))?.targetAction).toBe('discard');
  });

  it('points at the trump colour', () => {
    expect(getRookHint(state({ trumpColor: 2, reason: 'name_trump' }))?.targetAction).toBe('trump');
  });

  // **色 0 も正当。**
  it('keeps a trump hint naming colour zero', () => {
    expect(getRookHint(state({ trumpColor: 0, reason: 'name_trump' }))?.targetAction).toBe('trump');
  });

  it('points at pass', () => {
    expect(getRookHint(state({ pass: true, reason: 'weak_hand' }))?.targetAction).toBe('pass');
  });

  it('points at the bid controls', () => {
    expect(getRookHint(state({ bid: 100, reason: 'strong_hand' }))?.targetAction).toBe('bid');
  });

  // **ビッド 0 も正当。**
  it('keeps a bid of zero', () => {
    expect(getRookHint(state({ bid: 0, reason: 'strong_hand' }))?.targetAction).toBe('bid');
  });

  it('returns null when the hint names no decision', () => {
    expect(getRookHint(state({ reason: 'strong_hand' }))).toBeNull();
  });
});
