import { afterEach, describe, expect, it } from 'vitest';
import i18n from './index';

const canastaErrorCodes = [
  'canasta.errDiscardPileEmpty',
  'canasta.errBlackThreeCannotTakeDiscardPile',
  'canasta.errWildCardCannotTakeDiscardPile',
  'canasta.errNaturalPairIndicesRequired',
  'canasta.errCardIndexOutOfRange',
  'canasta.errSameCard',
  'canasta.errPairMustBeNatural',
  'canasta.errPairRankMismatch',
  'canasta.errInitialMeldMinimumNotMet',
  'canasta.errTopCardMustBeMelded',
  'canasta.errDuplicateCardIndex',
  'canasta.errRedThreeCannotBeDiscarded',
  'canasta.errPozzettoRequiredToGoOut',
  'canasta.errCanastaRequiredToGoOut',
  'canasta.errHandMustHaveAtMostOneCardToGoOut',
  'canasta.errMeldNeedsAtLeastThreeCards',
  'canasta.errMeldCardsMustHaveSameRank',
  'canasta.errBlackThreeCannotMeld',
  'canasta.errMeldNeedsAtLeastTwoNaturalCards',
  'canasta.errMeldAllowsAtMostThreeWildCards',
  'canasta.errWildCardsCannotExceedNaturalCards',
  'canasta.errMeldCardRankMismatch',
  'canasta.errSequenceMustUseSameSuit',
  'canasta.errSequenceCannotDuplicateRank',
  'canasta.errSequenceRanksNotConsecutive',
  'canasta.errWildCardsExceedSequenceRange',
] as const;

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

  it('resolves Canasta error message codes through the common namespace', async () => {
    await i18n.changeLanguage('ja');
    const jaDiscardPileEmpty = i18n.t(`messageCode.${canastaErrorCodes[0]}`);
    expect(jaDiscardPileEmpty).toBe('捨て札の山が空です');

    for (const code of canastaErrorCodes) {
      const translation = i18n.t(`messageCode.${code}`);
      expect(translation).not.toBe('');
      expect(translation).not.toMatch(/^messageCode\./);
    }

    await i18n.changeLanguage('en');
    const enDiscardPileEmpty = i18n.t(`messageCode.${canastaErrorCodes[0]}`);
    expect(enDiscardPileEmpty).toBe('The discard pile is empty.');
    expect(enDiscardPileEmpty).not.toBe(jaDiscardPileEmpty);

    for (const code of canastaErrorCodes) {
      const translation = i18n.t(`messageCode.${code}`);
      expect(translation).not.toBe('');
      expect(translation).not.toMatch(/^messageCode\./);
    }
  });
});
