import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useSolitaireGameBase } from './useSolitaireGameBase';

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

interface FakeState {
  message: string;
  hint?: { kind: string } | null;
}
interface NestedHintRes {
  payload: { tip: { kind: string } | null };
}

const okState: FakeState = { message: 'ok' };

describe('useSolitaireGameBase', () => {
  let mockExec: ReturnType<typeof vi.fn<(...args: unknown[]) => Promise<FakeState>>>;

  beforeEach(() => {
    mockExec = vi.fn(async () => okState);
  });

  it('uses the default `.hint` selector when no selectHint option is passed', async () => {
    const hintApi = vi.fn(async () => ({ ...okState, hint: { kind: 'default-path' } }));
    const { result } = renderHook(
      () =>
        useSolitaireGameBase<FakeState, ['reset' | 'hint'], { kind: string }, FakeState>(mockExec, {
          hintApi,
        }),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(result.current.state).not.toBeNull());

    await act(async () => {
      await result.current.handleHint();
    });

    expect(hintApi).toHaveBeenCalled();
    expect(result.current.hint).toEqual({ kind: 'default-path' });
  });

  it('routes through a custom selectHint option when provided', async () => {
    const hintApi = vi.fn(async (): Promise<NestedHintRes> => ({ payload: { tip: { kind: 'custom-path' } } }));
    const selectHint = vi.fn((res: NestedHintRes) => res.payload.tip);
    const { result } = renderHook(
      () =>
        useSolitaireGameBase<FakeState, ['reset' | 'hint'], { kind: string }, NestedHintRes>(mockExec, {
          hintApi,
          selectHint,
        }),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(result.current.state).not.toBeNull());

    await act(async () => {
      await result.current.handleHint();
    });

    expect(selectHint).toHaveBeenCalledWith({ payload: { tip: { kind: 'custom-path' } } });
    expect(result.current.hint).toEqual({ kind: 'custom-path' });
  });

  it('handleHint is a no-op when no hintApi option is passed', async () => {
    const { result } = renderHook(
      () => useSolitaireGameBase<FakeState, ['reset' | 'hint'], { kind: string }, FakeState>(mockExec),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(result.current.state).not.toBeNull());

    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
    expect(result.current.hintError).toBeNull();
  });

  it('keeps the returned object identity stable across re-renders', async () => {
    const { result, rerender } = renderHook(
      () => useSolitaireGameBase<FakeState, ['reset'], { kind: string }, FakeState>(mockExec),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(result.current.state).not.toBeNull());
    const before = result.current;
    rerender();
    expect(result.current).toBe(before);
  });

  it('ignores repeated actions while the first action is in flight', async () => {
    let resolveAction: ((value: FakeState) => void) | undefined;
    mockExec = vi.fn(async (...args: unknown[]) => {
      if (args[0] === 'reset') return okState;
      return new Promise<FakeState>((resolve) => {
        resolveAction = resolve;
      });
    });
    const { result } = renderHook(
      () => useSolitaireGameBase<FakeState, ['reset' | 'deal'], { kind: string }>(mockExec),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    act(() => {
      result.current.runAction('deal');
      result.current.runAction('deal');
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
    expect(mockExec).toHaveBeenCalledTimes(2);

    resolveAction?.(okState);
  });

  it('allows the next action after the API call fails', async () => {
    mockExec = vi
      .fn<(...args: unknown[]) => Promise<FakeState>>()
      .mockResolvedValueOnce(okState)
      .mockRejectedValueOnce(new Error('network failure'))
      .mockResolvedValueOnce(okState);
    const { result } = renderHook(
      () => useSolitaireGameBase<FakeState, ['reset' | 'deal'], { kind: string }>(mockExec),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    act(() => result.current.runAction('deal'));
    await waitFor(() => expect(result.current.error).not.toBeNull());

    act(() => result.current.runAction('deal'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledTimes(3));
    expect(mockExec).toHaveBeenNthCalledWith(2, 'deal');
    expect(mockExec).toHaveBeenNthCalledWith(3, 'deal');
  });

  // Without the reset in `runAction`'s catch, the throw leaves the in-flight flag
  // set and every later action is swallowed by the guard -- so the second dispatch
  // below is what fails if that catch is removed.
  it('does not wedge the page when onClearSelection throws', async () => {
    const onClearSelection = vi.fn<() => void>(() => {
      throw new Error('selection cleanup failed');
    });
    const { result } = renderHook(
      () =>
        useSolitaireGameBase<FakeState, ['reset' | 'deal'], { kind: string }>(mockExec, {
          onClearSelection,
        }),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    expect(() => {
      act(() => result.current.runAction('deal'));
    }).toThrow('selection cleanup failed');
    expect(mockExec).not.toHaveBeenCalledWith('deal');

    onClearSelection.mockImplementation(() => undefined);
    act(() => result.current.runAction('deal'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  // `selectHint` runs AFTER the mounted check, so it is the observable that
  // distinguishes a guarded hook from an unguarded one: asserting on the returned
  // hint cannot, because React no-ops a post-unmount setState either way. #4447
  it('does not process a hint that arrives after unmount', async () => {
    let resolveHint: ((value: FakeState) => void) | undefined;
    const hintApi = vi.fn(
      () =>
        new Promise<FakeState>((resolve) => {
          resolveHint = resolve;
        }),
    );
    const selectHint = vi.fn((res: FakeState) => res.hint ?? null);
    const { result, unmount } = renderHook(
      () => useSolitaireGameBase<FakeState, unknown[], { kind: string }>(mockExec, { hintApi, selectHint }),
      { wrapper: createWrapper() },
    );
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    act(() => {
      void result.current.handleHint();
    });
    await waitFor(() => expect(hintApi).toHaveBeenCalled());

    unmount();
    resolveHint?.({ message: 'late', hint: { kind: 'too-late' } });
    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });

    expect(selectHint).not.toHaveBeenCalled();
  });
});
