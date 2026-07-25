import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { SoundProvider, useSound } from '../providers/SoundProvider';
import { useGameApi } from './useGameApi';

// Track plays per sound file so central-tap tests can assert WHICH sound
// fired (the shared setup.ts mock can't distinguish Howl instances).
const { playCalls } = vi.hoisted(() => ({ playCalls: [] as string[] }));
vi.mock('howler', () => ({
  Howl: class MockHowl {
    private src: string;
    constructor(opts: { src: string[] }) {
      this.src = opts.src[0];
    }
    play() {
      playCalls.push(this.src);
      return 1;
    }
    volume() {}
    rate() {}
  },
  Howler: { ctx: { state: 'running' } },
}));

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

/** QueryClient + SoundProvider wrapper for central-tap tests. */
function createSoundWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, createElement(SoundProvider, null, children));
}

describe('useGameApi', () => {
  it('has correct initial state', () => {
    const apiFn = vi.fn();
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });
    expect(result.current.state).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('sets state on success', async () => {
    const apiFn = vi.fn().mockResolvedValue({ value: 42 });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec('arg1', 'arg2');
    });

    expect(apiFn).toHaveBeenCalledWith('arg1', 'arg2');
    expect(result.current.state).toEqual({ value: 42 });
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('sets loading during exec', async () => {
    let resolve!: (v: unknown) => void;
    const apiFn = vi.fn().mockReturnValue(
      new Promise((r) => {
        resolve = r;
      }),
    );
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    act(() => {
      result.current.exec();
    });

    await waitFor(() => expect(result.current.loading).toBe(true));

    await act(async () => {
      resolve({ done: true });
    });

    expect(result.current.loading).toBe(false);
  });

  it('sets error on failure', async () => {
    const apiFn = vi.fn().mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });

    expect(result.current.error).toBe(NETWORK_ERROR_MESSAGE());
    expect(result.current.state).toBeNull();
    expect(result.current.loading).toBe(false);
  });

  it('clears error on successful retry', async () => {
    const apiFn = vi.fn().mockRejectedValueOnce(new Error('fail')).mockResolvedValueOnce({ ok: true });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.error).not.toBeNull();

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.error).toBeNull();
    expect(result.current.state).toEqual({ ok: true });
  });

  it('calls onSuccess with response', async () => {
    const apiFn = vi.fn().mockResolvedValue({ data: 'hello' });
    const onSuccess = vi.fn();
    const { result } = renderHook(() => useGameApi(apiFn, { onSuccess }), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });

    expect(onSuccess).toHaveBeenCalledWith({ data: 'hello' });
  });

  it('works without onSuccess option', async () => {
    const apiFn = vi.fn().mockResolvedValue({ v: 1 });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });

    expect(result.current.state).toEqual({ v: 1 });
  });

  it('does not call onSuccess on error', async () => {
    const apiFn = vi.fn().mockRejectedValue(new Error('err'));
    const onSuccess = vi.fn();
    const { result } = renderHook(() => useGameApi(apiFn, { onSuccess }), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });

    expect(onSuccess).not.toHaveBeenCalled();
  });

  it('allows direct setState usage', async () => {
    const apiFn = vi.fn().mockResolvedValue({ init: true });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.state).toEqual({ init: true });

    act(() => {
      result.current.setState({ manual: true });
    });
    expect(result.current.state).toEqual({ manual: true });
  });

  it('passes all arguments through to the api function', async () => {
    const apiFn = vi.fn().mockResolvedValue(null);
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec('cmd', [1, 2], 99);
    });

    expect(apiFn).toHaveBeenCalledWith('cmd', [1, 2], 99);
  });

  it('preserves state on error', async () => {
    const apiFn = vi.fn().mockResolvedValueOnce({ initial: true }).mockRejectedValueOnce(new Error('fail'));
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.state).toEqual({ initial: true });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.state).toEqual({ initial: true });
    expect(result.current.error).not.toBeNull();
  });

  it('uses latest apiFn via ref', async () => {
    const apiFn1 = vi.fn().mockResolvedValue({ v: 1 });
    const apiFn2 = vi.fn().mockResolvedValue({ v: 2 });
    const wrapper = createWrapper();
    const { result, rerender } = renderHook(({ fn }) => useGameApi(fn), {
      initialProps: { fn: apiFn1 },
      wrapper,
    });

    await act(async () => {
      await result.current.exec();
    });
    expect(result.current.state).toEqual({ v: 1 });

    rerender({ fn: apiFn2 });

    await act(async () => {
      await result.current.exec();
    });
    expect(apiFn2).toHaveBeenCalled();
    expect(result.current.state).toEqual({ v: 2 });
  });

  it('keeps loading true until async onSuccess resolves', async () => {
    let resolveCallback!: () => void;
    const apiFn = vi.fn().mockResolvedValue({ v: 1 });
    const onSuccess = vi.fn().mockReturnValue(
      new Promise<void>((r) => {
        resolveCallback = r;
      }),
    );
    const { result } = renderHook(() => useGameApi(apiFn, { onSuccess }), { wrapper: createWrapper() });

    act(() => {
      result.current.exec();
    });

    await waitFor(() => expect(onSuccess).toHaveBeenCalled());
    expect(result.current.loading).toBe(true);

    await act(async () => {
      resolveCallback();
    });

    expect(result.current.loading).toBe(false);
  });

  it('retries the last command on retry()', async () => {
    const apiFn = vi.fn().mockRejectedValueOnce(new Error('fail')).mockResolvedValueOnce({ ok: true });
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.exec('cmd1', 'arg2');
    });
    expect(result.current.error).not.toBeNull();

    await act(async () => {
      await result.current.retry();
    });
    expect(apiFn).toHaveBeenLastCalledWith('cmd1', 'arg2');
    expect(result.current.error).toBeNull();
    expect(result.current.state).toEqual({ ok: true });
  });

  it('retry is a no-op when no previous call', async () => {
    const apiFn = vi.fn();
    const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.retry();
    });
    expect(apiFn).not.toHaveBeenCalled();
  });

  describe('central sound tap', () => {
    beforeEach(() => {
      playCalls.length = 0;
    });

    it("plays shuffle when the command is 'reset'", async () => {
      const apiFn = vi.fn().mockResolvedValue({ ok: true });
      const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createSoundWrapper() });
      await act(async () => {
        await result.current.exec('reset');
      });
      expect(playCalls).toContain('/sounds/shuffle.ogg');
      expect(playCalls).not.toContain('/sounds/card-place.ogg');
    });

    it('plays cardPlace for non-reset commands', async () => {
      const apiFn = vi.fn().mockResolvedValue({ ok: true });
      const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createSoundWrapper() });
      await act(async () => {
        await result.current.exec('hit', [0]);
      });
      expect(playCalls).toContain('/sounds/card-place.ogg');
      expect(playCalls).not.toContain('/sounds/shuffle.ogg');
    });

    it('stays silent when the response reports a rejected action (illegal move)', async () => {
      // Rule rejections come back 200 with `message` set and NO `messageCode`
      // (every WebPresenter's lastErr branch). No card moved, so no card sound.
      const apiFn = vi.fn().mockResolvedValue({ message: 'そのカードは置けません' });
      const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createSoundWrapper() });
      await act(async () => {
        await result.current.exec('move', 0, 1);
      });
      expect(playCalls).toEqual([]);
    });

    it('still plays when a message carries a messageCode (real state change)', async () => {
      const apiFn = vi.fn().mockResolvedValue({ message: 'ゲームクリア！', messageCode: 'bakersdozen.gameClear' });
      const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createSoundWrapper() });
      await act(async () => {
        await result.current.exec('move', 0, 1);
      });
      expect(playCalls).toContain('/sounds/card-place.ogg');
    });

    it('plays no sound on exec failure (errorBuzz belongs to ErrorAlert)', async () => {
      const apiFn = vi.fn().mockRejectedValue(new Error('network'));
      const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createSoundWrapper() });
      await act(async () => {
        await result.current.exec('hit');
      });
      expect(playCalls).toEqual([]);
    });

    it('skips the generic sound when an exec claim is pending, and consumes it', async () => {
      const apiFn = vi.fn().mockResolvedValue({ ok: true });
      const { result } = renderHook(() => ({ api: useGameApi(apiFn), sound: useSound() }), {
        wrapper: createSoundWrapper(),
      });
      act(() => result.current.sound.claimExecSound());
      await act(async () => {
        await result.current.api.exec('bet', 10);
      });
      expect(playCalls).toEqual([]);

      await act(async () => {
        await result.current.api.exec('deal');
      });
      expect(playCalls).toContain('/sounds/card-place.ogg');
    });

    it('consumes a pending claim on exec failure too', async () => {
      const apiFn = vi.fn().mockRejectedValueOnce(new Error('fail')).mockResolvedValueOnce({ ok: true });
      const { result } = renderHook(() => ({ api: useGameApi(apiFn), sound: useSound() }), {
        wrapper: createSoundWrapper(),
      });
      act(() => result.current.sound.claimExecSound());
      await act(async () => {
        await result.current.api.exec('bet');
      });
      await act(async () => {
        await result.current.api.exec('deal');
      });
      // The failed bet consumed the claim, so the deal's sound plays.
      expect(playCalls).toContain('/sounds/card-place.ogg');
    });

    it('renders without SoundProvider (QueryClient only) — no crash, no sounds', async () => {
      const apiFn = vi.fn().mockResolvedValue({ ok: true });
      const { result } = renderHook(() => useGameApi(apiFn), { wrapper: createWrapper() });
      await act(async () => {
        await result.current.exec('reset');
      });
      expect(result.current.state).toEqual({ ok: true });
      expect(playCalls).toEqual([]);
    });

    it('CRITICAL REGRESSION: exec identity survives a mute toggle', async () => {
      const apiFn = vi.fn().mockResolvedValue({ ok: true });
      const { result } = renderHook(() => ({ api: useGameApi(apiFn), sound: useSound() }), {
        wrapper: createSoundWrapper(),
      });
      const execBefore = result.current.api.exec;
      act(() => result.current.sound.toggleMute());
      expect(result.current.api.exec).toBe(execBefore);
    });
  });

  it('uses latest onSuccess via ref', async () => {
    const apiFn = vi.fn().mockResolvedValue({ v: 1 });
    const onSuccess1 = vi.fn();
    const onSuccess2 = vi.fn();
    const wrapper = createWrapper();
    const { result, rerender } = renderHook(({ cb }) => useGameApi(apiFn, { onSuccess: cb }), {
      initialProps: { cb: onSuccess1 },
      wrapper,
    });

    await act(async () => {
      await result.current.exec();
    });
    expect(onSuccess1).toHaveBeenCalled();

    rerender({ cb: onSuccess2 });

    await act(async () => {
      await result.current.exec();
    });
    expect(onSuccess2).toHaveBeenCalled();
  });
});
