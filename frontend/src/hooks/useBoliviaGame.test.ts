import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { boliviaApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeBoliviaState } from '../test/stateFactories';
import { useBoliviaGame } from './useBoliviaGame';

vi.mock('../api/gameApi', () => ({
  boliviaApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(boliviaApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState = makeBoliviaState();

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useBoliviaGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 10000 }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDrawStock dispatches drawstock command', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawStock();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('handleDrawDiscard dispatches drawdiscard with the two selected indices', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawDiscard();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard', undefined, undefined, [0, 1]));
  });

  it('handleDrawDiscard does nothing unless exactly 2 cards are selected', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(0);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleDrawDiscard();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleMeldSelected dispatches meld with the selected group', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
      result.current.toggleCard(2);
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleMeldSelected();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, undefined, [[0, 1, 2]]));
  });

  it('handleMeldSelected does nothing with fewer than 3 cards', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleMeldSelected();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleSkipMeld dispatches skipmeld command', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleSkipMeld();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('handleDiscard dispatches discard with the single selected card', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(2);
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDiscard();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 2));
  });

  it('handleDiscard does nothing when no card is selected', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleDiscard();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleGoOut dispatches goout command', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleGoOut();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('goout'));
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config with a valid number', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.handleConfigChange('pointLimit', '7500');
    });
    expect(result.current.boliviaConfig.pointLimit).toBe(7500);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.handleConfigChange('pointLimit', 'abc');
    });
    expect(result.current.boliviaConfig.pointLimit).toBe(10000);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => useBoliviaGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(0);
    });
    expect(result.current.selectedCardIndices).toEqual([0]);
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawStock();
    });
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
