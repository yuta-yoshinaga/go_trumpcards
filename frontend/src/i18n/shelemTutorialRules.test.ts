import { describe, expect, it } from 'vitest';
import en from './locales/en/shelem.json';
import ja from './locales/ja/shelem.json';

describe('Shelem tutorial point rules', () => {
  it('explains that 55 is the minimum bid and 100 is the round total in Japanese', () => {
    expect(ja.tutorial.points).toContain('入札の下限は**55点**、上限は**100点**');
    expect(ja.tutorial.points).toContain('1ラウンドのカード点の合計');
    expect(ja.tutorial.points).not.toContain('下限が100点');
  });

  it('explains that 55 is the minimum bid and 100 is the round total in English', () => {
    expect(en.tutorial.points).toContain('minimum bid is **55** and the maximum is **100**');
    expect(en.tutorial.points).toContain('total card points in a round');
    expect(en.tutorial.points).not.toContain('bidding starts there');
  });
});
