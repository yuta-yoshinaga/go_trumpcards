import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { klondikeApi } from '../api/gameApi';
import type { KlondikeResponse } from '../types/card';
import { useKlondikeGame } from './useKlondikeGame';

vi.mock('../api/gameApi', () => ({
  klondikeApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(klondikeApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: KlondikeResponse = {
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
  message: '',
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useKlondikeGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleReset dispatches reset command', async () => {
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleReset();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('handleGiveUp dispatches giveup command', async () => {
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleGiveUp();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleAutoComplete dispatches autocomplete command', async () => {
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleAutoComplete();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('handleHint calls exec with hint and sets hint state', async () => {
    const hintResponse: KlondikeResponse = {
      ...defaultState,
      hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    };
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue(hintResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toEqual(hintResponse.hint);
  });

  it('handleHint sets hint to null when no hint returned', async () => {
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockResolvedValue({ ...defaultState, hint: undefined });
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('handleSelectSource sets selectedSource', async () => {
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectSource({ zone: 'waste' });
    });

    expect(result.current.selectedSource).toEqual({ zone: 'waste' });
  });

  it('handleSelectSource toggles off when same source clicked', async () => {
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleSelectTarget({ zone: 'tableau', col: 3 });
    });

    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDraw clears selectedSource and hint', async () => {
    const hintResponse: KlondikeResponse = {
      ...defaultState,
      hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    };
    const { result } = renderHook(() => useKlondikeGame(), { wrapper: createWrapper() });
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
});
