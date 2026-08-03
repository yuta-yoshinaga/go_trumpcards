import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { braidApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { BraidResponse } from '../types/card';
import { useBraidGame } from './useBraidGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  braidApi: { exec: vi.fn() },
  actionLogApi: { braid: vi.fn() },
}));

const mockExec = vi.mocked(braidApi.exec);

const baseState: BraidResponse = {
  braid: [],
  fields: Array.from({ length: 4 }, () => null),
  helpers: Array.from({ length: 8 }, () => null),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 71,
  waste: [],
  baseRank: 5,
  direction: 1,
  awaitingDirection: false,
  redealsLeft: 2,
  canRedeal: false,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('useBraidGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the simple commands', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
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

  // The flag rides in the 5th argument, so a wrong slot silently sends
  // `ascending: undefined` and the backend rejects the whole command.
  it('handleChooseDirection sends the flag in the ascending slot', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleChooseDirection(true));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('dir', undefined, undefined, undefined, true));

    act(() => result.current.handleChooseDirection(false));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('dir', undefined, undefined, undefined, false));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleUndoEscape(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('handleHint stores the hint payload from the API', async () => {
    mockExec.mockResolvedValueOnce(baseState).mockResolvedValueOnce({
      ...baseState,
      hint: { fromZone: 'field', fromIdx: 2, toZone: 'foundation', toIdx: 1 },
    });
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual({ fromZone: 'field', fromIdx: 2, toZone: 'foundation', toIdx: 1 });
  });

  it('handleHint sets hintError when the API rejects', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockRejectedValueOnce(new Error('boom'));

    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).not.toBeNull();
  });

  it('handleSelectSource distinguishes slots and toggles off', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'field', col: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'field', col: 1 });

    // Same index, different zone -- these must not be treated as the same slot.
    act(() => result.current.handleSelectSource({ zone: 'helper', col: 1 }));
    expect(result.current.selectedSource).toEqual({ zone: 'helper', col: 1 });

    act(() => result.current.handleSelectSource({ zone: 'helper', col: 1 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource toggles the braid and the waste', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => result.current.handleSelectSource({ zone: 'braid' }));
    expect(result.current.selectedSource).toEqual({ zone: 'braid' });

    act(() => result.current.handleSelectSource({ zone: 'braid' }));
    expect(result.current.selectedSource).toBeNull();

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    expect(result.current.selectedSource).toEqual({ zone: 'waste' });
  });

  it('handleSelectTarget no-ops without a selected source', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 1 }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('handleSelectTarget sends a braid field to a foundation through the move API', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'field', col: 2 }));
    act(() => result.current.handleSelectTarget({ zone: 'foundation', col: 0 }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'field', col: 2 }, { zone: 'foundation', col: 0 }),
    );
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget parks the waste in a helper', async () => {
    const { result } = renderHook(() => useBraidGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleSelectSource({ zone: 'waste' }));
    act(() => result.current.handleSelectTarget({ zone: 'helper', col: 3 }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'helper', col: 3 }));
  });
});
