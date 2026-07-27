import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gaigelApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { GaigelResponse } from '../types/card';
import { GaigelPhase } from '../types/phases';
import { DEFAULT_GAIGEL_CONFIG, useGaigelGame } from './useGaigelGame';

vi.mock('../api/gameApi', () => ({
  gaigelApi: { exec: vi.fn() },
  actionLogApi: { gaigel: vi.fn() },
}));

const mockExec = vi.mocked(gaigelApi.exec);

function makeState(overrides: Partial<GaigelResponse> = {}): GaigelResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: GaigelPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    stockRemaining: 28,
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundMarriage: [0, 0],
    marriageIndices: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 101 },
    ...overrides,
  };
}

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('useGaigelGame', () => {
  it('dispatches reset on mount with the default config', async () => {
    renderHook(() => useGaigelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, DEFAULT_GAIGEL_CONFIG));
  });

  it('handleMarriage dispatches marriage with the card index', async () => {
    const { result } = renderHook(() => useGaigelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleMarriage(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('marriage', undefined, 2));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useGaigelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play with the selected card index', async () => {
    const { result } = renderHook(() => useGaigelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 2));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useGaigelGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    act(() => result.current.handleConfigChange('targetScore', '201'));
    mockExec.mockClear();
    act(() => {
      void (result.current.exec as unknown as (c: string, a1?: number, ci?: number, cfg?: unknown) => Promise<void>)(
        'reset',
        undefined,
        undefined,
        result.current.gaigelConfig,
      );
    });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        ...DEFAULT_GAIGEL_CONFIG,
        targetScore: 201,
      }),
    );
  });
});
