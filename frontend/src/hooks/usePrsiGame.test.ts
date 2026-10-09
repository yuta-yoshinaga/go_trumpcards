import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { prsiApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { PrsiResponse } from '../types/card';
import { usePrsiGame } from './usePrsiGame';

vi.mock('../api/gameApi', () => ({
  prsiApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(prsiApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: PrsiResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 5, cards: [{ design: 'SPADE', value: 7 }] },
    { id: 1, isHuman: false, cardCount: 5, cards: [] },
    { id: 2, isHuman: false, cardCount: 5, cards: [] },
    { id: 3, isHuman: false, cardCount: 5, cards: [] },
  ],
  phase: 0,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 9 },
  drawPileCount: 30,
  discardPileCount: 3,
  penaltyDrawCount: 0,
  pendingSkips: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1 },
};

const playableState: PrsiResponse = {
  ...defaultState,
  players: defaultState.players.map((player) =>
    player.isHuman
      ? {
          ...player,
          cards: [
            { design: 'HEART', value: 10 },
            { design: 'SPADE', value: 7 },
          ],
        }
      : player,
  ),
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('usePrsiGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handlePlay dispatches play with single selected card', async () => {
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    mockExec.mockResolvedValue(playableState);
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePlay();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('does not play a selected card that is illegal on the discard pile', async () => {
    mockExec.mockResolvedValue(defaultState);
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => result.current.toggleCard(0));
    mockExec.mockClear();

    act(() => result.current.handlePlay());

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('allows only a 7 while a penalty is active', async () => {
    mockExec.mockResolvedValue({
      ...playableState,
      penaltyDrawCount: 2,
    });
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => result.current.toggleCard(0));
    expect(result.current.canPlay).toBe(false);
    act(() => result.current.toggleCard(0));
    act(() => result.current.toggleCard(1));
    expect(result.current.canPlay).toBe(true);
  });

  it('handlePlay does nothing when no card selected', async () => {
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handlePlay();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay does nothing when multiple cards selected', async () => {
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
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

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleConfigChange updates config with a valid number', async () => {
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('cpuDifficulty', '2');
    });

    expect(result.current.prsiConfig.cpuDifficulty).toBe(2);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('cpuDifficulty', 'abc');
    });

    expect(result.current.prsiConfig.cpuDifficulty).toBe(1);
  });

  it('clears selection on success', async () => {
    mockExec.mockResolvedValue(playableState);
    const { result } = renderHook(() => usePrsiGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
    });
    expect(result.current.selectedCardIndices).toEqual([0]);

    mockExec.mockResolvedValue(playableState);
    act(() => {
      result.current.handlePlay();
    });

    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
