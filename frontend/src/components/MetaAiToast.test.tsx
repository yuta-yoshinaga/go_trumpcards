import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { MetaAiToast } from './MetaAiToast';

describe('MetaAiToast', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('does not show toast on first render', () => {
    render(<MetaAiToast strategyStyle="balanced" />);
    expect(screen.queryByTestId('meta-ai-toast')).not.toBeInTheDocument();
  });

  it('shows toast when strategy changes', () => {
    const { rerender } = render(<MetaAiToast strategyStyle="balanced" />);
    rerender(<MetaAiToast strategyStyle="aggressive" />);
    expect(screen.getByTestId('meta-ai-toast')).toHaveTextContent('CPU戦略変更: 攻撃的');
  });

  it('does not show toast when strategy stays the same', () => {
    const { rerender } = render(<MetaAiToast strategyStyle="balanced" />);
    rerender(<MetaAiToast strategyStyle="balanced" />);
    expect(screen.queryByTestId('meta-ai-toast')).not.toBeInTheDocument();
  });

  it('auto-dismisses after 3000ms', () => {
    const { rerender } = render(<MetaAiToast strategyStyle="balanced" />);
    rerender(<MetaAiToast strategyStyle="defensive" />);
    expect(screen.getByTestId('meta-ai-toast')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(screen.queryByTestId('meta-ai-toast')).not.toBeInTheDocument();
  });

  it('resets timer on subsequent strategy change', () => {
    const { rerender } = render(<MetaAiToast strategyStyle="balanced" />);
    rerender(<MetaAiToast strategyStyle="defensive" />);
    expect(screen.getByTestId('meta-ai-toast')).toHaveTextContent('CPU戦略変更: 守備的');

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    rerender(<MetaAiToast strategyStyle="aggressive" />);
    expect(screen.getByTestId('meta-ai-toast')).toHaveTextContent('CPU戦略変更: 攻撃的');

    // Should still be visible after original 3000ms since timer was reset
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(screen.getByTestId('meta-ai-toast')).toBeInTheDocument();

    // Should dismiss after the full 3000ms from the second change
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.queryByTestId('meta-ai-toast')).not.toBeInTheDocument();
  });
});
