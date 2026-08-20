import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { heartsApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { HeartsResponse } from '../types/card';
import { useHeartsGame } from './useHeartsGame';

vi.mock('../api/gameApi', () => ({
  heartsApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(heartsApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: HeartsResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      penaltyCards: [],
      tookOmnibusJD: false,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      penaltyCards: [],
      tookOmnibusJD: false,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 13,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      penaltyCards: [],
      tookOmnibusJD: false,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 13,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      penaltyCards: [],
      tookOmnibusJD: false,
    },
  ],
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentTrick: [],
  heartsBroken: false,
  passDirection: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 100, omnibusJD: false },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useHeartsGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
        omnibusJD: false,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handlePass dispatches pass with selected indices', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
      result.current.toggleCard(2);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePass();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass', [0, 1, 2]));
  });

  it('handlePlay dispatches play with single selected card', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(3);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePlay();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 3));
  });

  it('handlePlay does nothing when no card selected', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handlePlay();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay does nothing when multiple cards selected', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });

    mockExec.mockClear();
    act(() => {
      result.current.handlePlay();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleNextTrick dispatches next command', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextTrick();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config with valid number', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', '200');
    });

    expect(result.current.heartsConfig.pointLimit).toBe(200);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', 'abc');
    });

    expect(result.current.heartsConfig.pointLimit).toBe(100);
  });

  it('handleToggle updates boolean config', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    expect(result.current.heartsConfig.omnibusJD).toBe(false);

    act(() => {
      result.current.handleToggle('omnibusJD', true);
    });

    expect(result.current.heartsConfig.omnibusJD).toBe(true);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => useHeartsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });
    expect(result.current.selectedCardIndices).toEqual([0, 1]);

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePass();
    });

    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
