import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bauernschnapsenApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { BauernschnapsenResponse } from '../types/card';
import { BauernschnapsenPhase } from '../types/phases';
import { DEFAULT_BAUERNSCHNAPSEN_CONFIG, useBauernschnapsenGame } from './useBauernschnapsenGame';

vi.mock('../api/gameApi', () => ({
  bauernschnapsenApi: { exec: vi.fn() },
  actionLogApi: { bauernschnapsen: vi.fn() },
}));

const mockExec = vi.mocked(bauernschnapsenApi.exec);

function makeState(overrides: Partial<BauernschnapsenResponse> = {}): BauernschnapsenResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: BauernschnapsenPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    contract: 1,
    declarerIdx: 0,
    validPlayIndices: [0, 1, 2, 3, 4],
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundMarriage: [0, 0],
    marriageIndices: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 24 },
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

describe('useBauernschnapsenGame', () => {
  it('dispatches reset on mount with the default config', async () => {
    renderHook(() => useBauernschnapsenGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, DEFAULT_BAUERNSCHNAPSEN_CONFIG),
    );
  });

  it('handleMarriage dispatches marriage with the card index', async () => {
    const { result } = renderHook(() => useBauernschnapsenGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleMarriage(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('marriage', undefined, 2));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useBauernschnapsenGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play with the selected card index', async () => {
    const { result } = renderHook(() => useBauernschnapsenGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 2));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useBauernschnapsenGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    act(() => result.current.handleConfigChange('targetScore', '201'));
    mockExec.mockClear();
    act(() => {
      void (result.current.exec as unknown as (c: string, a1?: number, ci?: number, cfg?: unknown) => Promise<void>)(
        'reset',
        undefined,
        undefined,
        result.current.bauernschnapsenConfig,
      );
    });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        ...DEFAULT_BAUERNSCHNAPSEN_CONFIG,
        targetScore: 201,
      }),
    );
  });
});
