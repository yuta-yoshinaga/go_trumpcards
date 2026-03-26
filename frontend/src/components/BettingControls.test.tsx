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
    expect(screen.getByRole('button', { name: 'コール' }).className).toContain('bg-emerald-600');
    expect(screen.getByRole('button', { name: 'レイズ' }).className).toContain('bg-sky-500');
    expect(screen.getByRole('button', { name: 'フォールド' }).className).toContain('bg-gray-500');
    expect(screen.getByRole('button', { name: 'オールイン' }).className).toContain('bg-amber-500');
  });

  it('applies poker-themed styles to bet/check buttons', () => {
    render(<BettingControls {...makeProps()} />);
    expect(screen.getByRole('button', { name: 'ベット' }).className).toContain('bg-sky-500');
    expect(screen.getByRole('button', { name: 'チェック' }).className).toContain('bg-emerald-600');
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

  it('clamps input value to maxBetAmount', () => {
    const onBetAmountChange = vi.fn();
    render(<BettingControls {...makeProps({ maxBetAmount: 50, onBetAmountChange })} />);
    const input = screen.getByLabelText('ベット額:');
    fireEvent.change(input, { target: { value: '80' } });
    expect(onBetAmountChange).toHaveBeenCalledWith(50);
  });

  it('does not clamp when value is within maxBetAmount', () => {
    const onBetAmountChange = vi.fn();
    render(<BettingControls {...makeProps({ maxBetAmount: 100, onBetAmountChange })} />);
    const input = screen.getByLabelText('ベット額:');
    fireEvent.change(input, { target: { value: '80' } });
    expect(onBetAmountChange).toHaveBeenCalledWith(80);
  });
});
