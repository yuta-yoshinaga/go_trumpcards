import { describe, expect, it } from 'vitest';
import en from './locales/en/courchevel.json';
import ja from './locales/ja/courchevel.json';

describe('Courchevel community card tutorial', () => {
  it('explains that the first card is exposed before betting in Japanese', () => {
    expect(ja.tutorial.communityCards).toContain('プリフロップ開始時');
    expect(ja.tutorial.communityCards).toContain('最初のベット前');
  });

  it('explains that the first card is exposed before betting in English', () => {
    expect(en.tutorial.communityCards).toContain('pre-flop');
    expect(en.tutorial.communityCards).toContain('before the opening bet');
  });
});
