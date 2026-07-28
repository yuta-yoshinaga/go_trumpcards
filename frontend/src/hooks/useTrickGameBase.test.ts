import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { heartsApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { HeartsResponse } from '../types/card';
import { useTrickGameBase } from './useTrickGameBase';

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

const DEFAULT_CONFIG = { cpuDifficulty: 1, pointLimit: 100, omnibusJD: false };

function renderBase() {
  return renderHook(
    () =>
      useTrickGameBase({
        apiFn: heartsApi.exec,
        defaultConfig: DEFAULT_CONFIG,
        getHint: (s) => s.hint ?? null,
      }),
    { wrapper: createWrapper() },
  );
}

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useTrickGameBase', () => {
  it('calls reset on mount with defaultConfig', async () => {
    renderBase();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, DEFAULT_CONFIG));
  });

  it('returns state after successful reset', async () => {
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handlePlay dispatches play with single selected card', async () => {
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(2);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePlay();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 2));
  });

  it('handlePlay does nothing when no card selected', async () => {
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handlePlay();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay does nothing when multiple cards selected', async () => {
    const { result } = renderBase();
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
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextTrick();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('clears selection and hint on success', async () => {
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });
    expect(result.current.selectedCardIndices).toEqual([0, 1]);

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.exec('next');
    });

    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hint from response', async () => {
    const stateWithHint: HeartsResponse = {
      ...defaultState,
      hint: { cardIndices: [1], reason: 'test reason' },
    };
    mockExec.mockResolvedValue(stateWithHint);

    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual({ cardIndices: [1], reason: 'test reason' });
    expect(result.current.hintError).toBeNull();
    expect(result.current.hintLoading).toBe(false);
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hintLoading).toBe(false);
  });

  it('handleConfigChange updates numeric config field', async () => {
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', '200');
    });

    expect(result.current.config.pointLimit).toBe(200);
  });

  it('handleToggle updates boolean config field', async () => {
    const { result } = renderBase();
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleToggle('omnibusJD', true);
    });

    expect(result.current.config.omnibusJD).toBe(true);
  });

  // `getHint` runs AFTER the mounted check — see the sibling test in
  // useSolitaireGameBase.test.ts for why it, not the returned hint, is the observable.
  it('does not process a hint that arrives after unmount', async () => {
    let resolveHint: ((value: HeartsResponse) => void) | undefined;
    mockExec.mockResolvedValue(defaultState);
    const getHint = vi.fn((s: HeartsResponse) => s.hint ?? null);
    const { result, unmount } = renderHook(
      () =>
        useTrickGameBase({
          apiFn: heartsApi.exec,
          defaultConfig: DEFAULT_CONFIG,
          getHint,
        }),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockImplementation(
      () =>
        new Promise<HeartsResponse>((resolve) => {
          resolveHint = resolve;
        }),
    );
    act(() => {
      void result.current.handleHint();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));

    unmount();
    resolveHint?.(defaultState);
    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });

    expect(getHint).not.toHaveBeenCalled();
  });
});
