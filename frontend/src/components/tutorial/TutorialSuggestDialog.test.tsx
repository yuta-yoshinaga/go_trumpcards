import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TutorialSuggestDialog } from './TutorialSuggestDialog';

describe('TutorialSuggestDialog', () => {
  const defaultProps = {
    open: true,
    onStartTutorial: vi.fn(),
    onSkip: vi.fn(),
    dontShowAgain: false,
    onDontShowAgainChange: vi.fn(),
  };

  it('renders nothing when open is false', () => {
    const { container } = render(<TutorialSuggestDialog {...defaultProps} open={false} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders dialog when open', () => {
    render(<TutorialSuggestDialog {...defaultProps} />);
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('displays title and message', () => {
    render(<TutorialSuggestDialog {...defaultProps} />);
    expect(screen.getByText('チュートリアル')).toBeInTheDocument();
    expect(screen.getByText(/チュートリアルで遊び方を確認/)).toBeInTheDocument();
  });

  it('calls onStartTutorial when start button is clicked', () => {
    const onStart = vi.fn();
    render(<TutorialSuggestDialog {...defaultProps} onStartTutorial={onStart} />);
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアルを開始' }));
    expect(onStart).toHaveBeenCalledTimes(1);
  });

  it('calls onSkip when skip button is clicked', () => {
    const onSkip = vi.fn();
    render(<TutorialSuggestDialog {...defaultProps} onSkip={onSkip} />);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    expect(onSkip).toHaveBeenCalledTimes(1);
  });

  it('renders checkbox and calls onDontShowAgainChange', () => {
    const onChange = vi.fn();
    render(<TutorialSuggestDialog {...defaultProps} onDontShowAgainChange={onChange} />);
    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox);
    expect(onChange).toHaveBeenCalledWith(true);
  });

  it('checkbox reflects dontShowAgain prop', () => {
    render(<TutorialSuggestDialog {...defaultProps} dontShowAgain={true} />);
    expect(screen.getByRole('checkbox')).toBeChecked();
  });

  it('closes on Escape key', () => {
    const onSkip = vi.fn();
    render(<TutorialSuggestDialog {...defaultProps} onSkip={onSkip} />);
    fireEvent.keyDown(screen.getByRole('alertdialog'), { key: 'Escape' });
    expect(onSkip).toHaveBeenCalledTimes(1);
  });

  it('closes on backdrop click', () => {
    const onSkip = vi.fn();
    render(<TutorialSuggestDialog {...defaultProps} onSkip={onSkip} />);
    // Click the backdrop (presentation layer)
    fireEvent.click(screen.getByRole('presentation'));
    expect(onSkip).toHaveBeenCalledTimes(1);
  });

  it('does not close when clicking inside dialog', () => {
    const onSkip = vi.fn();
    render(<TutorialSuggestDialog {...defaultProps} onSkip={onSkip} />);
    fireEvent.click(screen.getByRole('alertdialog'));
    expect(onSkip).not.toHaveBeenCalled();
  });
});
