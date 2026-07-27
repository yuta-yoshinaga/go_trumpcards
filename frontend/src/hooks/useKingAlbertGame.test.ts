import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { kingAlbertApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { KingAlbertResponse } from '../types/card';
import { useKingAlbertGame } from './useKingAlbertGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  kingAlbertApi: { exec: vi.fn() },
  actionLogApi: { kingalbert: vi.fn() },
}));

const mockExec = vi.mocked(kingAlbertApi.exec);

const baseState: KingAlbertResponse = {
  tableau: Array.from({ length: 9 }, () => []),
  reserve: Array.from({ length: 7 }, () => null),
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useKingAlbertGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleReset / handleGiveUp / handleUndo forward to exec', async () => {
    const { result } = renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
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
    const { result } = renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleAutoComplete());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores hint payload from API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'reserve', fromCol: 0, cardIndex: 0, toZone: 'tableau', toCol: 1 },
    });
    const { result } = renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({
      fromZone: 'reserve',
      fromCol: 0,
      cardIndex: 0,
      toZone: 'tableau',
      toCol: 1,
    });
  });

  it('handleHint sets hintError when API rejects', async () => {
    const { result } = renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource toggles selection', async () => {
    const { result } = renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'reserve', col: 0 }));
    expect(result.current.selectedSource).toEqual({ zone: 'reserve', col: 0 });

    // Re-selecting the same zone clears the source.
    act(() => result.current.handleSelectSource({ zone: 'reserve', col: 0 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget dispatches reserve-to-tableau move when source is set', async () => {
    const { result } = renderHook(() => useKingAlbertGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'reserve', col: 2 }));
    act(() => result.current.handleSelectTarget({ zone: 'tableau', col: 1 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'reserve', col: 2 }, { zone: 'tableau', col: 1 }),
    );
    expect(result.current.selectedSource).toBeNull();
  });
});
