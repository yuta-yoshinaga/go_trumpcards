import i18n from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';

import enBaccarat from './locales/en/baccarat.json';
import enBlackjack from './locales/en/blackjack.json';
import enCommon from './locales/en/common.json';
import enDaifugo from './locales/en/daifugo.json';
import enDoubt from './locales/en/doubt.json';
import enFreecell from './locales/en/freecell.json';
import enHearts from './locales/en/hearts.json';
import enHoldem from './locales/en/holdem.json';
import enKlondike from './locales/en/klondike.json';
import enMemory from './locales/en/memory.json';
import enOldmaid from './locales/en/oldmaid.json';
import enOmaha from './locales/en/omaha.json';
import enPoker from './locales/en/poker.json';
import enSevens from './locales/en/sevens.json';
import enSpades from './locales/en/spades.json';
import jaBaccarat from './locales/ja/baccarat.json';
import jaBlackjack from './locales/ja/blackjack.json';
import jaCommon from './locales/ja/common.json';
import jaDaifugo from './locales/ja/daifugo.json';
import jaDoubt from './locales/ja/doubt.json';
import jaFreecell from './locales/ja/freecell.json';
import jaHearts from './locales/ja/hearts.json';
import jaHoldem from './locales/ja/holdem.json';
import jaKlondike from './locales/ja/klondike.json';
import jaMemory from './locales/ja/memory.json';
import jaOldmaid from './locales/ja/oldmaid.json';
import jaOmaha from './locales/ja/omaha.json';
import jaPoker from './locales/ja/poker.json';
import jaSevens from './locales/ja/sevens.json';
import jaSpades from './locales/ja/spades.json';

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
        holdem: jaHoldem,
        omaha: jaOmaha,
        hearts: jaHearts,
        spades: jaSpades,
        memory: jaMemory,
        klondike: jaKlondike,
        freecell: jaFreecell,
        baccarat: jaBaccarat,
      },
      en: {
        common: enCommon,
        blackjack: enBlackjack,
        poker: enPoker,
        oldmaid: enOldmaid,
        daifugo: enDaifugo,
        sevens: enSevens,
        doubt: enDoubt,
        holdem: enHoldem,
        omaha: enOmaha,
        hearts: enHearts,
        spades: enSpades,
        memory: enMemory,
        klondike: enKlondike,
        freecell: enFreecell,
        baccarat: enBaccarat,
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
      'holdem',
      'omaha',
      'hearts',
      'spades',
      'memory',
      'klondike',
      'freecell',
      'baccarat',
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

export default i18n;
