import { afterEach, describe, expect, it } from 'vitest';
import i18n from './index';
import enCommon from './locales/en/common.json';
import jaCommon from './locales/ja/common.json';

const domainErrorCodes = Object.keys(enCommon.messageCode).reduce<Record<string, string[]>>((codesByGame, code) => {
  const [game, errorCode] = code.split('.', 2);
  if (game && errorCode && /^err[A-Z]/.test(errorCode)) {
    const gameCodes = codesByGame[game] ?? [];
    gameCodes.push(code);
    codesByGame[game] = gameCodes;
  }
  return codesByGame;
}, {});

const allDomainErrorCodes = Object.values(domainErrorCodes).flat();

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

  it('derives at least 200 domain error message codes', () => {
    expect(allDomainErrorCodes.length).toBeGreaterThanOrEqual(200);
  });

  it('resolves every derived domain error message code in both locales', async () => {
    expect(allDomainErrorCodes.length).toBeGreaterThanOrEqual(200);

    await i18n.changeLanguage('ja');
    for (const code of allDomainErrorCodes) {
      expect(Object.hasOwn(jaCommon.messageCode, code)).toBe(true);

      const jaTranslation = i18n.t(`messageCode.${code}`);
      expect(jaTranslation).not.toBe('');
      expect(jaTranslation).not.toMatch(/^messageCode\./);
    }

    await i18n.changeLanguage('en');
    for (const code of allDomainErrorCodes) {
      const enTranslation = i18n.t(`messageCode.${code}`);
      expect(enTranslation).not.toBe('');
      expect(enTranslation).not.toMatch(/^messageCode\./);
    }
  });

  it('resolves Canasta error message codes through the common namespace', async () => {
    const canastaErrorCodes = domainErrorCodes.canasta;
    // Typed string[] but undefined at runtime if the canasta.err* keys are ever
    // renamed. Assert it here so that regression reads as a failed expectation
    // rather than a bare TypeError from the loop below.
    expect(canastaErrorCodes?.length ?? 0).toBeGreaterThan(0);
    await i18n.changeLanguage('ja');
    const jaDiscardPileEmpty = i18n.t('messageCode.canasta.errDiscardPileEmpty');
    expect(jaDiscardPileEmpty).toBe('捨て札の山が空です');

    for (const code of canastaErrorCodes) {
      const translation = i18n.t(`messageCode.${code}`);
      expect(translation).not.toBe('');
      expect(translation).not.toMatch(/^messageCode\./);
    }

    await i18n.changeLanguage('en');
    const enDiscardPileEmpty = i18n.t('messageCode.canasta.errDiscardPileEmpty');
    expect(enDiscardPileEmpty).toBe('The discard pile is empty.');
    expect(enDiscardPileEmpty).not.toBe(jaDiscardPileEmpty);

    for (const code of canastaErrorCodes) {
      const translation = i18n.t(`messageCode.${code}`);
      expect(translation).not.toBe('');
      expect(translation).not.toMatch(/^messageCode\./);
    }
  });

  it.each([
    ['contractrummy', 'contractrummy.errDiscardPileEmpty', '捨て札が空です', 'The discard pile is empty.'],
    ['carioca', 'carioca.errDiscardPileEmpty', '捨て札が空です', 'The discard pile is empty.'],
    ['kalooki', 'kalooki.errDiscardPileEmpty', '捨て札が空です', 'The discard pile is empty.'],
    [
      'sevenbridge',
      'sevenbridge.errPonCardIndicesRequired',
      'ポンには手札2枚のインデックスが必要です',
      'Pon requires two hand card indices.',
    ],
    ['bolivia', 'bolivia.errDiscardPileEmpty', '捨て札の山が空です', 'The discard pile is empty.'],
    ['samba', 'samba.errDiscardPileEmpty', '捨て札の山が空です', 'The discard pile is empty.'],
    ['handandfoot', 'handandfoot.errDiscardPileEmpty', '捨て札の山が空です', 'The discard pile is empty.'],
    [
      'bauernschnapsen',
      'bauernschnapsen.errNotContractPhase',
      '契約フェーズではありません',
      'This is not the contract phase.',
    ],
    ['gaigel', 'gaigel.errCardIndexOutOfRange', 'カードインデックスが範囲外です', 'The card index is out of range.'],
    ['watten', 'watten.errInvalidSchlagRank', '無効な Schlag ランクです', 'Invalid Schlag rank.'],
    [
      'binokel',
      'binokel.errBidMinimum',
      'ビッドは{{min}}以上でなければなりません。',
      'The bid must be at least {{min}}.',
    ],
    [
      'pinochle',
      'pinochle.errBidMinimum',
      'ビッドは{{min}}以上でなければなりません。',
      'The bid must be at least {{min}}.',
    ],
    [
      'mighty',
      'mighty.errBidRange',
      '{{min}}〜{{max}}のビッドを指定してください（0でパス）。',
      'Specify a bid from {{min}} to {{max}} (0 passes).',
    ],
  ] as const)(
    'resolves %s error message codes through the common namespace',
    async (_game, code, jaExpected, enExpected) => {
      await i18n.changeLanguage('ja');
      const ja = i18n.t(`messageCode.${code}`);
      expect(ja).toBe(jaExpected);

      await i18n.changeLanguage('en');
      const en = i18n.t(`messageCode.${code}`);
      expect(en).toBe(enExpected);
      expect(en).not.toBe(ja);
    },
  );

  it.each([
    ['binokel', 'ビッドフェーズではありません。', 'This is not the bidding phase.'],
    ['pinochle', 'ビッドフェーズではありません。', 'This is not the bidding phase.'],
  ] as const)('resolves %s wrong-phase messages consistently', async (game, jaExpected, enExpected) => {
    await i18n.changeLanguage('ja');
    expect(i18n.t(`messageCode.${game}.errWrongPhase`)).toBe(jaExpected);

    await i18n.changeLanguage('en');
    expect(i18n.t(`messageCode.${game}.errWrongPhase`)).toBe(enExpected);
  });
});
