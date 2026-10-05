import { describe, expect, it } from 'vitest';
import en from './locales/en/madrasso.json';
import ja from './locales/ja/madrasso.json';

describe('Madrasso scoring text matches the game rules', () => {
  it('lists the implemented card values and deal threshold in both languages', () => {
    expect(en.tutorial.pointInfo).toContain('A = 11');
    expect(en.tutorial.pointInfo).toContain('3 = 10');
    expect(en.tutorial.pointInfo).toContain('at least 61 points');
    expect(ja.tutorial.pointInfo).toContain('A=11点');
    expect(ja.tutorial.pointInfo).toContain('3=10点');
    expect(ja.tutorial.pointInfo).toContain('61点以上');
  });

  it('explains 21 as the number of deals won and removes the old thirds scoring', () => {
    expect(en.tutorial.pointInfo).toContain('target number of deals (default 21)');
    expect(en.pointLegend.note).toContain('number of deals won');
    expect(ja.tutorial.pointInfo).toContain('獲得ディール数の目標（既定21）');
    expect(ja.pointLegend.note).toContain('獲得ディール数');
    for (const text of [en.tutorial.pointInfo, en.pointLegend.note, ja.tutorial.pointInfo, ja.pointLegend.note]) {
      expect(text).not.toMatch(/1\/3|11 points total|1ラウンド.*11点/);
    }
  });
});
