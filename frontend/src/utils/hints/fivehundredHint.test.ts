import { describe, expect, it } from 'vitest';
import type { FiveHundredResponse } from '../../types/card';
import { getFiveHundredHint } from './fivehundredHint';

const state = (hint?: FiveHundredResponse['hint']): FiveHundredResponse => ({ hint }) as FiveHundredResponse;

describe('getFiveHundredHint', () => {
  it('returns null without a hint', () => {
    expect(getFiveHundredHint(state())).toBeNull();
  });

  it('points at the card to play', () => {
    expect(getFiveHundredHint(state({ cardIndex: 2, reason: 'lead_high' }))).toEqual({
      targetAction: 'card-2',
      reason: 'hint.lead_high',
      confidence: 'moderate',
    });
  });

  // **札 0 は正当な手。**真偽値で見ると先頭だけ落ちる。
  it('keeps a hint that points at card index 0', () => {
    expect(getFiveHundredHint(state({ cardIndex: 0, reason: 'lead_high' }))?.targetAction).toBe('card-0');
  });

  it('points at the kitty discard', () => {
    expect(getFiveHundredHint(state({ discardIndices: [0], reason: 'discard_low' }))?.targetAction).toBe('discard');
  });

  it('keeps an empty discard list as a discard hint', () => {
    expect(getFiveHundredHint(state({ discardIndices: [], reason: 'discard_low' }))?.targetAction).toBe('discard');
  });

  it('points at pass', () => {
    expect(getFiveHundredHint(state({ pass: true, reason: 'weak_hand' }))?.targetAction).toBe('pass');
  });

  // `pass: false` は「パスするな」であって「パスしろ」ではない。
  it('does not read a false pass as a pass suggestion', () => {
    expect(getFiveHundredHint(state({ pass: false, bidTricks: 7, reason: 'strong_hand' }))?.targetAction).toBe('bid');
  });

  it('points at the bid controls', () => {
    expect(getFiveHundredHint(state({ bidKind: 1, bidTricks: 6, reason: 'strong_hand' }))?.targetAction).toBe('bid');
  });

  // **ビッド種別 0 も正当。**
  it('keeps a bid of kind zero', () => {
    expect(getFiveHundredHint(state({ bidKind: 0, reason: 'strong_hand' }))?.targetAction).toBe('bid');
  });

  it('returns null when the hint names no decision', () => {
    expect(getFiveHundredHint(state({ reason: 'strong_hand' }))).toBeNull();
  });
});
