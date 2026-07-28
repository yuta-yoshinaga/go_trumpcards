import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { napoleonsSquareApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { NapoleonsSquareResponse } from '../types/card';
import { useNapoleonsSquareGame } from './useNapoleonsSquareGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  napoleonsSquareApi: { exec: vi.fn() },
  actionLogApi: { napoleonssquare: vi.fn() },
}));

const mockExec = vi.mocked(napoleonsSquareApi.exec);

const baseState: NapoleonsSquareResponse = {
  tableau: Array.from({ length: 12 }, () => []),
  stockCount: 48,
  waste: [],
  foundation: Array.from({ length: 8 }, () => []),
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useNapoleonsSquareGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
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
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores the hint payload from the API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'tableau', fromCol: 1, cardIndex: 2, toZone: 'tableau', toCol: 4 },
    });
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({
      fromZone: 'tableau',
      fromCol: 1,
      cardIndex: 2,
      toZone: 'tableau',
      toCol: 4,
    });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  // Selection is per card, not per column, because any card can head a run.
  it('handleSelectSource distinguishes cards within a column', async () => {
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 1 });

    // A different card in the same column replaces the selection rather than clearing it.
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 2 });

    // Re-selecting the same card clears it.
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  // The waste zone carries no col/cardIndex, so the toggle compares undefined
  // against undefined — a path the tableau cases never exercise.
  it('handleSelectSource toggles the waste zone', async () => {
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toEqual({ zone: 'waste' });

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('clears the selection and the hint on every board-resetting action', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'stock', fromCol: -1, cardIndex: -1, toZone: 'waste', toCol: -1 },
    });
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.hint).not.toBeNull();
    expect(result.current.selectedSource).not.toBeNull();

    act(() => result.current.handleDraw());
    await waitFor(() => expect(result.current.hint).toBeNull());
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget dispatches move when a source is set', async () => {
    const { result } = renderHook(() => useNapoleonsSquareGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 1 }));
    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 5 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 1 },
        { zone: 'tableau', col: 5 },
      ),
    );
    expect(result.current.selectedSource).toBeNull();
  });
});
