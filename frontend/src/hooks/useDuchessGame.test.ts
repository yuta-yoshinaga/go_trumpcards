import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { duchessApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { DuchessResponse } from '../types/card';
import { useDuchessGame } from './useDuchessGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  duchessApi: { exec: vi.fn() },
  actionLogApi: { duchess: vi.fn() },
}));

const mockExec = vi.mocked(duchessApi.exec);

const baseState: DuchessResponse = {
  reserve: Array.from({ length: 4 }, () => []),
  tableau: Array.from({ length: 4 }, () => []),
  foundation: Array.from({ length: 4 }, () => []),
  stockCount: 35,
  waste: [],
  baseRank: 5,
  awaitingBaseRank: false,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useDuchessGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
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

  // Choosing the base rank is its own command, not a move — it has no destination.
  it('handleChooseBase sends the fan index as a reserve zone', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleChooseBase(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('base', { zone: 'reserve', col: 2 }));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores the hint payload from the API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'reserve', fromIdx: 1, cardIndex: -1, toZone: 'tableau', toIdx: 2 },
    });
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({
      fromZone: 'reserve',
      fromIdx: 1,
      cardIndex: -1,
      toZone: 'tableau',
      toIdx: 2,
    });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource distinguishes cards within a column', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 1 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 2 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource distinguishes reserve fans', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'reserve', col: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'reserve', col: 1 });

    act(() => result.current.handleSelectSource({ zone: 'reserve', col: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'reserve', col: 2 });

    act(() => result.current.handleSelectSource({ zone: 'reserve', col: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource toggles the waste zone', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toEqual({ zone: 'waste' });

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget moves a reserve card through the move API', async () => {
    const { result } = renderHook(() => useDuchessGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'reserve', col: 1 }));
    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 3 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve', col: 1 }, { zone: 'tableau', col: 3 }),
    );
    expect(result.current.selectedSource).toBeNull();
  });
});
