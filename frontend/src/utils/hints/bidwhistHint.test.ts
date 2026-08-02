import { describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import type { BidWhistResponse } from '../../types/card';
import { getBidWhistHint } from './bidwhistHint';

const state = (hint?: BidWhistResponse['hint']): BidWhistResponse => ({ hint }) as BidWhistResponse;

describe('getBidWhistHint', () => {
  it('returns null without a hint', () => {
    expect(getBidWhistHint(state())).toBeNull();
  });

  it('points at the card to play', () => {
    expect(getBidWhistHint(state({ cardIndex: 3, reason: 'lead_high' }))).toEqual({
      targetAction: 'card-3',
      reason: 'hint.lead_high',
      confidence: 'moderate',
    });
  });

  // **札 0 は正当な手。**真偽値で見ると先頭の札だけ落ちる。
  it('keeps a hint that points at card index 0', () => {
    expect(getBidWhistHint(state({ cardIndex: 0, reason: 'lead_high' }))?.targetAction).toBe('card-0');
  });

  it('points at the discard during the kitty exchange', () => {
    expect(getBidWhistHint(state({ discardIndices: [0, 1], reason: 'discard_low' }))?.targetAction).toBe('discard');
  });

  // **空の捨て札リストも「捨てる局面」。**長さで見ると消える。
  it('keeps an empty discard list as a discard hint', () => {
    expect(getBidWhistHint(state({ discardIndices: [], reason: 'discard_low' }))?.targetAction).toBe('discard');
  });

  it('points at the trump choice', () => {
    expect(getBidWhistHint(state({ trumpSuit: 2, reason: 'name_trump' }))?.targetAction).toBe('trump');
  });

  // **スペードは 1 だが、0 (ノートランプ) も正当。**
  it('keeps a trump hint naming suit zero', () => {
    expect(getBidWhistHint(state({ trumpSuit: 0, reason: 'name_trump' }))?.targetAction).toBe('trump');
  });

  it('points at pass', () => {
    expect(getBidWhistHint(state({ pass: true, reason: 'weak_hand' }))?.targetAction).toBe('pass');
  });

  it('points at the bid controls', () => {
    expect(getBidWhistHint(state({ bidTricks: 4, bidDirection: 0, reason: 'strong_hand' }))?.targetAction).toBe('bid');
  });

  // **ビッド 0 も正当。**真偽値で見ると消える。
  it('keeps a bid of zero tricks', () => {
    expect(getBidWhistHint(state({ bidTricks: 0, reason: 'strong_hand' }))?.targetAction).toBe('bid');
  });

  // pass が false のときは「パスしろ」ではない。
  it('does not read a false pass as a pass suggestion', () => {
    expect(getBidWhistHint(state({ pass: false, bidTricks: 4, reason: 'strong_hand' }))?.targetAction).toBe('bid');
  });

  it('returns null when the hint names no decision', () => {
    expect(getBidWhistHint(state({ reason: 'strong_hand' }))).toBeNull();
  });
});

// **バックエンドが出す理由キーが、訳を持っているか。**Bridge (#4601) では
// `hintReason` に 3 キー欠けていて、ビッド局面のヒントがキー文字列を
// そのまま表示していた。同じ確認をここでもする。
describe('bidwhist hint keys', () => {
  // BidWhist.GetHint と各フェーズのヘルパーが返しうる全て。
  const REASONS = [
    'strategic_bid',
    'pass_recommended',
    'trump_longest',
    'discard_weakest',
    'lead_trump',
    'lead_strong',
    'follow_suit',
    'trump_cut',
    'discard_weak',
  ];

  it.each(REASONS)('translates %s', (key) => {
    expect(i18n.t(`bidwhist:hint.${key}`)).not.toBe(`hint.${key}`);
  });
});
