import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { BettingControls } from './BettingControls';

function makeProps(overrides: Partial<Parameters<typeof BettingControls>[0]> = {}) {
  return {
    inputId: 'bet-input',
    betAmount: 20,
    onBetAmountChange: vi.fn(),
    minRaise: 10,
    hasOutstandingBet: false,
    loading: false,
    onCall: vi.fn(),
    onRaise: vi.fn(),
    onBet: vi.fn(),
    onCheck: vi.fn(),
    onFold: vi.fn(),
    onAllIn: vi.fn(),
    ...overrides,
  };
}

describe('BettingControls', () => {
  it('renders bet/check buttons when no outstanding bet', () => {
    render(<BettingControls {...makeProps()} />);
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'コール' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'レイズ' })).not.toBeInTheDocument();
  });

  it('renders call/raise buttons when there is an outstanding bet', () => {
    render(<BettingControls {...makeProps({ hasOutstandingBet: true })} />);
    expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'チェック' })).not.toBeInTheDocument();
  });

  it('always renders fold and all-in buttons', () => {
    render(<BettingControls {...makeProps()} />);
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オールイン' })).toBeInTheDocument();
  });

  it('applies poker-themed styles to action buttons', () => {
    render(<BettingControls {...makeProps({ hasOutstandingBet: true })} />);
    expect(screen.getByRole('button', { name: 'コール' }).className).toContain('bg-poker-call');
    expect(screen.getByRole('button', { name: 'レイズ' }).className).toContain('bg-poker-raise');
    expect(screen.getByRole('button', { name: 'フォールド' }).className).toContain('bg-poker-fold');
    expect(screen.getByRole('button', { name: 'オールイン' }).className).toContain('bg-poker-allin');
  });

  it('applies poker-themed styles to bet/check buttons', () => {
    render(<BettingControls {...makeProps()} />);
    expect(screen.getByRole('button', { name: 'ベット' }).className).toContain('bg-poker-raise');
    expect(screen.getByRole('button', { name: 'チェック' }).className).toContain('bg-poker-call');
  });

  it('disables buttons when loading', () => {
    render(<BettingControls {...makeProps({ loading: true })} />);
    for (const btn of screen.getAllByRole('button')) {
      expect(btn).toBeDisabled();
    }
  });

  it('calls onBetAmountChange when input changes', () => {
    const onBetAmountChange = vi.fn();
    render(<BettingControls {...makeProps({ onBetAmountChange })} />);
    const input = screen.getByLabelText('ベット額:');
    fireEvent.change(input, { target: { value: '50' } });
    expect(onBetAmountChange).toHaveBeenCalledWith(50);
  });

  it('calls onBet when bet button clicked', () => {
    const onBet = vi.fn();
    render(<BettingControls {...makeProps({ onBet })} />);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    expect(onBet).toHaveBeenCalled();
  });

  it('calls onCheck when check button clicked', () => {
    const onCheck = vi.fn();
    render(<BettingControls {...makeProps({ onCheck })} />);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    expect(onCheck).toHaveBeenCalled();
  });

  it('calls onCall when call button clicked', () => {
    const onCall = vi.fn();
    render(<BettingControls {...makeProps({ hasOutstandingBet: true, onCall })} />);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    expect(onCall).toHaveBeenCalled();
  });

  it('calls onRaise when raise button clicked', () => {
    const onRaise = vi.fn();
    render(<BettingControls {...makeProps({ hasOutstandingBet: true, onRaise })} />);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    expect(onRaise).toHaveBeenCalled();
  });

  it('calls onFold when fold button clicked', () => {
    const onFold = vi.fn();
    render(<BettingControls {...makeProps({ onFold })} />);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    expect(onFold).toHaveBeenCalled();
  });

  it('calls onAllIn when all-in button clicked', () => {
    const onAllIn = vi.fn();
    render(<BettingControls {...makeProps({ onAllIn })} />);
    fireEvent.click(screen.getByRole('button', { name: 'オールイン' }));
    expect(onAllIn).toHaveBeenCalled();
  });

  it('renders input with correct min attribute', () => {
    render(<BettingControls {...makeProps({ minRaise: 25 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).toHaveAttribute('min', '25');
  });

  it('sets max attribute on input when maxBetAmount is positive', () => {
    render(<BettingControls {...makeProps({ maxBetAmount: 100 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).toHaveAttribute('max', '100');
  });

  it('does not set max attribute when maxBetAmount is 0', () => {
    render(<BettingControls {...makeProps({ maxBetAmount: 0 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).not.toHaveAttribute('max');
  });

  it('does not set max attribute when maxBetAmount is undefined', () => {
    render(<BettingControls {...makeProps()} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).not.toHaveAttribute('max');
  });

  it('passes through value exceeding maxBetAmount without clamping', () => {
    const onBetAmountChange = vi.fn();
    render(<BettingControls {...makeProps({ maxBetAmount: 50, onBetAmountChange })} />);
    const input = screen.getByLabelText('ベット額:');
    fireEvent.change(input, { target: { value: '80' } });
    expect(onBetAmountChange).toHaveBeenCalledWith(80);
  });

  it('shows range hint when value exceeds maxBetAmount', () => {
    render(<BettingControls {...makeProps({ betAmount: 80, maxBetAmount: 50 })} />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/10 〜 50/)).toBeInTheDocument();
  });

  it('shows range hint when value is below minRaise', () => {
    render(<BettingControls {...makeProps({ betAmount: 5, minRaise: 10, maxBetAmount: 100 })} />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/10 〜 100/)).toBeInTheDocument();
  });

  it('does not show range hint when value is within range', () => {
    render(<BettingControls {...makeProps({ betAmount: 30, minRaise: 10, maxBetAmount: 100 })} />);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('sets aria-invalid on input when out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: 80, maxBetAmount: 50 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input).toHaveAttribute('aria-invalid', 'true');
  });

  it('applies error styling to input when out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: 80, maxBetAmount: 50 })} />);
    const input = screen.getByLabelText('ベット額:');
    expect(input.className).toContain('bg-red-900/40');
  });

  it('disables bet/raise button when value is out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: 5 })} />);
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
  });

  it('disables raise button when value is out of range with outstanding bet', () => {
    render(<BettingControls {...makeProps({ betAmount: 5, hasOutstandingBet: true })} />);
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeDisabled();
  });

  it('treats NaN betAmount as out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: NaN })} />);
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByLabelText('ベット額:')).toHaveAttribute('aria-invalid', 'true');
  });

  it('does not disable check/call/fold/all-in when out of range', () => {
    render(<BettingControls {...makeProps({ betAmount: 5, hasOutstandingBet: true })} />);
    expect(screen.getByRole('button', { name: 'コール' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: 'フォールド' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: 'オールイン' })).not.toBeDisabled();
  });
});
