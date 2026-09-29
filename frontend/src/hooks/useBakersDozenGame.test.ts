import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { bakersDozenApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { BakersDozenResponse } from '../types/card';
import { useBakersDozenGame } from './useBakersDozenGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  bakersDozenApi: { exec: vi.fn() },
  actionLogApi: { bakersdozen: vi.fn() },
}));

const mockExec = vi.mocked(bakersDozenApi.exec);

const baseState: BakersDozenResponse = {
  tableau: Array.from({ length: 13 }, () => []),
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useBakersDozenGame', () => {
  beforeEach(() => {
    mockExec.mockReset();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    mockExec.mockReset();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleReset / handleGiveUp / handleUndo forward to exec', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    act(() => result.current.handleGiveUp());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));

    act(() => result.current.handleUndo());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleAutoComplete dispatches autocomplete', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleAutoComplete());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('announces in-progress again when retrying a failed auto-complete', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockRejectedValueOnce(new Error('network failure'));

    await act(async () => {
      await result.current.handleAutoComplete();
    });
    await waitFor(() => expect(result.current.autoCompleteStatus).toBe('interrupted'));

    let finishRetry: ((response: BakersDozenResponse) => void) | undefined;
    mockExec.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finishRetry = resolve;
        }),
    );
    let retryPromise: Promise<void> | undefined;
    act(() => {
      retryPromise = result.current.retry();
    });
    expect(result.current.autoCompleteStatus).toBe('inProgress');
    expect(mockExec).toHaveBeenLastCalledWith('autocomplete');
    await waitFor(() => expect(finishRetry).toBeDefined());
    await act(async () => {
      finishRetry?.({ ...baseState, phase: 1 });
      await retryPromise;
    });
  });

  it('clears the auto-complete status when resetting', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockResolvedValue({ ...baseState, phase: 1 });

    await act(async () => {
      result.current.handleAutoComplete();
      await Promise.resolve();
    });
    await waitFor(() => expect(result.current.autoCompleteStatus).toBe('finished'));

    await act(async () => {
      result.current.handleReset();
      await Promise.resolve();
    });
    await waitFor(() => expect(mockExec).toHaveBeenLastCalledWith('reset'));
    expect(result.current.autoCompleteStatus).toBeNull();
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores hint payload from API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 1 },
    });
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 1 });
  });

  it('handleHint sets hintError when API rejects', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource toggles selection', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 0 });

    // Selecting the same zone again clears it.
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 1 }));
    // No move call should have been issued.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget dispatches move when source is set', async () => {
    const { result } = renderHook(() => useBakersDozenGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 }));
    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 1 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'tableau', col: 1 },
      ),
    );
    expect(result.current.selectedSource).toBeNull();
  });
});
