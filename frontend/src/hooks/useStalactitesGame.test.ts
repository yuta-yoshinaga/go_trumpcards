import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { stalactitesApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { StalactitesResponse } from '../types/card';
import { useStalactitesGame } from './useStalactitesGame';

vi.mock('../api/gameApi', () => ({
  stalactitesApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(stalactitesApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: StalactitesResponse = {
  tableau: [[], [], [], [], [], [], [], []],
  baseRank: 1,
  cells: [null, null, null, null],
  foundation: [[], [], [], []],
  // 空きセル4 + 空き列8 → (1+4) * 2^8 = 1280。空き列宛てはその列を除いて 640。
  maxMovableCards: 1280,
  maxMovableCardsToEmptyColumn: 640,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useStalactitesGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleReset dispatches reset command and clears selection and hint', async () => {
    const hintResponse: StalactitesResponse = {
      ...defaultState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    };
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    // Set up source and hint
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.selectedSource).not.toBeNull();
    expect(result.current.hint).not.toBeNull();

    // handleReset should clear both
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleReset();
    });

    expect(result.current.selectedSource).toBeNull();
    expect(result.current.hint).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleGiveUp dispatches giveup command and clears selection and hint', async () => {
    const hintResponse: StalactitesResponse = {
      ...defaultState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    };
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    // Set up source and hint
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.selectedSource).not.toBeNull();
    expect(result.current.hint).not.toBeNull();

    // handleGiveUp should clear both
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleGiveUp();
    });

    expect(result.current.selectedSource).toBeNull();
    expect(result.current.hint).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleHint calls exec with hint and sets hint state', async () => {
    const hintResponse: StalactitesResponse = {
      ...defaultState,
      hint: { fromZone: 'stalactites', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    };
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });

  it('handleAutoComplete dispatches autocomplete and clears selection and hint', async () => {
    const hintResponse: StalactitesResponse = {
      ...defaultState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    };
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    // Set up source and hint
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.selectedSource).not.toBeNull();
    expect(result.current.hint).not.toBeNull();

    // handleAutoComplete should clear both
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleAutoComplete();
    });

    expect(result.current.selectedSource).toBeNull();
    expect(result.current.hint).toBeNull();
    expect(result.current.isAutoCompleting).toBe(true);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('isAutoCompleting resets to false after timeout', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleAutoComplete();
    });
    expect(result.current.isAutoCompleting).toBe(true);

    await waitFor(() => expect(result.current.isAutoCompleting).toBe(false), { timeout: 4000 });
  });

  it('handleUndo dispatches undo and clears selection and hint', async () => {
    const hintResponse: StalactitesResponse = {
      ...defaultState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    };
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    // Set up source and hint
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.selectedSource).not.toBeNull();
    expect(result.current.hint).not.toBeNull();

    // handleUndo should clear both
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndo();
    });

    expect(result.current.selectedSource).toBeNull();
    expect(result.current.hint).toBeNull();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndoEscape(4);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 4));
  });

  it('handleSelectSource sets selectedSource', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 0 });
  });

  it('handleSelectSource toggles off when same source clicked', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 0 });

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource switches to new source', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 1, cardIndex: 0 });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 1, cardIndex: 0 });
  });

  it('handleSelectSource with cell field (stalactites zone)', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'stalactites', cell: 2 });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'stalactites', cell: 2 });
  });

  it('handleSelectSource toggles off stalactites source when same cell clicked', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'stalactites', cell: 2 });
    });
    expect(result.current.selectedSource).toEqual({ zone: 'stalactites', cell: 2 });

    act(() => {
      result.current.handleSelectSource({ zone: 'stalactites', cell: 2 });
    });
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget dispatches move and clears selection', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
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

  it('handleSelectTarget does nothing when no source selected', async () => {
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleSelectTarget({ zone: 'foundation', col: 0 });
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleSelectTarget clears hint', async () => {
    const hintResponse: StalactitesResponse = {
      ...defaultState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    };
    const { result } = renderHook(() => useStalactitesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    // Set hint
    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).not.toBeNull();

    // Select source then target
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleSelectTarget({ zone: 'foundation', col: 0 });
    });

    expect(result.current.hint).toBeNull();
  });
});
