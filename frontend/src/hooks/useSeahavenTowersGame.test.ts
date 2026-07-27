import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { seahaventowersApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { SeahavenTowersResponse } from '../types/card';
import { useSeahavenTowersGame } from './useSeahavenTowersGame';

vi.mock('../api/gameApi', () => ({
  seahaventowersApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(seahaventowersApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: SeahavenTowersResponse = {
  tableau: [[], [], [], [], [], [], [], [], [], []],
  reservedCells: [null, null],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useSeahavenTowersGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleReset clears selection and dispatches reset', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    expect(result.current.selectedSource).not.toBeNull();

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleReset();
    });

    expect(result.current.selectedSource).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleGiveUp clears selection and dispatches giveup', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    act(() => {
      result.current.handleGiveUp();
    });
    expect(result.current.selectedSource).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleHint sets hint state on response', async () => {
    const hintResponse: SeahavenTowersResponse = {
      ...defaultState,
      hint: { fromZone: 'reserved', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    };
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });

  it('handleAutoComplete sets autocompleting flag and dispatches command', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleAutoComplete();
    });
    expect(result.current.isAutoCompleting).toBe(true);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('handleUndo dispatches undo', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndo();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndoEscape(4);
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 4));
  });

  it('handleSelectSource sets selectedSource', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 0 });
  });

  it('handleSelectSource toggles off when same tableau source clicked', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource switches to a different tableau source', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 1, cardIndex: 0 });
    });
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 1, cardIndex: 0 });
  });

  it('handleSelectSource toggles off the same reserved-cell source', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'reserved', cell: 1 });
    });
    expect(result.current.selectedSource).toEqual({ zone: 'reserved', cell: 1 });

    act(() => {
      result.current.handleSelectSource({ zone: 'reserved', cell: 1 });
    });
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget dispatches move and clears selection', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleSelectTarget({ zone: 'foundation', col: 0 });
    });

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'foundation', col: 0 },
      ),
    );
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget does nothing when no source is selected', async () => {
    const { result } = renderHook(() => useSeahavenTowersGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleSelectTarget({ zone: 'foundation', col: 0 });
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
