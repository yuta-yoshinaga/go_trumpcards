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
  it('renders reset button', () => {
    render(<BjEndPhaseControls {...defaultProps()} />);
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('calls onReset when reset button is clicked', () => {
    const onReset = vi.fn();
    render(<BjEndPhaseControls {...defaultProps({ onReset })} />);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(onReset).toHaveBeenCalled();
  });

  it('disables reset button when loading is true', () => {
    render(<BjEndPhaseControls {...defaultProps({ loading: true })} />);
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();
  });

  it('enables reset button when loading is false', () => {
    render(<BjEndPhaseControls {...defaultProps({ loading: false })} />);
    expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled();
  });

  it('has animate-pulse class on reset button', () => {
    render(<BjEndPhaseControls {...defaultProps()} />);
    expect(screen.getByRole('button', { name: 'リセット' })).toHaveClass('animate-pulse');
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
      expect(screen.getByRole('button', { name: 'リセット (5s)' })).toBeInTheDocument();
    });

    it('does not show countdown when autoAdvanceSeconds is undefined', () => {
      render(<BjEndPhaseControls {...defaultProps()} />);
      expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
    });

    it('does not show countdown when autoAdvanceSeconds is 0', () => {
      render(<BjEndPhaseControls {...defaultProps({ autoAdvanceSeconds: 0 })} />);
      expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
    });

    it('decrements countdown every second', async () => {
      render(<BjEndPhaseControls {...defaultProps({ autoAdvanceSeconds: 3 })} />);
      expect(screen.getByRole('button', { name: 'リセット (3s)' })).toBeInTheDocument();

      vi.advanceTimersByTime(1000);
      await waitFor(() => expect(screen.getByRole('button', { name: 'リセット (2s)' })).toBeInTheDocument());

      vi.advanceTimersByTime(1000);
      await waitFor(() => expect(screen.getByRole('button', { name: 'リセット (1s)' })).toBeInTheDocument());
    });

    it('calls onReset when countdown reaches 0', async () => {
      const onReset = vi.fn();
      render(<BjEndPhaseControls {...defaultProps({ onReset, autoAdvanceSeconds: 2 })} />);

      vi.advanceTimersByTime(2000);
      await waitFor(() => expect(onReset).toHaveBeenCalled());
    });

    it('clears countdown text after reaching 0', async () => {
      const onReset = vi.fn();
      render(<BjEndPhaseControls {...defaultProps({ onReset, autoAdvanceSeconds: 1 })} />);

      vi.advanceTimersByTime(1000);
      await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    });
  });
});
