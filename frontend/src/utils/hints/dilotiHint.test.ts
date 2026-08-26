import { describe, expect, it } from 'vitest';
import { makeDilotiState } from '../../test/stateFactories';
import { getDilotiHint } from './dilotiHint';

describe('getDilotiHint', () => {
  it('returns null when there is nothing to suggest', () => {
    expect(getDilotiHint(makeDilotiState())).toBeNull();
    expect(getDilotiHint(makeDilotiState({ hintHandIdx: 0, hintReason: '' }))).toBeNull();
    expect(getDilotiHint(makeDilotiState({ hintHandIdx: 0, hintReason: 'none' }))).toBeNull();
    expect(getDilotiHint(makeDilotiState({ hintHandIdx: -1, hintReason: 'capture' }))).toBeNull();
  });

  // **理由は `hint.<snake_case>` の形で返す。** `ns:key` にすると i18next の
  // 既定の nsSeparator に当たって、生の識別子がそのまま画面に出る。
  it('maps each reason onto a hint.* key', () => {
    for (const reason of ['capture', 'declare', 'trail']) {
      const hint = getDilotiHint(makeDilotiState({ hintHandIdx: 1, hintReason: reason }));
      expect(hint).not.toBeNull();
      expect(hint?.reason).toBe(`hint.${reason}`);
      expect(hint?.reason).not.toContain(':');
      expect(hint?.targetAction).toBe('play');
      expect(hint?.confidence).toBe('moderate');
    }
  });
});
