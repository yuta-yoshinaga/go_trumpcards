import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { missMilliganApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { MissMilliganResponse } from '../types/card';
import { useMissMilliganGame } from './useMissMilliganGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  missMilliganApi: { exec: vi.fn() },
  actionLogApi: { missmilligan: vi.fn() },
}));

const mockExec = vi.mocked(missMilliganApi.exec);

const baseState: MissMilliganResponse = {
  tableau: Array.from({ length: 8 }, () => []),
  stockCount: 96,
  foundation: Array.from({ length: 8 }, () => []),
  waived: [],
  canWaive: false,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useMissMilliganGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    act(() => result.current.handleDeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));

    act(() => result.current.handleGiveUp());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));

    act(() => result.current.handleUndo());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));

    act(() => result.current.handleAutoComplete());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  // Waiving is its own command, not a move — it has no destination.
  it('handleWaive sends the column and defaults the run head', async () => {
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleWaive(3));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('waive', { zone: 'tableau', col: 3, cardIndex: undefined }),
    );

    act(() => result.current.handleWaive(3, 1));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('waive', { zone: 'tableau', col: 3, cardIndex: 1 }));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores the hint payload from the API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'waived', fromCol: -1, cardIndex: -1, toZone: 'tableau', toIdx: 2 },
    });
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({
      fromZone: 'waived',
      fromCol: -1,
      cardIndex: -1,
      toZone: 'tableau',
      toIdx: 2,
    });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource distinguishes cards within a column', async () => {
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 1 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 2 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource toggles the waived zone', async () => {
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'waived' }));
    expect(result.current.selectedSource).toEqual({ zone: 'waived' });

    act(() => result.current.handleSelectSource({ zone: 'waived' }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget moves the waived cards back through the move API', async () => {
    const { result } = renderHook(() => useMissMilliganGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'waived' }));
    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 5 }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waived' }, { zone: 'tableau', col: 5 }));
    expect(result.current.selectedSource).toBeNull();
  });
});
