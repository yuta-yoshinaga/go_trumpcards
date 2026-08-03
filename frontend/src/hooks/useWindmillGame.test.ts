import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { windmillApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { WindmillResponse } from '../types/card';
import { useWindmillGame } from './useWindmillGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  windmillApi: { exec: vi.fn() },
  actionLogApi: { windmill: vi.fn() },
}));

const mockExec = vi.mocked(windmillApi.exec);

const baseState: WindmillResponse = {
  sails: Array.from({ length: 8 }, () => null),
  center: [],
  corners: Array.from({ length: 4 }, () => []),
  stockCount: 95,
  waste: [],
  transferBlocked: false,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useWindmillGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
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
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores the hint payload from the API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'corner', fromIdx: 1, toZone: 'center', toIdx: -1 },
    });
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ fromZone: 'corner', fromIdx: 1, toZone: 'center', toIdx: -1 });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource distinguishes sails from one another', async () => {
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'sail', col: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'sail', col: 1 });

    act(() => result.current.handleSelectSource({ zone: 'sail', col: 2 }));
    expect(result.current.selectedSource).toEqual({ zone: 'sail', col: 2 });

    act(() => result.current.handleSelectSource({ zone: 'sail', col: 2 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource toggles the waste zone', async () => {
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toEqual({ zone: 'waste' });

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'center' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget sends a sail to a corner through the move API', async () => {
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'sail', col: 3 }));
    act(() => result.current.handleSelectTarget({ zone: 'corner', col: 1 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'sail', col: 3 }, { zone: 'corner', col: 1 }),
    );
    expect(result.current.selectedSource).toBeNull();
  });

  // The pull-back is dispatched as an ordinary move with a corner source.
  it('handleSelectTarget sends a corner back to the centre', async () => {
    const { result } = renderHook(() => useWindmillGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'corner', col: 0 }));
    act(() => result.current.handleSelectTarget({ zone: 'center' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'corner', col: 0 }, { zone: 'center' }));
  });
});
