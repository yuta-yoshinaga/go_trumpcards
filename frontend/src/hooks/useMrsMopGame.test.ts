import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mrsMopApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { MrsMopResponse } from '../types/card';
import { useMrsMopGame } from './useMrsMopGame';

vi.mock('../api/gameApi', () => ({
  mrsMopApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(mrsMopApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: MrsMopResponse = {
  tableau: [
    [{ card: { design: 'SPADE', value: 1 }, faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: { design: 'HEART', value: 5 }, faceUp: true },
    ],
  ],
  stockCount: 0,
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

describe('useMrsMopGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleReset dispatches reset command', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleReset();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleResetWithConfig dispatches reset with config', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...defaultState, difficulty: 2 });
    act(() => {
      result.current.handleResetWithConfig({ difficulty: 2 });
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { difficulty: 2 }));
  });

  it('handleGiveUp dispatches giveup command', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleGiveUp();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleHint calls exec with hint and sets hint state', async () => {
    const hintResponse: MrsMopResponse = {
      ...defaultState,
      hint: { fromCol: 0, cardIndex: 0, toCol: 3 },
    };
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });

  it('handleAutoComplete dispatches autocomplete and sets isAutoCompleting', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndo();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndoEscape(3);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, undefined, 3));
  });

  it('handleSelectSource sets selectedSource', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'tableau', col: 0, cardIndex: 0 });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'tableau', col: 0, cardIndex: 0 });
  });

  it('handleSelectSource toggles off when same source clicked', async () => {
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleSelectTarget({ zone: 'tableau', col: 3 });
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDeal clears selectedSource and hint', async () => {
    const hintResponse: MrsMopResponse = {
      ...defaultState,
      hint: { fromCol: 0, cardIndex: 0, toCol: 3 },
    };
    const { result } = renderHook(() => useMrsMopGame(), { wrapper: createWrapper() });
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

    // **配る操作は無い。**選択とヒントを落とす役目は他の操作が担う。
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleUndo();
    });

    expect(result.current.selectedSource).toBeNull();
    expect(result.current.hint).toBeNull();
  });
});
