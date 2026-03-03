import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
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
});
