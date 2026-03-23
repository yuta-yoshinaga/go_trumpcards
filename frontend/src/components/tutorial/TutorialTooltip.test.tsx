import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TutorialTooltip } from './TutorialTooltip';

describe('TutorialTooltip', () => {
  const defaultProps = {
    message: 'ここをクリックしてください',
    placement: 'bottom' as const,
    stepIndex: 0,
    totalSteps: 5,
    onNext: vi.fn(),
    onSkip: vi.fn(),
    advanceOn: 'next' as const,
  };

  it('renders the message text', () => {
    render(<TutorialTooltip {...defaultProps} />);
    expect(screen.getByText('ここをクリックしてください')).toBeInTheDocument();
  });

  it('renders step indicator', () => {
    render(<TutorialTooltip {...defaultProps} stepIndex={2} totalSteps={5} />);
    expect(screen.getByText('3 / 5')).toBeInTheDocument();
  });

  it('renders next and skip buttons when advanceOn is next', () => {
    render(<TutorialTooltip {...defaultProps} />);
    expect(screen.getByRole('button', { name: '次へ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('hides next button when advanceOn is click', () => {
    render(<TutorialTooltip {...defaultProps} advanceOn="click" />);
    expect(screen.queryByRole('button', { name: '次へ' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('calls onNext when next button is clicked', () => {
    const onNext = vi.fn();
    render(<TutorialTooltip {...defaultProps} onNext={onNext} />);
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));
    expect(onNext).toHaveBeenCalledTimes(1);
  });

  it('calls onSkip when skip button is clicked', () => {
    const onSkip = vi.fn();
    render(<TutorialTooltip {...defaultProps} onSkip={onSkip} />);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    expect(onSkip).toHaveBeenCalledTimes(1);
  });

  it('shows complete label on last step', () => {
    render(<TutorialTooltip {...defaultProps} stepIndex={4} totalSteps={5} />);
    expect(screen.getByRole('button', { name: '完了' })).toBeInTheDocument();
  });

  it('has aria-live attribute for accessibility', () => {
    render(<TutorialTooltip {...defaultProps} />);
    const tooltip = screen.getByRole('status');
    expect(tooltip).toHaveAttribute('aria-live', 'polite');
  });
});
