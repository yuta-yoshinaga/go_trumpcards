import { describe, expect, it } from 'vitest';
import { ruleHintText } from './maoRuleHint';

// Stands in for i18next: known keys translate, unknown ones come back as-is.
const t = (key: string): string => (key === 'ruleHint.hintSuit' ? 'あるスートを出したときに言葉が必要です。' : key);

describe('ruleHintText', () => {
  it('translates the code rather than showing the server string', () => {
    expect(ruleHintText({ ruleHint: 'サーバの言語で届いた文', ruleHintCode: 'hintSuit' }, t)).toBe(
      'あるスートを出したときに言葉が必要です。',
    );
  });

  it('falls back to the server string when no code was sent', () => {
    expect(ruleHintText({ ruleHint: 'サーバの言語で届いた文' }, t)).toBe('サーバの言語で届いた文');
  });

  // **キーをそのまま出さない。**未知のコードは訳せずキーが返ってくるので、
  // それを画面に出すと `ruleHint.hintWhatever` と表示される。
  it('falls back to the server string when the code has no translation', () => {
    expect(ruleHintText({ ruleHint: 'サーバの言語で届いた文', ruleHintCode: 'hintUnknown' }, t)).toBe(
      'サーバの言語で届いた文',
    );
  });

  it('is empty when the hint is locked', () => {
    expect(ruleHintText({ ruleHint: '' }, t)).toBe('');
  });
});
