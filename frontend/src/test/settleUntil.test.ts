import { configure, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { settleUntil } from './settleUntil';

describe('settleUntil', () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] });
  });
  afterEach(() => {
    vi.clearAllTimers();
    vi.useRealTimers();
  });

  it('resolves when the assertion already passes', async () => {
    const spy = vi.fn();
    spy();
    await settleUntil(() => expect(spy).toHaveBeenCalled());
  });

  it('resolves a condition that only becomes true on a later turn', async () => {
    const spy = vi.fn();
    Promise.resolve().then(() => {
      spy();
    });
    await settleUntil(() => expect(spy).toHaveBeenCalled());
  });

  // Without this the helper would be indistinguishable from one that returns
  // unconditionally, and every call site would pass while broken.
  it('rethrows the last failure when the condition never becomes true', async () => {
    const spy = vi.fn();
    await expect(settleUntil(() => expect(spy).toHaveBeenCalled(), 3)).rejects.toThrow();
  });

  // The reason this helper exists, kept executable so it cannot rot into a
  // comment that is no longer true. If @testing-library ever learns to detect
  // Vitest's fake timers, this test fails and the helper can be deleted.
  it('covers a case waitFor cannot: waitFor has no retry while setInterval is faked', async () => {
    configure({ asyncUtilTimeout: 150 });
    try {
      const spy = vi.fn();
      Promise.resolve().then(() => {
        spy();
      });
      // waitFor polls through the faked setInterval, so its check never runs
      // again; only its real-clock deadline still fires.
      await expect(waitFor(() => expect(spy).toHaveBeenCalled())).rejects.toThrow();
    } finally {
      configure({ asyncUtilTimeout: 1000 });
    }
  });
});
