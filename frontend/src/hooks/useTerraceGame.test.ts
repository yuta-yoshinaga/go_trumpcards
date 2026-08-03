import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { terraceApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { TerraceResponse } from '../types/card';
import { useTerraceGame } from './useTerraceGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  terraceApi: { exec: vi.fn() },
  actionLogApi: { terrace: vi.fn() },
}));

const mockExec = vi.mocked(terraceApi.exec);

const baseState: TerraceResponse = {
  reserve: [],
  tableau: Array.from({ length: 9 }, () => []),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 84,
  waste: [],
  baseRank: 5,
  awaitingBaseRank: false,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useTerraceGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
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
    const { result } = renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores the hint payload from the API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'reserve', fromIdx: -1, toZone: 'foundation', toIdx: 2 },
    });
    const { result } = renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ fromZone: 'reserve', fromIdx: -1, toZone: 'foundation', toIdx: 2 });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource distinguishes piles and toggles off', async () => {
    const { result } = renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 1 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 2 });

    act(() => result.current.handleSelectSource({ zone: 'tableau', col: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource toggles the terrace and the waste', async () => {
    const { result } = renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'reserve' }));
    expect(result.current.selectedSource).toEqual({ zone: 'reserve' });

    act(() => result.current.handleSelectSource({ zone: 'reserve' }));
    expect(result.current.selectedSource).toBeNull();

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toEqual({ zone: 'waste' });
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget sends the terrace to a foundation through the move API', async () => {
    const { result } = renderHook(() => useTerraceGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'reserve' }));
    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 0 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve' }, { zone: 'foundation', col: 0 }),
    );
    expect(result.current.selectedSource).toBeNull();
  });
});
