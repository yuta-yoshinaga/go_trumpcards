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
  bauernschnapsen: [
    'bauernschnapsen.errNotContractPhase',
    'bauernschnapsen.errNotYourTurn',
    'bauernschnapsen.errContractUnavailable',
    'bauernschnapsen.errTrumpSuitRequired',
    'bauernschnapsen.errCardIndexOutOfRange',
    'bauernschnapsen.errMarriageLeadOnly',
    'bauernschnapsen.errMarriageUnavailable',
    'bauernschnapsen.errInvalidPhase',
    'bauernschnapsen.errInvalidTrumpSuit',
    'bauernschnapsen.errInvalidPlayerCount',
    'bauernschnapsen.errPlayerIndexOutOfRange',
    'bauernschnapsen.errLeadWinnerIndexOutOfRange',
    'bauernschnapsen.errWinnerTeamOutOfRange',
    'bauernschnapsen.errStateSliceInvalid',
    'bauernschnapsen.errPlayerNil',
    'bauernschnapsen.errTrickCardNil',
    'bauernschnapsen.errActionLogEntryNil',
  ],
  gaigel: [
    'gaigel.errCardIndexOutOfRange',
    'gaigel.errMarriageLeadOnly',
    'gaigel.errMarriageUnavailable',
    'gaigel.errInvalidPhase',
    'gaigel.errInvalidTrumpSuit',
    'gaigel.errInvalidPlayerCount',
    'gaigel.errPlayerIndexOutOfRange',
    'gaigel.errLeadWinnerIndexOutOfRange',
    'gaigel.errWinnerTeamOutOfRange',
    'gaigel.errStateSliceInvalid',
    'gaigel.errPlayerNil',
    'gaigel.errTrickCardNil',
    'gaigel.errActionLogEntryNil',
  ],
  watten: [
    'watten.errInvalidSchlagRank',
    'watten.errInvalidCriticalSuit',
    'watten.errCardIndexOutOfRange',
    'watten.errCannotRaiseNow',
    'watten.errMustFollowTrump',
    'watten.errMustFollowLeadSuit',
    'watten.errInvalidPhase',
    'watten.errInvalidPlayerCount',
    'watten.errPlayerNil',
    'watten.errPlayerTeamOutOfRange',
    'watten.errTooManyTrickCards',
    'watten.errTrickCardNil',
    'watten.errTrickPlayerIndexOutOfRange',
    'watten.errActionLogTooLarge',
    'watten.errDeclarationRequired',
    'watten.errCurrentPlayerIndexOutOfRange',
    'watten.errDealerIndexOutOfRange',
    'watten.errPlayerIndexOutOfRange',
    'watten.errTeamIndexOutOfRange',
  ],
  binokel: [
    'binokel.errBidMinimum',
    'binokel.errBidHigherThanHighest',
    'binokel.errBidStep',
    'binokel.errDabbPhase',
    'binokel.errDabbCardCount',
    'binokel.errCardIndexOutOfRange',
    'binokel.errDuplicateCardIndex',
    'binokel.errTrumpPhase',
    'binokel.errPlayPhase',
    'binokel.errInvalidCardIndex',
    'binokel.errCardCannotBePlayed',
    'binokel.errWrongPhase',
    'binokel.errNotHumanTurn',
    'binokel.errInvalidSuit',
  ],
  pinochle: [
    'pinochle.errBidMinimum',
    'pinochle.errBidHigherThanHighest',
    'pinochle.errCannotPass',
    'pinochle.errTrumpPhase',
    'pinochle.errPlayPhase',
    'pinochle.errInvalidCardIndex',
    'pinochle.errCardCannotBePlayed',
    'pinochle.errWrongPhase',
    'pinochle.errNotHumanTurn',
    'pinochle.errInvalidSuit',
  ],
  mighty: [
    'mighty.errBidRange',
    'mighty.errBidHigherThanHighest',
    'mighty.errNoTrumpCannotHaveSuit',
    'mighty.errInvalidSuit',
    'mighty.errJokerValue',
    'mighty.errInvalidPartnerSuit',
    'mighty.errInvalidPartnerValue',
    'mighty.errKittyCardCount',
    'mighty.errCardIndexOutOfRange',
    'mighty.errDuplicateCardIndex',
    'mighty.errJokerLeadSuitRequired',
    'mighty.errJokerLeadOnlyAtTrickStart',
    'mighty.errJokerRequired',
    'mighty.errInvalidDemandSuit',
    'mighty.errJokerCallRequiresJoker',
    'mighty.errMustFollowLeadSuit',
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
    ['bauernschnapsen', '契約フェーズではありません', 'This is not the contract phase.'],
    ['gaigel', 'カードインデックスが範囲外です', 'The card index is out of range.'],
    ['watten', '無効な Schlag ランクです', 'Invalid Schlag rank.'],
    ['binokel', 'ビッドは{{min}}以上でなければなりません。', 'The bid must be at least {{min}}.'],
    ['pinochle', 'ビッドは{{min}}以上でなければなりません。', 'The bid must be at least {{min}}.'],
    [
      'mighty',
      '{{min}}〜{{max}}のビッドを指定してください（0でパス）。',
      'Specify a bid from {{min}} to {{max}} (0 passes).',
    ],
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
