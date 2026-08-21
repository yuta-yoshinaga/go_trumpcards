import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { whiteheadApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { WhiteheadResponse } from '../types/card';
import { useWhiteheadGame } from './useWhiteheadGame';

vi.mock('../api/gameApi', () => ({
  whiteheadApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(whiteheadApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: WhiteheadResponse = {
  tableau: [
    [{ card: { design: 'SPADE', value: 1 }, faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: { design: 'HEART', value: 5 }, faceUp: true },
    ],
  ],
  stockCount: 20,
  waste: [{ design: 'CLOVER', value: 3 }],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 0,
  drawCount: 1,
  canUndo: false,
  isStalemate: false,
  score: -52,
  scoringMode: 0,
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useWhiteheadGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleReset dispatches reset command', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleReset();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleGiveUp dispatches giveup command', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleGiveUp();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleAutoComplete dispatches autocomplete command and sets isAutoCompleting', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleAutoComplete();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
    expect(result.current.isAutoCompleting).toBe(true);

    // Wait for the 3s timeout to clear isAutoCompleting
    await waitFor(() => expect(result.current.isAutoCompleting).toBe(false), { timeout: 4000 });
  });

  it('handleHint calls exec with hint and sets hint state', async () => {
    const hintResponse: WhiteheadResponse = {
      ...defaultState,
      hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    };
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });

  it('handleSelectSource sets selectedSource', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'waste' });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'waste' });
  });

  it('handleSelectSource toggles off when same source clicked', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'waste' });
    });
    expect(result.current.selectedSource).toEqual({ zone: 'waste' });

    act(() => {
      result.current.handleSelectSource({ zone: 'waste' });
    });
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectSource switches to new source', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'waste' });
    });
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 0 });
  });

  it('handleSelectTarget dispatches move and clears selection', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'waste' });
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleSelectTarget({ zone: 'tableau', col: 3 });
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'tableau', col: 3 }));
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget does nothing when no source selected', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleSelectTarget({ zone: 'tableau', col: 3 });
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDraw clears selectedSource and hint', async () => {
    const hintResponse: WhiteheadResponse = {
      ...defaultState,
      hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    };
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    // Set up source and hint
    act(() => {
      result.current.handleSelectSource({ zone: 'waste' });
    });
    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.selectedSource).not.toBeNull();
    expect(result.current.hint).not.toBeNull();

    // handleDraw should clear both
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });

    expect(result.current.selectedSource).toBeNull();
    expect(result.current.hint).toBeNull();
  });

  it('handleUndo dispatches undo command', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndo();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndoEscape(5);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, undefined, 5));
  });

  it('handleResetWithConfig dispatches reset with config', async () => {
    const { result } = renderHook(() => useWhiteheadGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...defaultState, drawCount: 3 });
    act(() => {
      result.current.handleResetWithConfig({ drawCount: 3 });
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { drawCount: 3 }));
  });
});
