import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { diplomatApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { DiplomatResponse } from '../types/card';
import { useDiplomatGame } from './useDiplomatGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  diplomatApi: { exec: vi.fn() },
  actionLogApi: { diplomat: vi.fn() },
}));

const mockExec = vi.mocked(diplomatApi.exec);

const baseState: DiplomatResponse = {
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

describe('useDiplomatGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
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
    const { result } = renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
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
    const { result } = renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 2 });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource distinguishes piles and toggles off', async () => {
    const { result } = renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 1 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 2 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  // The stock is draw-only in Diplomat, so the two sources are the tableau and
  // the waste. (Congress also lets the stock be a source; this one does not.)
  it('handleSelectSource toggles a column and the waste', async () => {
    const { result } = renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    expect(result.current.selectedSource).toBeNull();

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toEqual({ zone: 'waste' });
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // A column into an empty column is the move Diplomat adds and Congress
  // forbids, so it is the one worth pinning here.
  it('handleSelectTarget moves a column top into another column', async () => {
    const { result } = renderHook(() => useDiplomatGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0 }));
    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 3 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }),
    );
    expect(result.current.selectedSource).toBeNull();
  });
});
