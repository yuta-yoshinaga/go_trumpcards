import { describe, expect, it } from 'vitest';
import type { NertzResponse } from '../../types/card';
import { getNertzHint } from './nertzHint';

const state = (hint: NertzResponse['hint']): NertzResponse => ({ hint }) as NertzResponse;

// **CUI の方が具体的だった。**NertzCuiPresenter は同じフィールドから
// 「ナッツ → ファウンデーション2」を組み立てているのに、Web の reason は
// 「移動先を選んでください」の固定文だった (#4885)。
describe('getNertzHint', () => {
  it('returns null without a server hint', () => {
    expect(getNertzHint(state(undefined))).toBeNull();
  });

  it('names the source and the destination', () => {
    const r = getNertzHint(state({ fromZone: 'nertz', fromCol: -1, cardIndex: -1, toZone: 'foundation', toCol: 2 }));
    expect(r?.reason).toBe('messages.hintMove');
    expect(r?.reasonParams).toEqual({ from: 'ナッツ', to: 'ファウンデーション2' });
    expect(r?.confidence).toBe('moderate');
  });

  it('includes the tableau column and card position of the source', () => {
    const r = getNertzHint(state({ fromZone: 'tableau', fromCol: 3, cardIndex: 1, toZone: 'tableau', toCol: 0 }));
    expect(r?.reasonParams).toEqual({ from: 'タブロー3(idx=1)', to: 'タブロー0' });
  });

  // **0 は実在する。**タブローの一番下から丸ごと動かす手がこれ。「0枚目」のような
  // 序数表現にすると日本語として壊れるので、CUI と同じ生の索引で出す。
  it('renders index 0 without pretending it is an ordinal', () => {
    const r = getNertzHint(state({ fromZone: 'tableau', fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: 1 }));
    expect(r?.reasonParams).toEqual({ from: 'タブロー2(idx=0)', to: 'ファウンデーション1' });
  });

  it('keeps the targetAction identifier the page keys off', () => {
    const r = getNertzHint(state({ fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 1 }));
    expect(r?.targetAction).toBe('waste-to-tableau-1');
    expect(r?.reasonParams).toEqual({ from: 'ウェイスト', to: 'タブロー1' });
  });
});
