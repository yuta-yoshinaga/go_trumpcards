import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { congressApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { CongressResponse } from '../types/card';
import { useCongressGame } from './useCongressGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  congressApi: { exec: vi.fn() },
  actionLogApi: { congress: vi.fn() },
}));

const mockExec = vi.mocked(congressApi.exec);

const baseState: CongressResponse = {
  tableau: Array.from({ length: 8 }, () => []),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 96,
  waste: [],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useCongressGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    act(() => result.current.handleDraw());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));

    act(() => result.current.handleGiveUp());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));

    act(() => result.current.handleUndo());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));

    act(() => result.current.handleAutoComplete());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores the hint payload from the API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 2 },
    });
    const { result } = renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 2 });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource distinguishes piles and toggles off', async () => {
    const { result } = renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 1 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 2 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  // The stock is a move source too, not only the draw button.
  it('handleSelectSource toggles the stock and the waste', async () => {
    const { result } = renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'stock' }));
    expect(result.current.selectedSource).toEqual({ zone: 'stock' });

    act(() => result.current.handleSelectSource({ zone: 'stock' }));
    expect(result.current.selectedSource).toBeNull();

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toEqual({ zone: 'waste' });
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget fills a gap from the stock through the move API', async () => {
    const { result } = renderHook(() => useCongressGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'stock' }));
    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 3 }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'tableau', col: 3 }));
    expect(result.current.selectedSource).toBeNull();
  });
});
