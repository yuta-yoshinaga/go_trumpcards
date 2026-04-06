import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { WinCelebration } from './WinCelebration';

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

describe('WinCelebration', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders nothing when show is false', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const { container } = render(<WinCelebration show={false} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders particles after delay when show is true', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<WinCelebration show={true} />);
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(screen.getByTestId('win-celebration')).toBeInTheDocument();
  });

  it('uses custom delayMs', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<WinCelebration show={true} delayMs={1000} />);
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
    act(() => {
      vi.advanceTimersByTime(600);
    });
    expect(screen.getByTestId('win-celebration')).toBeInTheDocument();
  });

  it('calls onCelebrate when particles start', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onCelebrate = vi.fn();
    render(<WinCelebration show={true} onCelebrate={onCelebrate} />);
    expect(onCelebrate).not.toHaveBeenCalled();
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(onCelebrate).toHaveBeenCalledTimes(1);
  });

  it('renders text banner in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<WinCelebration show={true} />);
    const celebration = screen.getByTestId('win-celebration');
    expect(celebration).toBeInTheDocument();
    expect(celebration).toHaveAttribute('role', 'status');
  });

  it('renders nothing when show is false in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const { container } = render(<WinCelebration show={false} />);
    expect(container.innerHTML).toBe('');
  });

  it('calls onCelebrate immediately in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onCelebrate = vi.fn();
    render(<WinCelebration show={true} onCelebrate={onCelebrate} />);
    expect(onCelebrate).toHaveBeenCalledTimes(1);
  });

  it('includes ARIA live region with win text', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<WinCelebration show={true} />);
    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('cleans up timer on unmount', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const { unmount } = render(<WinCelebration show={true} />);
    unmount();
    act(() => {
      vi.advanceTimersByTime(500);
    });
  });
});
