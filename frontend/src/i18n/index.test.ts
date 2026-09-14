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

const domainErrorCodes = {
  contractrummy: [
    'contractrummy.errDiscardPileEmpty',
    'contractrummy.errContractAlreadyMet',
    'contractrummy.errContractMeldCount',
    'contractrummy.errContractSlotCardCount',
    'contractrummy.errCardIndexOutOfRange',
    'contractrummy.errDuplicateCardIndex',
    'contractrummy.errContractSlotInvalid',
    'contractrummy.errExtraMeldContractRequired',
    'contractrummy.errMeldMinimumCards',
    'contractrummy.errInvalidMeld',
    'contractrummy.errLayoffContractRequired',
    'contractrummy.errTargetPlayerInvalid',
    'contractrummy.errTargetContractNotMet',
    'contractrummy.errTargetMeldInvalid',
    'contractrummy.errLayoffCardCannotAdd',
    'contractrummy.errContractRequiredToGoOut',
  ],
  carioca: [
    'carioca.errDiscardPileEmpty',
    'carioca.errContractAlreadyMet',
    'carioca.errContractMeldCount',
    'carioca.errContractSlotCardCount',
    'carioca.errCardIndexOutOfRange',
    'carioca.errDuplicateCardIndex',
    'carioca.errContractSlotInvalid',
    'carioca.errExtraMeldContractRequired',
    'carioca.errMeldMinimumCards',
    'carioca.errInvalidMeld',
    'carioca.errLayoffContractRequired',
    'carioca.errTargetPlayerInvalid',
    'carioca.errTargetContractNotMet',
    'carioca.errTargetMeldInvalid',
    'carioca.errLayoffCardCannotAdd',
    'carioca.errContractRequiredToGoOut',
  ],
  kalooki: [
    'kalooki.errDiscardPileEmpty',
    'kalooki.errMeldRequired',
    'kalooki.errMeldMinimumCards',
    'kalooki.errCardIndexOutOfRange',
    'kalooki.errDuplicateCardIndex',
    'kalooki.errInvalidMeld',
    'kalooki.errOpeningMinimumNotMet',
    'kalooki.errLayoffOpenRequired',
    'kalooki.errTargetPlayerInvalid',
    'kalooki.errTargetNotOpen',
    'kalooki.errTargetMeldInvalid',
    'kalooki.errLayoffCardCannotAdd',
  ],
  sevenbridge: [
    'sevenbridge.errPonCardIndicesRequired',
    'sevenbridge.errDiscardPileEmpty',
    'sevenbridge.errPonRankMismatch',
    'sevenbridge.errChiCardIndicesRequired',
    'sevenbridge.errChiSuitMismatch',
    'sevenbridge.errChiNotConsecutive',
    'sevenbridge.errMeldMinimumCards',
    'sevenbridge.errInvalidMeld',
    'sevenbridge.errTargetPlayerInvalid',
    'sevenbridge.errTargetMeldInvalid',
    'sevenbridge.errCardIndexOutOfRange',
    'sevenbridge.errLayoffCardCannotAdd',
    'sevenbridge.errMeldRequiredToGoOut',
    'sevenbridge.errDiscardRestriction',
  ],
  bolivia: [
    'bolivia.errDiscardPileEmpty',
    'bolivia.errBlackThreeCannotTakeDiscardPile',
    'bolivia.errWildCardCannotTakeDiscardPile',
    'bolivia.errNaturalPairIndicesRequired',
    'bolivia.errCardIndexOutOfRange',
    'bolivia.errSameCard',
    'bolivia.errPairMustBeNatural',
    'bolivia.errPairRankMismatch',
    'bolivia.errInitialMeldMinimumNotMet',
    'bolivia.errTopCardMustBeMelded',
    'bolivia.errDuplicateCardIndex',
    'bolivia.errRedThreeCannotBeDiscarded',
    'bolivia.errHandMustHaveAtMostOneCardToGoOut',
    'bolivia.errMeldNeedsAtLeastThreeCards',
    'bolivia.errBlackThreeCannotMeld',
    'bolivia.errSetMeldCardsMustHaveSameRank',
    'bolivia.errMeldNeedsAtLeastTwoNaturalCards',
    'bolivia.errMeldAllowsAtMostThreeWildCards',
    'bolivia.errWildCardsCannotExceedNaturalCards',
    'bolivia.errMeldCardRankMismatch',
    'bolivia.errCompletedBoliviaCannotBeExtended',
    'bolivia.errSequenceNeedsAtLeastThreeCards',
    'bolivia.errSequenceCannotUseWildCards',
    'bolivia.errThreeCannotBeUsedInSequence',
    'bolivia.errSequenceMeldMustUseSameSuit',
    'bolivia.errSequenceMeldCannotDuplicateCard',
    'bolivia.errSequenceMeldRanksMustBeConsecutive',
    'bolivia.errCompletedMeldsRequiredToGoOut',
    'bolivia.errEscaleraRequiredToGoOut',
  ],
  samba: [
    'samba.errDiscardPileEmpty',
    'samba.errBlackThreeCannotTakeDiscardPile',
    'samba.errWildCardCannotTakeDiscardPile',
    'samba.errNaturalPairIndicesRequired',
    'samba.errCardIndexOutOfRange',
    'samba.errSameCard',
    'samba.errPairMustBeNatural',
    'samba.errPairRankMismatch',
    'samba.errInitialMeldMinimumNotMet',
    'samba.errTopCardMustBeMelded',
    'samba.errDuplicateCardIndex',
    'samba.errRedThreeCannotBeDiscarded',
    'samba.errHandMustHaveAtMostOneCardToGoOut',
    'samba.errMeldNeedsAtLeastThreeCards',
    'samba.errBlackThreeCannotMeld',
    'samba.errSetMeldCardsMustHaveSameRank',
    'samba.errMeldNeedsAtLeastTwoNaturalCards',
    'samba.errMeldAllowsAtMostThreeWildCards',
    'samba.errWildCardsCannotExceedNaturalCards',
    'samba.errMeldCardRankMismatch',
    'samba.errSequenceNeedsAtLeastThreeCards',
    'samba.errSequenceCannotUseWildCards',
    'samba.errThreeCannotBeUsedInSequence',
    'samba.errSequenceMeldMustUseSameSuit',
    'samba.errSequenceMeldCannotDuplicateCard',
    'samba.errSequenceMeldRanksMustBeConsecutive',
    'samba.errCompletedMeldsRequiredToGoOut',
  ],
  handandfoot: [
    'handandfoot.errDiscardPileEmpty',
    'handandfoot.errBlackThreeCannotTakeDiscardPile',
    'handandfoot.errWildCardCannotTakeDiscardPile',
    'handandfoot.errNaturalPairIndicesRequired',
    'handandfoot.errCardIndexOutOfRange',
    'handandfoot.errSameCard',
    'handandfoot.errPairMustBeNatural',
    'handandfoot.errPairRankMismatch',
    'handandfoot.errTopCardMustBeMelded',
    'handandfoot.errDuplicateCardIndex',
    'handandfoot.errRedThreeCannotBeDiscarded',
    'handandfoot.errHandMustHaveAtMostOneCardToGoOut',
    'handandfoot.errMeldNeedsAtLeastThreeCards',
    'handandfoot.errMeldCardsMustHaveSameRank',
    'handandfoot.errBlackThreeCannotMeld',
    'handandfoot.errMeldNeedsAtLeastTwoNaturalCards',
    'handandfoot.errMeldAllowsAtMostThreeWildCards',
    'handandfoot.errWildCardsCannotExceedNaturalCards',
    'handandfoot.errMeldCardRankMismatch',
    'handandfoot.errGoOutRequirementsNotMet',
  ],
} as const;

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

  it.each([
    ['contractrummy', '捨て札が空です', 'The discard pile is empty.'],
    ['carioca', '捨て札が空です', 'The discard pile is empty.'],
    ['kalooki', '捨て札が空です', 'The discard pile is empty.'],
    ['sevenbridge', 'ポンには手札2枚のインデックスが必要です', 'Pon requires two hand card indices.'],
    ['bolivia', '捨て札の山が空です', 'The discard pile is empty.'],
    ['samba', '捨て札の山が空です', 'The discard pile is empty.'],
    ['handandfoot', '捨て札の山が空です', 'The discard pile is empty.'],
  ] as const)('resolves %s error message codes through the common namespace', async (game, jaExpected, enExpected) => {
    await i18n.changeLanguage('ja');
    const ja = i18n.t(`messageCode.${domainErrorCodes[game][0]}`);
    expect(ja).toBe(jaExpected);
    for (const code of domainErrorCodes[game]) {
      const translation = i18n.t(`messageCode.${code}`);
      expect(translation).not.toBe('');
      expect(translation).not.toMatch(/^messageCode\./);
    }

    await i18n.changeLanguage('en');
    const en = i18n.t(`messageCode.${domainErrorCodes[game][0]}`);
    expect(en).toBe(enExpected);
    expect(en).not.toBe(ja);
    for (const code of domainErrorCodes[game]) {
      const translation = i18n.t(`messageCode.${code}`);
      expect(translation).not.toBe('');
      expect(translation).not.toMatch(/^messageCode\./);
    }
  });
});
