import { describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import type { BridgeResponse } from '../../types/card';
import { getBridgeHint } from './bridgeHint';

const state = (hint?: BridgeResponse['hint']): BridgeResponse => ({ hint }) as BridgeResponse;

describe('getBridgeHint', () => {
  it('returns null without a hint', () => {
    expect(getBridgeHint(state())).toBeNull();
  });

  it('points at the card to play', () => {
    expect(getBridgeHint(state({ cardIndex: 5, reason: 'follow_suit' }))).toEqual({
      targetAction: 'card-5',
      reason: 'hintReason.follow_suit',
      confidence: 'moderate',
    });
  });

  // **札 0 は正当な手。**cardIndex は省略可能なので真偽値で見ると、手札の
  // 先頭を出せというヒントだけが黙って消える。
  it('keeps a hint that points at card index 0', () => {
    expect(getBridgeHint(state({ cardIndex: 0, reason: 'lead_strong' }))?.targetAction).toBe('card-0');
  });

  it('raises the confidence when the suggestion is forced by the lead', () => {
    expect(getBridgeHint(state({ cardIndex: 2, reason: 'follow_suit' }))?.confidence).toBe('moderate');
    expect(getBridgeHint(state({ cardIndex: 2, reason: 'trump_cut' }))?.confidence).toBe('strong');
  });

  it('points at the bid controls during the auction', () => {
    expect(getBridgeHint(state({ bidType: 0, bidLevel: 3, bidSuit: 2, reason: 'strategic_bid' }))).toEqual({
      targetAction: 'bid',
      reason: 'hintReason.strategic_bid',
      confidence: 'moderate',
    });
  });

  // **パスも 0 の一種。**bidLevel 0 / bidSuit 0 のパス提案を真偽値で
  // 落とさないこと。
  it('keeps a pass suggestion', () => {
    expect(getBridgeHint(state({ bidType: 1, bidLevel: 0, bidSuit: 0, reason: 'strategic_bid' }))?.targetAction).toBe(
      'bid',
    );
  });

  it('returns null when the hint names neither a card nor a bid', () => {
    expect(getBridgeHint(state({ reason: 'strategic_bid' }))).toBeNull();
  });
});

// **バックエンドが出す理由キーが、訳を持っているか。**`hintReason` には
// `strategic_bid` / `lead_strong` / `discard_weak` が無く、ビッド局面の
// ヒントは**キー文字列そのもの**を表示していた。ボタン式のヒントも同じ
// キーを読むので、ここが落ちれば両方が落ちる。
describe('bridge hintReason keys', () => {
  // Bridge.GetHint と playHintReason が返しうる全て。
  const REASONS = ['strategic_bid', 'lead_trump', 'lead_strong', 'follow_suit', 'trump_cut', 'discard_weak'];

  it.each(REASONS)('translates %s', (key) => {
    expect(i18n.t(`bridge:hintReason.${key}`)).not.toBe(`hintReason.${key}`);
  });
});
