import { describe, expect, it } from 'vitest';
import en from './locales/en/sueca.json';
import ja from './locales/ja/sueca.json';

describe('Sueca tutorial point rules', () => {
  it('explains the 61-point winning threshold in both tutorial locales', () => {
    expect(en.tutorial.info).toContain('61 or more card points');
    expect(ja.tutorial.info).toContain('61点以上');
    expect(en.pointLegend.note).toContain('61 or more');
    expect(ja.pointLegend.note).toContain('61点以上');
  });
});
