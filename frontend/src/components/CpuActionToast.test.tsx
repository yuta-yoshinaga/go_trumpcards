import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CpuActionToast } from './CpuActionToast';

describe('CpuActionToast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders nothing when actions is undefined', () => {
    const { container } = render(<CpuActionToast actions={undefined} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders nothing when actions is empty', () => {
    const { container } = render(<CpuActionToast actions={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders toast when actions appear', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByText(/Player 1/)).toBeInTheDocument();
  });

  it('auto-dismisses after 3 seconds', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(screen.queryByRole('status')).toBeNull();
  });

  it('resets timer when new actions arrive', () => {
    const actions1 = [{ playerIdx: 1, action: 2, amount: 0 }];
    const { rerender } = render(<CpuActionToast actions={actions1} />);
    expect(screen.getByRole('status')).toBeInTheDocument();

    // Advance 2s, then new actions arrive
    act(() => {
      vi.advanceTimersByTime(2000);
    });

    const actions2 = [
      { playerIdx: 1, action: 2, amount: 0 },
      { playerIdx: 2, action: 3, amount: 40 },
    ];
    rerender(<CpuActionToast actions={actions2} />);

    // After another 2s (total 4s), toast should still be visible because timer reset
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(screen.getByRole('status')).toBeInTheDocument();

    // After 3s from last update, toast disappears
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('shows amount when present', () => {
    const actions = [{ playerIdx: 2, action: 3, amount: 100 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByText(/100/)).toBeInTheDocument();
  });

  it('has aria-live polite attribute', () => {
    const actions = [{ playerIdx: 1, action: 2, amount: 0 }];
    render(<CpuActionToast actions={actions} />);
    expect(screen.getByRole('status')).toHaveAttribute('aria-live', 'polite');
  });
});
