import { describe, expect, it } from 'vitest';
import en from './locales/en/scopa.json';
import ja from './locales/ja/scopa.json';

describe('Scopa capture action translations', () => {
  it('uses natural Japanese word order and keeps the scopa marker', () => {
    expect(ja.actionCapture).toBe('{{played}}を出して{{count}}枚を捕獲{{suffix}}');
    expect(ja.actionScopaSuffix).toBe('（スコパ！）');
    expect(
      ja.actionCapture
        .replace('{{played}}', '7♦')
        .replace('{{count}}', '3')
        .replace('{{suffix}}', ja.actionScopaSuffix),
    ).toBe('7♦を出して3枚を捕獲（スコパ！）');
  });

  it('keeps the English capture action paired with the same keys', () => {
    expect(en.actionCapture).toContain('{{played}}');
    expect(en.actionCapture).toContain('{{count}}');
    expect(en.actionScopaSuffix).toContain('Scopa');
  });
});
