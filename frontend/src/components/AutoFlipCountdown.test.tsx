import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AutoFlipCountdown } from './AutoFlipCountdown';

describe('AutoFlipCountdown', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows ceil of the remaining seconds at mount', () => {
    render(<AutoFlipCountdown durationMs={1500} ariaLabel="aria" remainingLabel="in {{n}}s" />);
    const timer = screen.getByRole('timer', { name: 'aria' });
    expect(timer).toBeInTheDocument();
    expect(timer.querySelector('text')?.textContent).toBe('2');
    expect(screen.getByText('in 2s')).toBeInTheDocument();
  });

  it('drains to 0 once the duration has elapsed', () => {
    render(<AutoFlipCountdown durationMs={1500} ariaLabel="aria" remainingLabel="in {{n}}s" />);
    act(() => {
      vi.advanceTimersByTime(1600);
    });
    const timer = screen.getByRole('timer');
    expect(timer.querySelector('text')?.textContent).toBe('0');
  });

  it('exposes aria-label and aria-live for screen-reader announcements', () => {
    render(<AutoFlipCountdown durationMs={1500} ariaLabel="Countdown" remainingLabel="in {{n}}s" />);
    const timer = screen.getByRole('timer', { name: 'Countdown' });
    expect(timer).toHaveAttribute('aria-live', 'polite');
  });
});
