import '@testing-library/jest-dom/vitest';
import { cleanup, configure } from '@testing-library/react';
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { afterEach } from 'vitest';

import jaBlackjack from '../i18n/locales/ja/blackjack.json';
import jaCommon from '../i18n/locales/ja/common.json';
import jaDaifugo from '../i18n/locales/ja/daifugo.json';
import jaDoubt from '../i18n/locales/ja/doubt.json';
import jaHearts from '../i18n/locales/ja/hearts.json';
import jaHoldem from '../i18n/locales/ja/holdem.json';
import jaOldmaid from '../i18n/locales/ja/oldmaid.json';
import jaPoker from '../i18n/locales/ja/poker.json';
import jaSevens from '../i18n/locales/ja/sevens.json';

i18n.use(initReactI18next).init({
  lng: 'ja',
  fallbackLng: 'ja',
  defaultNS: 'common',
  ns: ['common', 'blackjack', 'poker', 'oldmaid', 'daifugo', 'sevens', 'doubt', 'holdem', 'hearts'],
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
      hearts: jaHearts,
    },
  },
  interpolation: { escapeValue: false },
});

configure({ asyncUtilTimeout: 15000 });

afterEach(() => {
  cleanup();
});
