import { describe, expect, it, vi } from 'bun:test';
import { fireEvent, render, screen } from '@testing-library/react';
import { BjInsurancePhaseControls, type BjInsurancePhaseControlsProps } from './BjInsurancePhaseControls';
import { BJ_SUGGEST_DECLINE_INSURANCE } from './bjConstants';

function defaultProps(overrides?: Partial<BjInsurancePhaseControlsProps>): BjInsurancePhaseControlsProps {
  return {
    loading: false,
    hintEnabled: false,
    suggestedAction: 0,
    onInsurance: vi.fn(),
    onDecline: vi.fn(),
    ...overrides,
  };
}

describe('BjInsurancePhaseControls', () => {
  it('renders insurance and decline buttons', () => {
    render(<BjInsurancePhaseControls {...defaultProps()} />);
    expect(screen.getByRole('button', { name: 'インシュランス' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '辞退' })).toBeInTheDocument();
  });

  it('calls onInsurance when insurance button is clicked', () => {
    const onInsurance = vi.fn();
    render(<BjInsurancePhaseControls {...defaultProps({ onInsurance })} />);
    fireEvent.click(screen.getByRole('button', { name: 'インシュランス' }));
    expect(onInsurance).toHaveBeenCalled();
  });

  it('calls onDecline when decline button is clicked', () => {
    const onDecline = vi.fn();
    render(<BjInsurancePhaseControls {...defaultProps({ onDecline })} />);
    fireEvent.click(screen.getByRole('button', { name: '辞退' }));
    expect(onDecline).toHaveBeenCalled();
  });

  it('disables buttons when loading is true', () => {
    render(<BjInsurancePhaseControls {...defaultProps({ loading: true })} />);
    expect(screen.getByRole('button', { name: 'インシュランス' })).toBeDisabled();
    expect(screen.getByRole('button', { name: '辞退' })).toBeDisabled();
  });

  it('enables buttons when loading is false', () => {
    render(<BjInsurancePhaseControls {...defaultProps({ loading: false })} />);
    expect(screen.getByRole('button', { name: 'インシュランス' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: '辞退' })).not.toBeDisabled();
  });

  it('highlights decline button when hint suggests decline insurance', () => {
    render(
      <BjInsurancePhaseControls
        {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_DECLINE_INSURANCE })}
      />,
    );
    expect(screen.getByRole('button', { name: '辞退' })).toHaveClass('ring-2');
  });

  it('does not highlight decline button when hintEnabled is false', () => {
    render(
      <BjInsurancePhaseControls
        {...defaultProps({ hintEnabled: false, suggestedAction: BJ_SUGGEST_DECLINE_INSURANCE })}
      />,
    );
    expect(screen.getByRole('button', { name: '辞退' })).not.toHaveClass('ring-2');
  });

  it('does not highlight decline button when suggestedAction is not decline', () => {
    render(<BjInsurancePhaseControls {...defaultProps({ hintEnabled: true, suggestedAction: 0 })} />);
    expect(screen.getByRole('button', { name: '辞退' })).not.toHaveClass('ring-2');
  });
});
