import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { describe, expect, it, vi } from 'vitest';

// Every wrapper hook under test pulls its api object off this module, so the
// mock has to cover all of them at once. The hooks only ever call `exec`, and
// this suite never asserts on API traffic -- it inspects the shape of the
// object each hook returns.
vi.mock('../api/gameApi', () => {
  // Exact export names, not derived from the game name -- a few are camelCased
  // (callBreakApi, twoTenJackApi) while the rest are all-lowercase.
  const names = [
    'beloteApi',
    'callBreakApi',
    'catchtenApi',
    'bauernschnapsenApi',
    'gaigelApi',
    'gongzhuApi',
    'heartsApi',
    'jassApi',
    'spadesApi',
    'tarneebApi',
    'tressetteApi',
    'twoTenJackApi',
    'whistApi',
    'bakersgameApi',
    'eightoffApi',
    'freecellApi',
    'klondikeApi',
    'penguinApi',
    'pyramidApi',
    'seahaventowersApi',
    'spiderApi',
    'spideretteApi',
    'tripeaksApi',
    'hintApi',
  ];
  const mod: Record<string, unknown> = {
    actionLogApi: new Proxy({}, { get: () => vi.fn() }),
  };
  for (const n of names) {
    mod[n] = { exec: vi.fn().mockResolvedValue(null), hint: vi.fn().mockResolvedValue(null) };
  }
  return mod;
});

import { useBakersGameGame } from './useBakersGameGame';
import { useBauernschnapsenGame } from './useBauernschnapsenGame';
import { useBeloteGame } from './useBeloteGame';
import { useCallBreakGame } from './useCallBreakGame';
import { useCatchTenGame } from './useCatchTenGame';
import { useEightOffGame } from './useEightOffGame';
import { useFreeCellGame } from './useFreeCellGame';
import { useGaigelGame } from './useGaigelGame';
import { useGongZhuGame } from './useGongZhuGame';
import { useHeartsGame } from './useHeartsGame';
import { useJassGame } from './useJassGame';
import { useKlondikeGame } from './useKlondikeGame';
import { usePenguinGame } from './usePenguinGame';
import { usePyramidGame } from './usePyramidGame';
import { useSeahavenTowersGame } from './useSeahavenTowersGame';
import { useSpadesGame } from './useSpadesGame';
import { useSpideretteGame } from './useSpideretteGame';
import { useSpiderGame } from './useSpiderGame';
import { useTarneebGame } from './useTarneebGame';
import { useTressetteGame } from './useTressetteGame';
import { useTriPeaksGame } from './useTriPeaksGame';
import { useTwoTenJackGame } from './useTwoTenJackGame';
import { useWhistGame } from './useWhistGame';

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return createElement(QueryClientProvider, { client }, children);
}

function keysOf(hook: () => object): string[] {
  const { result } = renderHook(hook, { wrapper });
  return Object.keys(result.current as object).sort();
}

/**
 * The wrappers around useTrickGameBase / useSolitaireGameBase re-exported the
 * base result field by field, which meant a field added to a base hook was
 * invisible to a game page until it was transcribed into that game's wrapper by
 * hand. The transcription had already drifted: useWhistGame exposed
 * clearSelection but not handleToggle, useBeloteGame the reverse -- both use the
 * same base and neither difference reflects a game rule.
 *
 * These tests pin the contract that a wrapper forwards its base's entire
 * surface, renaming only what it means to rename. See issue #4364.
 */
describe('trick-game hooks expose the full useTrickGameBase surface', () => {
  // Fields every useTrickGameBase consumer must be able to reach. `config` is
  // deliberately absent: each wrapper republishes it under a game-specific name.
  const BASE_FIELDS = [
    'clearSelection',
    'error',
    'exec',
    'handleConfigChange',
    'handleHint',
    'handleNextRound',
    'handleNextTrick',
    'handlePlay',
    'handleToggle',
    'hint',
    'hintError',
    'hintLoading',
    'loading',
    'retry',
    'selectedCardIndices',
    'state',
    'toggleCard',
  ] as const;

  it.each([
    ['useBeloteGame', useBeloteGame, 'beloteConfig'],
    ['useCallBreakGame', useCallBreakGame, 'callBreakConfig'],
    ['useCatchTenGame', useCatchTenGame, 'catchtenConfig'],
    ['useBauernschnapsenGame', useBauernschnapsenGame, 'bauernschnapsenConfig'],
    ['useGaigelGame', useGaigelGame, 'gaigelConfig'],
    ['useGongZhuGame', useGongZhuGame, 'gongzhuConfig'],
    ['useHeartsGame', useHeartsGame, 'heartsConfig'],
    ['useJassGame', useJassGame, 'jassConfig'],
    ['useSpadesGame', useSpadesGame, 'spadesConfig'],
    ['useTarneebGame', useTarneebGame, 'tarneebConfig'],
    ['useTressetteGame', useTressetteGame, 'tressetteConfig'],
    ['useTwoTenJackGame', useTwoTenJackGame, 'twoTenJackConfig'],
    ['useWhistGame', useWhistGame, 'whistConfig'],
  ])('%s forwards every base field and renames config', (_name, hook, configKey) => {
    const keys = keysOf(hook as () => object);
    for (const field of BASE_FIELDS) {
      expect(keys).toContain(field);
    }
    expect(keys).toContain(configKey);
    // The rename must be a rename, not an addition.
    expect(keys).not.toContain('config');
  });
});

describe('solitaire hooks expose the full useSolitaireGameBase surface', () => {
  // `apiCall` is absent by design: the wrappers republish it as `exec` to match
  // the naming every other game hook uses.
  const BASE_FIELDS = [
    'error',
    'exec',
    'handleAutoComplete',
    'handleGiveUp',
    'handleHint',
    'handleReset',
    'handleUndo',
    'hint',
    'hintError',
    'isAutoCompleting',
    'loading',
    'retry',
    'runAction',
    'setHint',
    'startAutoComplete',
    'state',
  ] as const;

  it.each([
    ['useBakersGameGame', useBakersGameGame],
    ['useEightOffGame', useEightOffGame],
    ['useFreeCellGame', useFreeCellGame],
    ['useKlondikeGame', useKlondikeGame],
    ['usePenguinGame', usePenguinGame],
    ['usePyramidGame', usePyramidGame],
    ['useSeahavenTowersGame', useSeahavenTowersGame],
    ['useSpiderGame', useSpiderGame],
    ['useSpideretteGame', useSpideretteGame],
    ['useTriPeaksGame', useTriPeaksGame],
  ])('%s forwards every base field', (_name, hook) => {
    const keys = keysOf(hook as () => object);
    for (const field of BASE_FIELDS) {
      expect(keys).toContain(field);
    }
    expect(keys).not.toContain('apiCall');
  });
});
