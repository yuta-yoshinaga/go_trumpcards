import '@testing-library/jest-dom/vitest';
import { cleanup, configure } from '@testing-library/react';
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { afterEach, beforeAll, vi } from 'vitest';

// Mock ResizeObserver for happy-dom (needed by TutorialOverlay)
if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = class ResizeObserver {
    observe = vi.fn();
    unobserve = vi.fn();
    disconnect = vi.fn();
  } as unknown as typeof globalThis.ResizeObserver;
}

// Mock matchMedia for happy-dom (needed by useReducedMotion)
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Global framer-motion mock: render motion.* as plain HTML elements
vi.mock('framer-motion', async () => {
  const React = await import('react');
  function createMotionProxy() {
    return new Proxy(
      {},
      {
        get: (_target: Record<string, unknown>, prop: string) =>
          React.forwardRef((props: Record<string, unknown>, ref: React.Ref<HTMLElement>) => {
            const {
              initial: _i,
              animate: _a,
              exit: _e,
              transition: _t,
              whileHover: _wh,
              whileTap: _wt,
              layout: _l,
              layoutId: _li,
              ...rest
            } = props;
            return React.createElement(prop, { ...rest, ref });
          }),
      },
    );
  }
  const AnimatePresence = ({ children }: { children: React.ReactNode }) =>
    React.createElement(React.Fragment, null, children);
  return { motion: createMotionProxy(), AnimatePresence };
});

import enCommon from '../i18n/locales/en/common.json';
import jaBaccarat from '../i18n/locales/ja/baccarat.json';
import jaBlackjack from '../i18n/locales/ja/blackjack.json';
import jaCommon from '../i18n/locales/ja/common.json';
import jaCrazyeights from '../i18n/locales/ja/crazyeights.json';
import jaDaifugo from '../i18n/locales/ja/daifugo.json';
import jaDoubt from '../i18n/locales/ja/doubt.json';
import jaFreecell from '../i18n/locales/ja/freecell.json';
import jaGinrummy from '../i18n/locales/ja/ginrummy.json';
import jaHearts from '../i18n/locales/ja/hearts.json';
import jaHoldem from '../i18n/locales/ja/holdem.json';
import jaIndianpoker from '../i18n/locales/ja/indianpoker.json';
import jaKlondike from '../i18n/locales/ja/klondike.json';
import jaMemory from '../i18n/locales/ja/memory.json';
import jaNapoleon from '../i18n/locales/ja/napoleon.json';
import jaOldmaid from '../i18n/locales/ja/oldmaid.json';
import jaOmaha from '../i18n/locales/ja/omaha.json';
import jaPoker from '../i18n/locales/ja/poker.json';
import jaSevens from '../i18n/locales/ja/sevens.json';
import jaSpades from '../i18n/locales/ja/spades.json';
import jaSpeed from '../i18n/locales/ja/speed.json';
import jaSpider from '../i18n/locales/ja/spider.json';
import jaThreecard from '../i18n/locales/ja/threecard.json';
import jaTutorial from '../i18n/locales/ja/tutorial.json';

i18n.use(initReactI18next).init({
  lng: 'ja',
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
    'baccarat',
    'crazyeights',
    'klondike',
    'freecell',
    'ginrummy',
    'napoleon',
    'spider',
    'indianpoker',
    'threecard',
    'speed',
    'tutorial',
  ],
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
      baccarat: jaBaccarat,
      crazyeights: jaCrazyeights,
      klondike: jaKlondike,
      freecell: jaFreecell,
      ginrummy: jaGinrummy,
      napoleon: jaNapoleon,
      spider: jaSpider,
      indianpoker: jaIndianpoker,
      threecard: jaThreecard,
      speed: jaSpeed,
      tutorial: jaTutorial,
    },
    en: {
      common: enCommon,
    },
  },
  interpolation: { escapeValue: false },
});

configure({ asyncUtilTimeout: 5000 });

// Suppress first-visit tutorial suggestion dialog in all tests by default.
// Individual tests (e.g., useFirstVisit.test.ts) clear localStorage before running.
beforeAll(() => {
  localStorage.setItem('tutorial_no_suggest', 'true');
});

afterEach(() => {
  cleanup();
});
