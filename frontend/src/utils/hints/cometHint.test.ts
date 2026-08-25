import { describe, expect, it } from 'vitest';
import { makeCometState } from '../../test/stateFactories';
import { getCometHint } from './cometHint';

describe('getCometHint', () => {
  it('returns null when there is nothing to suggest', () => {
    expect(getCometHint(makeCometState())).toBeNull();
    expect(getCometHint(makeCometState({ hintHandIdx: 0, hintReason: '' }))).toBeNull();
    expect(getCometHint(makeCometState({ hintHandIdx: 0, hintReason: 'none' }))).toBeNull();
    expect(getCometHint(makeCometState({ hintHandIdx: -1, hintReason: 'follow' }))).toBeNull();
  });

  // **理由は `hint.<snake_case>` の形で返す。** `ns:key` にすると i18next の
  // 既定の nsSeparator に当たって、生の識別子がそのまま画面に出る。
  it('maps each reason onto a hint.* key', () => {
    for (const reason of ['go_out', 'follow', 'comet', 'king', 'pass']) {
      const hint = getCometHint(makeCometState({ hintHandIdx: 1, hintReason: reason }));
      expect(hint).not.toBeNull();
      expect(hint?.reason).toBe(`hint.${reason}`);
      expect(hint?.reason).not.toContain(':');
      expect(hint?.targetAction).toBe('play');
      expect(hint?.confidence).toBe('moderate');
    }
  });
});
