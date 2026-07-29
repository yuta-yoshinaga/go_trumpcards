import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { grandfathersClockApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { GrandfathersClockResponse } from '../types/card';
import { useGrandfathersClockGame } from './useGrandfathersClockGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  grandfathersClockApi: { exec: vi.fn() },
  actionLogApi: { grandfathersclock: vi.fn() },
}));

const mockExec = vi.mocked(grandfathersClockApi.exec);

const baseState: GrandfathersClockResponse = {
  tableau: Array.from({ length: 8 }, () => []),
  foundation: Array.from({ length: 12 }, (_, i) => ({ cards: [], targetRank: i + 1, complete: false })),
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useGrandfathersClockGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    act(() => result.current.handleGiveUp());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));

    act(() => result.current.handleUndo());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));

    act(() => result.current.handleAutoComplete());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores the hint payload from the API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromCol: 1, toZone: 'foundation', toIdx: 4 },
    });
    const { result } = renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ fromCol: 1, toZone: 'foundation', toIdx: 4 });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource toggles selection', async () => {
    const { result } = renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('clears the selection and the hint on a board-resetting action', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromCol: 1, toZone: 'tableau', toIdx: 4 },
    });
    const { result } = renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    expect(result.current.hint).not.toBeNull();

    act(() => result.current.handleUndo());
    await waitFor(() => expect(result.current.hint).toBeNull());
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // The clock face index travels with the move; it cannot be derived server-side.
  it('handleSelectTarget dispatches move including the face index', async () => {
    const { result } = renderHook(() => useGrandfathersClockGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 7 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation', col: 7 }),
    );
    expect(result.current.selectedSource).toBeNull();
  });
});
