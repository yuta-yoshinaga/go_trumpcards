import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers';

declare module 'bun:test' {
  interface Matchers<T = unknown> extends TestingLibraryMatchers<typeof expect.stringContaining, T> {}
  interface AsymmetricMatchers extends TestingLibraryMatchers<unknown, unknown> {}

  // Extend useFakeTimers to accept Vitest-compatible toFake option
  function useFakeTimers(options?: { now?: number | Date; toFake?: string[] }): void;
  namespace jest {
    function useFakeTimers(options?: { now?: number | Date; toFake?: string[] }): void;
  }
}
