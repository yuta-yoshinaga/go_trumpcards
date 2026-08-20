import { describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { niuniuBankerResultText, niuniuRankText } from './niuniuRankText';

describe('niuniuRankText', () => {
  it('names each rank the server can send', () => {
    expect(niuniuRankText('none')).toBe('無牛');
    expect(niuniuRankText('niuniu')).toBe('牛牛');
    expect(niuniuRankText('n3')).toBe('牛3');
    expect(niuniuRankText('n9')).toBe('牛9');
  });

  it('renders nothing for a hidden hand or an unknown key', () => {
    // A hidden hand arrives with an empty key; rendering the key itself would
    // put "n0" or "" on screen where a rank belongs.
    for (const key of ['', 'n0', 'n10', 'niu', 'NIUNIU', '3']) {
      expect(niuniuRankText(key)).toBe('');
    }
  });

  it('follows a language switch instead of returning the first locale', async () => {
    // This is the whole point of the key: the server no longer picks the words.
    await i18n.changeLanguage('en');
    expect(niuniuRankText('niuniu')).toBe('Niu Niu');
    expect(niuniuRankText('none')).toBe('No Niu');
    expect(niuniuRankText('n3')).toBe('Niu 3');
    await i18n.changeLanguage('ja');
    expect(niuniuRankText('niuniu')).toBe('牛牛');
  });
});

describe('niuniuBankerResultText', () => {
  it('builds the headline around the rank', () => {
    expect(niuniuBankerResultText('niuniu')).toBe('親: 牛牛');
    expect(niuniuBankerResultText('none')).toBe('親: 無牛');
    expect(niuniuBankerResultText('n7')).toBe('親: 牛7');
  });

  it('is empty before the round settles, rather than a bare heading', () => {
    expect(niuniuBankerResultText('')).toBe('');
    expect(niuniuBankerResultText('bogus')).toBe('');
  });

  it('localises the heading and the rank together', async () => {
    await i18n.changeLanguage('en');
    expect(niuniuBankerResultText('niuniu')).toBe('Banker: Niu Niu');
    await i18n.changeLanguage('ja');
  });
});
