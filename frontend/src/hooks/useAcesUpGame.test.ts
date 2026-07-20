import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { acesupApi } from '../api/gameApi';
import type { AcesUpCard, AcesUpResponse, Card } from '../types/card';
import { useAcesUpGame } from './useAcesUpGame';

/** Builds a minimal Aces Up top card; only the `removable` flag is exercised here. */
function topCard(removable: boolean): AcesUpCard {
  return { card: {} as Card, top: true, removable, movable: false };
}

vi.mock('../api/gameApi', () => ({
  acesupApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(acesupApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: AcesUpResponse = {
  columns: [[], [], [], []],
  stockCount: 44,
  discardCount: 4,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useAcesUpGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleDraw());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleReset dispatches reset command', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleReset());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleGiveUp dispatches giveup command', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleGiveUp());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleUndo dispatches undo command', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleUndo());

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('handleUndoEscape dispatches undo_n with count', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleUndoEscape(3));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, 3));
  });

  it('handleRemove dispatches remove command with col', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleRemove(2));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove', 2));
  });

  it('handleMove dispatches move command with col', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleMove(1));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', 1));
  });

  it('handleHint sets hint state', async () => {
    const hintResponse: AcesUpResponse = { ...defaultState, hint: { type: 'move', col: 2 } };
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleRemoveAll discards every removable card, chaining newly exposed ones', async () => {
    // Start: col 0 removable, col 1 not. Removing col 0 exposes a card that
    // makes col 1 removable — a chain the batch loop must sweep up sequentially.
    const start: AcesUpResponse = {
      ...defaultState,
      columns: [[topCard(true)], [topCard(false)], [], []],
    };
    const afterFirst: AcesUpResponse = {
      ...defaultState,
      columns: [[], [topCard(true)], [], []],
    };
    const afterSecond: AcesUpResponse = {
      ...defaultState,
      columns: [[], [], [], []],
    };

    mockExec.mockResolvedValue(start);
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(start));

    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(afterFirst).mockResolvedValueOnce(afterSecond);

    await act(async () => {
      await result.current.handleRemoveAll();
    });

    expect(mockExec).toHaveBeenCalledTimes(2);
    expect(mockExec).toHaveBeenNthCalledWith(1, 'remove', 0);
    expect(mockExec).toHaveBeenNthCalledWith(2, 'remove', 1);
    expect(result.current.state).toEqual(afterSecond);
    expect(result.current.isRemovingAll).toBe(false);
  });

  it('handleRemoveAll is a no-op when no card is removable', async () => {
    const start: AcesUpResponse = {
      ...defaultState,
      columns: [[topCard(false)], [], [], []],
    };
    mockExec.mockResolvedValue(start);
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(start));

    mockExec.mockClear();
    await act(async () => {
      await result.current.handleRemoveAll();
    });

    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleRemoveAll surfaces a network error through exec', async () => {
    const start: AcesUpResponse = {
      ...defaultState,
      columns: [[topCard(true)], [topCard(false)], [], []],
    };
    mockExec.mockResolvedValue(start);
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(start));

    mockExec.mockClear();
    // First direct remove throws; the fallback re-issue via exec also rejects,
    // which useGameApi catches and turns into an error message.
    mockExec.mockRejectedValue(new Error('Network error'));

    await act(async () => {
      await result.current.handleRemoveAll();
    });

    expect(mockExec).toHaveBeenCalledWith('remove', 0);
    await waitFor(() => expect(result.current.error).toBeTruthy());
    expect(result.current.isRemovingAll).toBe(false);
  });

  it('handleHint sets hintError on failure', async () => {
    const { result } = renderHook(() => useAcesUpGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockRejectedValue(new Error('Network error'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toBeTruthy();
    expect(result.current.hint).toBeNull();
  });
});
