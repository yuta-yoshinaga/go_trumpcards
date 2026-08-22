import { describe, expect, it } from 'vitest';
import i18n from '../../i18n';
import type { RamschResponse } from '../../types/card';
import { RamschPhase } from '../../types/phases';
import { getRamschHint } from './ramschHint';

const state = (overrides: Partial<RamschResponse> = {}): RamschResponse =>
  ({
    players: [],
    phase: RamschPhase.PLAY,
    gameEndFlag: false,
    hint: { cardIndex: 2, reason: 'avoid_points' },
    ...overrides,
  }) as RamschResponse;

describe('getRamschHint', () => {
  it('points at the card the backend chose', () => {
    const hint = getRamschHint(state());
    expect(hint?.targetPos).toBe(2);
    expect(hint?.targetAction).toBe('play');
  });

  // **理由キーが実際に翻訳へ解決すること。**
  //
  // `ramsch:avoid_points` と書いていたときは、i18next の nsSeparator (既定 ':')
  // が効いて「名前空間 ramsch のキー avoid_points」を探しに行き、そんなキーは
  // 無いので**生の識別子がそのままツールチップに出ていた**。
  // 「文字列が返る」ことだけを見るテストでは通ってしまう。
  it.each(['avoid_points', 'lead_low', 'forced_discard'])('resolves the reason %s to real text', (reason) => {
    const hint = getRamschHint(state({ hint: { cardIndex: 0, reason } }));
    const key = hint?.reason ?? '';
    const text = i18n.t(key, { ns: 'ramsch' });

    expect(text).not.toBe(key);
    expect(text).not.toContain(reason);
    expect(text.length).toBeGreaterThan(4);
  });

  // **負のコントロール**: 存在しない理由なら解決せず、キーがそのまま返る。
  // 上の assert が「何でも通る」わけではないことを示す。
  it('leaves an unknown reason unresolved', () => {
    const hint = getRamschHint(state({ hint: { cardIndex: 0, reason: 'no_such_reason' } }));
    const key = hint?.reason ?? '';
    expect(i18n.t(key, { ns: 'ramsch' })).toBe(key);
  });

  // **理由が空でも壊れない。** サーバは必ず理由を入れるが、KV の版ずれや
  // 壊れた応答で空が届きうる。そのときも「解決しないキー」ではなく既定の
  // 理由に落として、生の `hint.` を出さない。
  it('falls back to a real reason when the response carries none', () => {
    const hint = getRamschHint(state({ hint: { cardIndex: 1, reason: '' } }));
    const key = hint?.reason ?? '';
    expect(key).toBe('hint.avoid_points');
    expect(i18n.t(key, { ns: 'ramsch' })).not.toBe(key);
  });

  it('stays quiet once the hand is over or with no hint', () => {
    expect(getRamschHint(state({ gameEndFlag: true }))).toBeNull();
    expect(getRamschHint(state({ hint: undefined }))).toBeNull();
    expect(getRamschHint(state({ phase: RamschPhase.ROUND_END }))).toBeNull();
    expect(getRamschHint(state({ hint: { reason: 'avoid_points' } }))).toBeNull();
  });
});
