import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AutoFlipCountdown } from './AutoFlipCountdown';

/** Format helper used in tests — mirrors what callers do (e.g. via i18n). */
const fmt = (n: number): string => `in ${n}s`;

describe('AutoFlipCountdown', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows ceil of the remaining seconds at mount', () => {
    render(<AutoFlipCountdown durationMs={1500} ariaLabel="aria" formatRemaining={fmt} />);
    const timer = screen.getByRole('timer', { name: 'aria' });
    expect(timer).toBeInTheDocument();
    expect(timer.querySelector('text')?.textContent).toBe('2');
    expect(screen.getByText('in 2s')).toBeInTheDocument();
  });

  it('drains to 0 once the duration has elapsed', () => {
    render(<AutoFlipCountdown durationMs={1500} ariaLabel="aria" formatRemaining={fmt} />);
    act(() => {
      vi.advanceTimersByTime(1600);
    });
    const timer = screen.getByRole('timer');
    expect(timer.querySelector('text')?.textContent).toBe('0');
  });

  it('exposes aria-label and aria-live for screen-reader announcements', () => {
    render(<AutoFlipCountdown durationMs={1500} ariaLabel="Countdown" formatRemaining={fmt} />);
    const timer = screen.getByRole('timer', { name: 'Countdown' });
    expect(timer).toHaveAttribute('aria-live', 'polite');
  });

  it('re-formats the screen-reader announcement on each tick (not pre-interpolated)', () => {
    const formatRemaining = vi.fn((n: number) => `${n} sec`);
    render(<AutoFlipCountdown durationMs={1500} ariaLabel="aria" formatRemaining={formatRemaining} />);
    // Initial: 2s remaining.
    expect(screen.getByText('2 sec')).toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    // After 1s elapsed: 1s remaining ⇒ the formatter must have been re-invoked
    // with the new value (the bug Gemini caught was passing an already-interpolated
    // string, which would freeze the announcement at "2 sec").
    expect(screen.getByText('1 sec')).toBeInTheDocument();
  });
});
