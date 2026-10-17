import { describe, expect, it } from 'vitest';
import { isAllowedLabelWrap } from './labelWrapAllowlist';

describe('isAllowedLabelWrap', () => {
  it('allows observed ordinary sentence wraps at mobile size', () => {
    expect(isAllowedLabelWrap('/bouillotte', '参加中', { width: 375, height: 667 })).toBe(true);
    expect(isAllowedLabelWrap('/threecardbrag', 'ブラインド', { width: 375, height: 667 })).toBe(true);
  });

  it('allows only listed paytable wraps on their routes at mobile size', () => {
    expect(isAllowedLabelWrap('/videopoker', 'ロイヤルフラッシュ', { width: 375, height: 667 })).toBe(true);
    expect(isAllowedLabelWrap('/deuceswild', 'フォーデューシーズ', { width: 375, height: 667 })).toBe(true);
    expect(isAllowedLabelWrap('/jokerpoker', 'ストレートフラッシュ', { width: 375, height: 667 })).toBe(true);
    expect(isAllowedLabelWrap('/jokerpoker', 'フォーデューシーズ', { width: 375, height: 667 })).toBe(false);
  });

  it('does not allow any wrap at other viewport sizes', () => {
    expect(isAllowedLabelWrap('/videopoker', 'ロイヤルフラッシュ', { width: 1280, height: 800 })).toBe(false);
    expect(isAllowedLabelWrap('/bouillotte', '参加中', { width: 375, height: 800 })).toBe(false);
  });
});
