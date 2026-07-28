import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spiderApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { SpiderResponse } from '../types/card';
import { useSpiderGame } from './useSpiderGame';

vi.mock('../api/gameApi', () => ({
  spiderApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(spiderApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: SpiderResponse = {
  tableau: [
    [{ card: { design: 'SPADE', value: 1 }, faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: { design: 'HEART', value: 5 }, faceUp: true },
    ],
  ],
  stockCount: 50,
  completedSuits: 0,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  score: 500,
  difficulty: 1,
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useSpiderGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDeal dispatches deal command', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDeal();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('handleReset dispatches reset command', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleReset();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleResetWithConfig dispatches reset with config', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...defaultState, difficulty: 2 });
    act(() => {
      result.current.handleResetWithConfig({ difficulty: 2 });
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { difficulty: 2 }));
  });

  it('handleGiveUp dispatches giveup command', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleGiveUp();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleHint calls exec with hint and sets hint state', async () => {
    const hintResponse: SpiderResponse = {
      ...defaultState,
      hint: { fromCol: 0, cardIndex: 0, toCol: 3 },
    };
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });

  it('handleAutoComplete dispatches autocomplete and sets isAutoCompleting', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleAutoComplete();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
    expect(result.current.isAutoCompleting).toBe(true);

    await waitFor(() => expect(result.current.isAutoCompleting).toBe(false), { timeout: 4000 });
  });

  it('handleUndo dispatches undo command', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndo();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndoEscape(3);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, undefined, 3));
  });

  it('handleSelectSource sets selectedSource', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 0 });
  });

  it('handleSelectSource toggles off when same source clicked', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });
    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 1, cardIndex: 1 });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 1, cardIndex: 1 });
  });

  it('handleSelectTarget dispatches move and clears selection', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleSelectTarget({ zone: 'tableau', col: 3 });
    });

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'tableau', col: 3 },
      ),
    );
    expect(result.current.selectedSource).toBeNull();
  });

  it('handleSelectTarget does nothing when no source selected', async () => {
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleSelectTarget({ zone: 'tableau', col: 3 });
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDeal clears selectedSource and hint', async () => {
    const hintResponse: SpiderResponse = {
      ...defaultState,
      hint: { fromCol: 0, cardIndex: 0, toCol: 3 },
    };
    const { result } = renderHook(() => useSpiderGame(), { wrapper: createWrapper() });
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

    // handleDeal should clear both
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDeal();
    });

    expect(result.current.selectedSource).toBeNull();
    expect(result.current.hint).toBeNull();
  });
});
