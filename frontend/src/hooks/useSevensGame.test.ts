import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { sevensApi } from '../api/gameApi';
import { asMocked } from '../test/viCompat';
import type { SevensResponse } from '../types/card';
import * as gameReplay from './gameReplay';
import { useSevensGame } from './useSevensGame';

let mockExec: ReturnType<typeof vi.fn>;
function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const baseState: SevensResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [],
      isFinished: false,
      rank: 0,
      passesUsed: 0,
      maxPasses: 5,
      lastPlayedJoker: false,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 5,
      cards: [],
      isFinished: false,
      rank: 0,
      passesUsed: 0,
      maxPasses: 5,
      lastPlayedJoker: false,
    },
  ],
  currentTurn: 0,
  tablePlaced: [0, 0, 0, 0, 0],
  tableMinVals: [0, 0, 0, 0, 0],
  tableMaxVals: [0, 0, 0, 0, 0],
  config: {
    tunnelEnabled: false,
    tunnelSkipWidth: 0,
    jokerCount: 0,
    cpuStrategy: 0,
    maxPasses: 5,
    noJokerFinish: false,
    jokerReclaimEnabled: false,
    endStopEnabled: false,
    jokerConsecutiveBanned: false,
  },
  gameEndFlag: false,
  cpuActions: [
    {
      playerIdx: 1,
      playedCard: { design: 'SPADE', value: 6 },
      targetSuit: 1,
      targetValue: 6,
      forcedPass: false,
    },
  ],
  humanAction: null,
  message: '',
};

describe('useSevensGame onSuccess replay skip', () => {
  let runReplaySpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(sevensApi, 'exec').mockImplementation(vi.fn());
    mockExec = asMocked(sevensApi.exec);
    runReplaySpy = vi.spyOn(gameReplay, 'runReplay').mockResolvedValue(undefined);
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('calls runReplay on first success (mount reset)', async () => {
    renderHook(() => useSevensGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));
  });

  it('skips replay when cpuActions unchanged', async () => {
    const { result } = renderHook(() => useSevensGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(baseState);
    act(() => {
      result.current.exec('play', 0);
    });
    await waitFor(() => expect(result.current.state).toBeDefined());
    expect(runReplaySpy).not.toHaveBeenCalled();
  });

  it('runs replay when cpuActions change', async () => {
    const { result } = renderHook(() => useSevensGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    const newState: SevensResponse = {
      ...baseState,
      cpuActions: [
        {
          playerIdx: 1,
          playedCard: { design: 'HEART', value: 8 },
          targetSuit: 2,
          targetValue: 8,
          forcedPass: false,
        },
      ],
    };
    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(newState);
    act(() => {
      result.current.exec('play', 0);
    });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));
  });

  it('sets displayState even when replay is skipped', async () => {
    const { result } = renderHook(() => useSevensGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    const updatedState: SevensResponse = { ...baseState, currentTurn: 1 };
    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(updatedState);
    act(() => {
      result.current.exec('play', 0);
    });
    await waitFor(() => expect(result.current.state?.currentTurn).toBe(1));
    expect(runReplaySpy).not.toHaveBeenCalled();
  });
});
