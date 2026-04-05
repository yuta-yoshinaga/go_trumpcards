import { afterEach, describe, expect, it } from 'vitest';
import i18n from './index';

describe('i18n lang sync', () => {
  const originalLang = document.documentElement.lang;

  afterEach(async () => {
    await i18n.changeLanguage(originalLang);
    document.documentElement.lang = originalLang;
  });

  it('sets document.documentElement.lang on module load', () => {
    expect(document.documentElement.lang).toBeTruthy();
  });

  it('updates document.documentElement.lang when language changes', async () => {
    await i18n.changeLanguage('en');
    expect(document.documentElement.lang).toBe('en');

    await i18n.changeLanguage('ja');
    expect(document.documentElement.lang).toBe('ja');
  });
});
