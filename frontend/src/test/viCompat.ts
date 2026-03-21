import type { Mock } from 'bun:test';

/** Cast a mocked function to Bun's Mock type (replaces vi.mocked). */
export function asMocked<T extends (...args: never[]) => unknown>(
  fn: T,
): Mock<(...args: Parameters<T>) => ReturnType<T>> {
  return fn as unknown as Mock<(...args: Parameters<T>) => ReturnType<T>>;
}
