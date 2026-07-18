import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { BjEndPhaseControls, type BjEndPhaseControlsProps } from './BjEndPhaseControls';

function defaultProps(overrides?: Partial<BjEndPhaseControlsProps>): BjEndPhaseControlsProps {
  return {
    loading: false,
    onReset: vi.fn(),
    ...overrides,
  };
}

describe('BjEndPhaseControls', () => {
  it('renders next-game button', () => {
    render(<BjEndPhaseControls {...defaultProps()} />);
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
  });

  it('calls onReset when the button is clicked', () => {
    const onReset = vi.fn();
    render(<BjEndPhaseControls {...defaultProps({ onReset })} />);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(onReset).toHaveBeenCalled();
  });

  it('calls onRequestReset instead of onReset when the button is clicked', () => {
    const onReset = vi.fn();
    const onRequestReset = vi.fn();
    render(<BjEndPhaseControls {...defaultProps({ onReset, onRequestReset })} />);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(onRequestReset).toHaveBeenCalled();
    expect(onReset).not.toHaveBeenCalled();
  });

  it('disables button when loading is true', () => {
    render(<BjEndPhaseControls {...defaultProps({ loading: true })} />);
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeDisabled();
  });

  it('enables button when loading is false', () => {
    render(<BjEndPhaseControls {...defaultProps({ loading: false })} />);
    expect(screen.getByRole('button', { name: '次のゲーム' })).not.toBeDisabled();
  });

  it('has animate-pulse class on button', () => {
    render(<BjEndPhaseControls {...defaultProps()} />);
    expect(screen.getByRole('button', { name: '次のゲーム' })).toHaveClass('animate-pulse');
  });

  // --- Auto-advance countdown tests ---

  describe('auto-advance countdown', () => {
    beforeEach(() => {
      vi.useFakeTimers({ toFake: ['setInterval', 'clearInterval'] });
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('shows countdown when autoAdvanceSeconds is provided', () => {
      render(<BjEndPhaseControls {...defaultProps({ autoAdvanceSeconds: 5 })} />);
      expect(screen.getByRole('button', { name: '次のゲーム (5s)' })).toBeInTheDocument();
    });

    it('does not show countdown when autoAdvanceSeconds is undefined', () => {
      render(<BjEndPhaseControls {...defaultProps()} />);
      expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
    });

    it('does not show countdown when autoAdvanceSeconds is 0', () => {
      render(<BjEndPhaseControls {...defaultProps({ autoAdvanceSeconds: 0 })} />);
      expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
    });

    it('decrements countdown every second', async () => {
      render(<BjEndPhaseControls {...defaultProps({ autoAdvanceSeconds: 3 })} />);
      expect(screen.getByRole('button', { name: '次のゲーム (3s)' })).toBeInTheDocument();

      vi.advanceTimersByTime(1000);
      await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム (2s)' })).toBeInTheDocument());

      vi.advanceTimersByTime(1000);
      await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム (1s)' })).toBeInTheDocument());
    });

    it('calls onReset when countdown reaches 0', async () => {
      const onReset = vi.fn();
      render(<BjEndPhaseControls {...defaultProps({ onReset, autoAdvanceSeconds: 2 })} />);

      vi.advanceTimersByTime(2000);
      await waitFor(() => expect(onReset).toHaveBeenCalled());
    });

    it('auto-advance fires onReset directly, bypassing onRequestReset', async () => {
      const onReset = vi.fn();
      const onRequestReset = vi.fn();
      render(<BjEndPhaseControls {...defaultProps({ onReset, onRequestReset, autoAdvanceSeconds: 2 })} />);

      vi.advanceTimersByTime(2000);
      await waitFor(() => expect(onReset).toHaveBeenCalled());
      // The confirmation path is skipped for auto-advance.
      expect(onRequestReset).not.toHaveBeenCalled();
    });

    it('clears countdown text after reaching 0', async () => {
      const onReset = vi.fn();
      render(<BjEndPhaseControls {...defaultProps({ onReset, autoAdvanceSeconds: 1 })} />);

      vi.advanceTimersByTime(1000);
      await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    });
  });
});
