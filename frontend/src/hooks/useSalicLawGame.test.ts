import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { salicLawApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { SalicLawResponse } from '../types/card';
import { useSalicLawGame } from './useSalicLawGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  salicLawApi: { exec: vi.fn() },
  actionLogApi: { saliclaw: vi.fn() },
}));

const mockExec = vi.mocked(salicLawApi.exec);

const baseState: SalicLawResponse = {
  tableau: Array.from({ length: 8 }, () => []),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 95,
  queens: [],
  openPiles: 1,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useSalicLawGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
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
    const { result } = renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
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
    const { result } = renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 2 });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource distinguishes piles and toggles off', async () => {
    const { result } = renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 1 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 2 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  // **移動元はタブローの列だけ。**捨て札は無く、山札は配りボタンであって
  // 移動元ではない。列が違えばトグルではなく選び直しになる。
  it('handleSelectSource toggles one column and switches between columns', async () => {
    const { result } = renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 2 });

    // 同じ列をもう一度押すと解除。
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toBeNull();

    // 別の列を押すと選び直し（解除ではない）。
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 5 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 5 });
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget fills a gap from the stock through the move API', async () => {
    const { result } = renderHook(() => useSalicLawGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'stock' }));
    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 3 }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'tableau', col: 3 }));
    expect(result.current.selectedSource).toBeNull();
  });
});
