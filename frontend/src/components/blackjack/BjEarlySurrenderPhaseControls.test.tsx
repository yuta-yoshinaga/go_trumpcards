import { describe, expect, it, vi } from 'bun:test';
import { fireEvent, render, screen } from '@testing-library/react';
import {
  BjEarlySurrenderPhaseControls,
  type BjEarlySurrenderPhaseControlsProps,
} from './BjEarlySurrenderPhaseControls';
import { BJ_SUGGEST_STAND, BJ_SUGGEST_SURRENDER } from './bjConstants';

function defaultProps(overrides?: Partial<BjEarlySurrenderPhaseControlsProps>): BjEarlySurrenderPhaseControlsProps {
  return {
    loading: false,
    hintEnabled: false,
    suggestedAction: 0,
    onSurrender: vi.fn(),
    onContinue: vi.fn(),
    ...overrides,
  };
}

describe('BjEarlySurrenderPhaseControls', () => {
  it('renders surrender and continue buttons', () => {
    render(<BjEarlySurrenderPhaseControls {...defaultProps()} />);
    expect(screen.getByRole('button', { name: 'アーリーサレンダー' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '続行' })).toBeInTheDocument();
  });

  it('calls onSurrender when surrender button is clicked', () => {
    const onSurrender = vi.fn();
    render(<BjEarlySurrenderPhaseControls {...defaultProps({ onSurrender })} />);
    fireEvent.click(screen.getByRole('button', { name: 'アーリーサレンダー' }));
    expect(onSurrender).toHaveBeenCalled();
  });

  it('calls onContinue when continue button is clicked', () => {
    const onContinue = vi.fn();
    render(<BjEarlySurrenderPhaseControls {...defaultProps({ onContinue })} />);
    fireEvent.click(screen.getByRole('button', { name: '続行' }));
    expect(onContinue).toHaveBeenCalled();
  });

  it('disables buttons when loading is true', () => {
    render(<BjEarlySurrenderPhaseControls {...defaultProps({ loading: true })} />);
    expect(screen.getByRole('button', { name: 'アーリーサレンダー' })).toBeDisabled();
    expect(screen.getByRole('button', { name: '続行' })).toBeDisabled();
  });

  it('highlights surrender button when hint suggests surrender', () => {
    render(
      <BjEarlySurrenderPhaseControls {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_SURRENDER })} />,
    );
    expect(screen.getByRole('button', { name: 'アーリーサレンダー' })).toHaveClass('ring-2');
  });

  it('highlights continue button when hint suggests stand', () => {
    render(
      <BjEarlySurrenderPhaseControls {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_STAND })} />,
    );
    expect(screen.getByRole('button', { name: '続行' })).toHaveClass('ring-2');
  });

  it('does not highlight buttons when hintEnabled is false', () => {
    render(
      <BjEarlySurrenderPhaseControls
        {...defaultProps({ hintEnabled: false, suggestedAction: BJ_SUGGEST_SURRENDER })}
      />,
    );
    expect(screen.getByRole('button', { name: 'アーリーサレンダー' })).not.toHaveClass('ring-2');
    expect(screen.getByRole('button', { name: '続行' })).not.toHaveClass('ring-2');
  });
});
