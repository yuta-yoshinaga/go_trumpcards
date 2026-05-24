import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { SlapBurst } from './SlapBurst';

describe('SlapBurst', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders nothing when triggerKey is 0', () => {
    render(<SlapBurst triggerKey={0} outcome="correct" label="PAIR!" />);
    expect(screen.queryByTestId('slap-burst')).not.toBeInTheDocument();
  });

  it('fires on first non-zero trigger and shows the label', () => {
    render(<SlapBurst triggerKey={1} outcome="correct" label="PAIR!" />);
    const burst = screen.getByTestId('slap-burst');
    expect(burst).toBeInTheDocument();
    expect(burst.getAttribute('data-outcome')).toBe('correct');
    expect(screen.getByText('PAIR!')).toBeInTheDocument();
  });

  it('auto-dismisses after the burst timeout', () => {
    render(<SlapBurst triggerKey={1} outcome="correct" label="PAIR!" />);
    expect(screen.getByTestId('slap-burst')).toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(1500);
    });
    expect(screen.queryByTestId('slap-burst')).not.toBeInTheDocument();
  });

  it('re-fires when triggerKey changes to a new non-zero value', () => {
    const { rerender } = render(<SlapBurst triggerKey={1} outcome="wrong" label="MISS!" />);
    act(() => {
      vi.advanceTimersByTime(1500);
    });
    expect(screen.queryByTestId('slap-burst')).not.toBeInTheDocument();

    rerender(<SlapBurst triggerKey={2} outcome="wrong" label="MISS!" />);
    expect(screen.getByTestId('slap-burst')).toBeInTheDocument();
  });

  it('does not refire when triggerKey stays the same on rerender', () => {
    const { rerender } = render(<SlapBurst triggerKey={1} outcome="correct" label="PAIR!" />);
    act(() => {
      vi.advanceTimersByTime(1500);
    });
    rerender(<SlapBurst triggerKey={1} outcome="correct" label="PAIR!" />);
    expect(screen.queryByTestId('slap-burst')).not.toBeInTheDocument();
  });
});
