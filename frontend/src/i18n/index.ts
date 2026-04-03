import i18n from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';

import enBaccarat from './locales/en/baccarat.json';
import enBlackjack from './locales/en/blackjack.json';
import enBridge from './locales/en/bridge.json';
import enCanasta from './locales/en/canasta.json';
import enCommon from './locales/en/common.json';
import enCrazyeights from './locales/en/crazyeights.json';
import enCribbage from './locales/en/cribbage.json';
import enDaifugo from './locales/en/daifugo.json';
import enDeuceswild from './locales/en/deuceswild.json';
import enDoubt from './locales/en/doubt.json';
import enEuchre from './locales/en/euchre.json';
import enFreecell from './locales/en/freecell.json';
import enGinrummy from './locales/en/ginrummy.json';
import enGofish from './locales/en/gofish.json';
import enHearts from './locales/en/hearts.json';
import enHoldem from './locales/en/holdem.json';
import enIndianpoker from './locales/en/indianpoker.json';
import enJokerpoker from './locales/en/jokerpoker.json';
import enKlondike from './locales/en/klondike.json';
import enMemory from './locales/en/memory.json';
import enNapoleon from './locales/en/napoleon.json';
import enOhhell from './locales/en/ohhell.json';
import enOldmaid from './locales/en/oldmaid.json';
import enOmaha from './locales/en/omaha.json';
import enPineapple from './locales/en/pineapple.json';
import enPinochle from './locales/en/pinochle.json';
import enPoker from './locales/en/poker.json';
import enPyramid from './locales/en/pyramid.json';
import enSevens from './locales/en/sevens.json';
import enShortdeck from './locales/en/shortdeck.json';
import enSpades from './locales/en/spades.json';
import enSpeed from './locales/en/speed.json';
import enSpider from './locales/en/spider.json';
import enThreecard from './locales/en/threecard.json';
import enTripeaks from './locales/en/tripeaks.json';
import enTutorial from './locales/en/tutorial.json';
import enVideopoker from './locales/en/videopoker.json';
import jaBaccarat from './locales/ja/baccarat.json';
import jaBlackjack from './locales/ja/blackjack.json';
import jaBridge from './locales/ja/bridge.json';
import jaCanasta from './locales/ja/canasta.json';
import jaCommon from './locales/ja/common.json';
import jaCrazyeights from './locales/ja/crazyeights.json';
import jaCribbage from './locales/ja/cribbage.json';
import jaDaifugo from './locales/ja/daifugo.json';
import jaDeuceswild from './locales/ja/deuceswild.json';
import jaDoubt from './locales/ja/doubt.json';
import jaEuchre from './locales/ja/euchre.json';
import jaFreecell from './locales/ja/freecell.json';
import jaGinrummy from './locales/ja/ginrummy.json';
import jaGofish from './locales/ja/gofish.json';
import jaHearts from './locales/ja/hearts.json';
import jaHoldem from './locales/ja/holdem.json';
import jaIndianpoker from './locales/ja/indianpoker.json';
import jaJokerpoker from './locales/ja/jokerpoker.json';
import jaKlondike from './locales/ja/klondike.json';
import jaMemory from './locales/ja/memory.json';
import jaNapoleon from './locales/ja/napoleon.json';
import jaOhhell from './locales/ja/ohhell.json';
import jaOldmaid from './locales/ja/oldmaid.json';
import jaOmaha from './locales/ja/omaha.json';
import jaPineapple from './locales/ja/pineapple.json';
import jaPinochle from './locales/ja/pinochle.json';
import jaPoker from './locales/ja/poker.json';
import jaPyramid from './locales/ja/pyramid.json';
import jaSevens from './locales/ja/sevens.json';
import jaShortdeck from './locales/ja/shortdeck.json';
import jaSpades from './locales/ja/spades.json';
import jaSpeed from './locales/ja/speed.json';
import jaSpider from './locales/ja/spider.json';
import jaThreecard from './locales/ja/threecard.json';
import jaTripeaks from './locales/ja/tripeaks.json';
import jaTutorial from './locales/ja/tutorial.json';
import jaVideopoker from './locales/ja/videopoker.json';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      ja: {
        common: jaCommon,
        blackjack: jaBlackjack,
        poker: jaPoker,
        oldmaid: jaOldmaid,
        daifugo: jaDaifugo,
        sevens: jaSevens,
        doubt: jaDoubt,
        euchre: jaEuchre,
        bridge: jaBridge,
        holdem: jaHoldem,
        omaha: jaOmaha,
        pineapple: jaPineapple,
        pinochle: jaPinochle,
        shortdeck: jaShortdeck,
        hearts: jaHearts,
        spades: jaSpades,
        napoleon: jaNapoleon,
        ohhell: jaOhhell,
        memory: jaMemory,
        klondike: jaKlondike,
        freecell: jaFreecell,
        baccarat: jaBaccarat,
        crazyeights: jaCrazyeights,
        ginrummy: jaGinrummy,
        canasta: jaCanasta,
        cribbage: jaCribbage,
        spider: jaSpider,
        indianpoker: jaIndianpoker,
        pyramid: jaPyramid,
        threecard: jaThreecard,
        tripeaks: jaTripeaks,
        videopoker: jaVideopoker,
        deuceswild: jaDeuceswild,
        jokerpoker: jaJokerpoker,
        speed: jaSpeed,
        gofish: jaGofish,
        tutorial: jaTutorial,
      },
      en: {
        common: enCommon,
        blackjack: enBlackjack,
        poker: enPoker,
        oldmaid: enOldmaid,
        daifugo: enDaifugo,
        sevens: enSevens,
        doubt: enDoubt,
        euchre: enEuchre,
        bridge: enBridge,
        holdem: enHoldem,
        omaha: enOmaha,
        pineapple: enPineapple,
        pinochle: enPinochle,
        shortdeck: enShortdeck,
        hearts: enHearts,
        spades: enSpades,
        napoleon: enNapoleon,
        ohhell: enOhhell,
        memory: enMemory,
        klondike: enKlondike,
        freecell: enFreecell,
        baccarat: enBaccarat,
        crazyeights: enCrazyeights,
        ginrummy: enGinrummy,
        canasta: enCanasta,
        cribbage: enCribbage,
        spider: enSpider,
        indianpoker: enIndianpoker,
        pyramid: enPyramid,
        threecard: enThreecard,
        tripeaks: enTripeaks,
        videopoker: enVideopoker,
        deuceswild: enDeuceswild,
        jokerpoker: enJokerpoker,
        speed: enSpeed,
        gofish: enGofish,
        tutorial: enTutorial,
      },
    },
    fallbackLng: 'ja',
    defaultNS: 'common',
    ns: [
      'common',
      'blackjack',
      'poker',
      'oldmaid',
      'daifugo',
      'sevens',
      'doubt',
      'euchre',
      'bridge',
      'holdem',
      'omaha',
      'pineapple',
      'pinochle',
      'shortdeck',
      'hearts',
      'spades',
      'napoleon',
      'ohhell',
      'memory',
      'klondike',
      'freecell',
      'baccarat',
      'crazyeights',
      'ginrummy',
      'canasta',
      'cribbage',
      'spider',
      'indianpoker',
      'pyramid',
      'threecard',
      'tripeaks',
      'videopoker',
      'deuceswild',
      'jokerpoker',
      'speed',
      'gofish',
      'tutorial',
    ],
    detection: {
      order: ['localStorage'],
      lookupLocalStorage: 'i18n_lang',
      caches: ['localStorage'],
    },
    interpolation: {
      escapeValue: false,
    },
  });

/** Configured i18next instance with ja/en translations and language detection. */
export default i18n;
