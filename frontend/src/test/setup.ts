import '@testing-library/jest-dom/vitest';
import { cleanup, configure } from '@testing-library/react';
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { afterEach, beforeEach, vi } from 'vitest';

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

// Global howler mock: keep the 200+ page/hook tests that render inside
// SoundProvider (via renderWithProviders) from hitting real Web Audio in
// jsdom/happy-dom once the central sound taps fire on every exec.
// Files that need to assert on Howl behavior re-mock 'howler' locally,
// which overrides this global mock.
vi.mock('howler', () => {
  class MockHowl {
    play = vi.fn(() => 1);
    volume = vi.fn();
    rate = vi.fn();
    stop = vi.fn();
    on = vi.fn();
    once = vi.fn();
    unload = vi.fn();
  }
  return {
    Howl: MockHowl,
    Howler: { ctx: { state: 'running' }, autoUnlock: true },
  };
});

import { buildResources } from '../i18n/buildResources';

// Eagerly import all locale JSON files via Vite glob
const jaModules = import.meta.glob('../i18n/locales/ja/*.json', {
  eager: true,
  import: 'default',
}) as Record<string, Record<string, string>>;

const enModules = import.meta.glob('../i18n/locales/en/*.json', {
  eager: true,
  import: 'default',
}) as Record<string, Record<string, string>>;

// Tests include the lazy namespaces eagerly so callers don't have to
// orchestrate dynamic imports inside jsdom — the runtime path that
// loads `discover` on /discover mount is covered by a dedicated test.
const jaResources = buildResources(jaModules, { skipLazy: false });
const enResources = buildResources(enModules, { skipLazy: false });

i18n.use(initReactI18next).init({
  lng: 'ja',
  fallbackLng: 'ja',
  defaultNS: 'common',
  ns: Object.keys(jaResources),
  resources: {
    ja: jaResources,
    en: enResources,
  },
  interpolation: { escapeValue: false },
});

configure({ asyncUtilTimeout: 5000 });

// Suppress first-visit tutorial suggestion dialog in all tests by default.
// Uses beforeEach so the flag survives tests that call localStorage.clear() in
// their own afterEach; individual tests that need the dialog can clear it
// inside their own beforeEach (runs after this one).
beforeEach(() => {
  localStorage.setItem('tutorial_no_suggest', 'true');
});

afterEach(() => {
  cleanup();
});
