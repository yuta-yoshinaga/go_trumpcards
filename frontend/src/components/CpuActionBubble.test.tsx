import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CpuActionBubble } from './CpuActionBubble';

describe('CpuActionBubble', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders no visible bubble initially but keeps an empty sr-only live region', () => {
    render(<CpuActionBubble message={undefined} triggerKey={undefined} />);
    expect(screen.queryByTestId('cpu-action-bubble')).not.toBeInTheDocument();
    expect(screen.getByTestId('cpu-action-bubble-live')).toBeInTheDocument();
  });

  it('shows the bubble when message + triggerKey are set', () => {
    render(<CpuActionBubble message="CPU 1 asks for 5" triggerKey="turn-1" />);
    const bubble = screen.getByTestId('cpu-action-bubble');
    expect(bubble).toHaveTextContent('CPU 1 asks for 5');
    expect(bubble).toHaveAttribute('role', 'status');
    expect(bubble).toHaveAttribute('aria-live', 'polite');
  });

  it('auto-dismisses after durationMs', () => {
    render(<CpuActionBubble message="hello" triggerKey="k1" durationMs={1000} />);
    expect(screen.getByTestId('cpu-action-bubble')).toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(999);
    });
    expect(screen.getByTestId('cpu-action-bubble')).toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(2);
    });
    expect(screen.queryByTestId('cpu-action-bubble')).not.toBeInTheDocument();
  });

  it('re-shows and resets the timer when triggerKey changes even if message is identical', () => {
    const { rerender } = render(<CpuActionBubble message="ping" triggerKey="k1" durationMs={500} />);
    act(() => {
      vi.advanceTimersByTime(400);
    });
    rerender(<CpuActionBubble message="ping" triggerKey="k2" durationMs={500} />);
    // Still visible right after re-trigger.
    expect(screen.getByTestId('cpu-action-bubble')).toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(400);
    });
    // Would have been dismissed without re-trigger (400+400 > 500), but timer was reset.
    expect(screen.getByTestId('cpu-action-bubble')).toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(101);
    });
    expect(screen.queryByTestId('cpu-action-bubble')).not.toBeInTheDocument();
  });

  it('does not show the bubble when triggerKey is undefined', () => {
    render(<CpuActionBubble message="hello" triggerKey={undefined} />);
    expect(screen.queryByTestId('cpu-action-bubble')).not.toBeInTheDocument();
  });

  it('does not show the bubble when message is empty', () => {
    render(<CpuActionBubble message="" triggerKey="k" />);
    expect(screen.queryByTestId('cpu-action-bubble')).not.toBeInTheDocument();
  });
});
