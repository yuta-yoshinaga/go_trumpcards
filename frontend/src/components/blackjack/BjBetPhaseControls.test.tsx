import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { BjBetPhaseControls, type BjBetPhaseControlsProps } from './BjBetPhaseControls';

function defaultProps(overrides?: Partial<BjBetPhaseControlsProps>): BjBetPhaseControlsProps {
  return {
    betAmount: 10,
    onBetAmountChange: vi.fn(),
    deckCount: 1,
    onDeckCountChange: vi.fn(),
    cpuPlayerCount: 0,
    onCpuPlayerCountChange: vi.fn(),
    hintEnabled: false,
    onToggleHint: vi.fn(),
    dealerHitsSoft17: false,
    onToggleSoft17: vi.fn(),
    countingEnabled: false,
    onToggleCounting: vi.fn(),
    doubleAfterSplit: true,
    onToggleDAS: vi.fn(),
    loading: false,
    onBet: vi.fn(),
    perfectPairsBet: 0,
    onPerfectPairsBetChange: vi.fn(),
    twentyOnePlus3Bet: 0,
    onTwentyOnePlus3BetChange: vi.fn(),
    ...overrides,
  };
}

describe('BjBetPhaseControls', () => {
  it('renders bet amount input with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ betAmount: 50 })} />);
    expect(screen.getByLabelText('ベット額:')).toHaveValue(50);
  });

  it('calls onBetAmountChange when bet input changes', () => {
    const onBetAmountChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onBetAmountChange })} />);
    fireEvent.change(screen.getByLabelText('ベット額:'), { target: { value: '100' } });
    expect(onBetAmountChange).toHaveBeenCalledWith(100);
  });

  it('renders deck count selector with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ deckCount: 6 })} />);
    expect(screen.getByLabelText('デッキ数:')).toHaveValue('6');
  });

  it('calls onDeckCountChange when deck selector changes', () => {
    const onDeckCountChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onDeckCountChange })} />);
    fireEvent.change(screen.getByLabelText('デッキ数:'), { target: { value: '4' } });
    expect(onDeckCountChange).toHaveBeenCalledWith(4);
  });

  it('renders CPU count selector with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ cpuPlayerCount: 2 })} />);
    expect(screen.getByLabelText('CPU人数:')).toHaveValue('2');
  });

  it('calls onCpuPlayerCountChange when CPU selector changes', () => {
    const onCpuPlayerCountChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onCpuPlayerCountChange })} />);
    fireEvent.change(screen.getByLabelText('CPU人数:'), { target: { value: '3' } });
    expect(onCpuPlayerCountChange).toHaveBeenCalledWith(3);
  });

  it('shows hint OFF when hintEnabled is false', () => {
    render(<BjBetPhaseControls {...defaultProps({ hintEnabled: false })} />);
    expect(screen.getByRole('button', { name: 'ヒント OFF' })).toBeInTheDocument();
  });

  it('shows hint ON when hintEnabled is true', () => {
    render(<BjBetPhaseControls {...defaultProps({ hintEnabled: true })} />);
    expect(screen.getByRole('button', { name: 'ヒント ON' })).toBeInTheDocument();
  });

  it('calls onToggleHint when hint button is clicked', () => {
    const onToggleHint = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onToggleHint })} />);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント OFF' }));
    expect(onToggleHint).toHaveBeenCalled();
  });

  it('shows S17 when dealerHitsSoft17 is false', () => {
    render(<BjBetPhaseControls {...defaultProps({ dealerHitsSoft17: false })} />);
    expect(screen.getByRole('button', { name: 'S17' })).toBeInTheDocument();
  });

  it('shows H17 when dealerHitsSoft17 is true', () => {
    render(<BjBetPhaseControls {...defaultProps({ dealerHitsSoft17: true })} />);
    expect(screen.getByRole('button', { name: 'H17' })).toBeInTheDocument();
  });

  it('calls onToggleSoft17 when S17/H17 button is clicked', () => {
    const onToggleSoft17 = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onToggleSoft17 })} />);
    fireEvent.click(screen.getByRole('button', { name: 'S17' }));
    expect(onToggleSoft17).toHaveBeenCalled();
  });

  it('shows counting OFF when countingEnabled is false', () => {
    render(<BjBetPhaseControls {...defaultProps({ countingEnabled: false })} />);
    expect(screen.getByRole('button', { name: 'カウント OFF' })).toBeInTheDocument();
  });

  it('shows counting ON when countingEnabled is true', () => {
    render(<BjBetPhaseControls {...defaultProps({ countingEnabled: true })} />);
    expect(screen.getByRole('button', { name: 'カウント ON' })).toBeInTheDocument();
  });

  it('calls onToggleCounting when counting button is clicked', () => {
    const onToggleCounting = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onToggleCounting })} />);
    fireEvent.click(screen.getByRole('button', { name: 'カウント OFF' }));
    expect(onToggleCounting).toHaveBeenCalled();
  });

  it('shows DAS ON when doubleAfterSplit is true', () => {
    render(<BjBetPhaseControls {...defaultProps({ doubleAfterSplit: true })} />);
    expect(screen.getByRole('button', { name: 'DAS ON' })).toBeInTheDocument();
  });

  it('shows DAS OFF when doubleAfterSplit is false', () => {
    render(<BjBetPhaseControls {...defaultProps({ doubleAfterSplit: false })} />);
    expect(screen.getByRole('button', { name: 'DAS OFF' })).toBeInTheDocument();
  });

  it('calls onToggleDAS when DAS button is clicked', () => {
    const onToggleDAS = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onToggleDAS })} />);
    fireEvent.click(screen.getByRole('button', { name: 'DAS ON' }));
    expect(onToggleDAS).toHaveBeenCalled();
  });

  it('calls onBet when bet button is clicked', () => {
    const onBet = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onBet })} />);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    expect(onBet).toHaveBeenCalled();
  });

  it('renders PP input with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ perfectPairsBet: 20 })} />);
    expect(screen.getByLabelText('PP:')).toHaveValue(20);
  });

  it('calls onPerfectPairsBetChange when PP input changes', () => {
    const onPerfectPairsBetChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onPerfectPairsBetChange })} />);
    fireEvent.change(screen.getByLabelText('PP:'), { target: { value: '30' } });
    expect(onPerfectPairsBetChange).toHaveBeenCalledWith(30);
  });

  it('renders 21+3 input with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ twentyOnePlus3Bet: 40 })} />);
    expect(screen.getByLabelText('21+3:')).toHaveValue(40);
  });

  it('calls onTwentyOnePlus3BetChange when 21+3 input changes', () => {
    const onTwentyOnePlus3BetChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onTwentyOnePlus3BetChange })} />);
    fireEvent.change(screen.getByLabelText('21+3:'), { target: { value: '50' } });
    expect(onTwentyOnePlus3BetChange).toHaveBeenCalledWith(50);
  });

  it('disables inputs and buttons when loading is true', () => {
    render(<BjBetPhaseControls {...defaultProps({ loading: true })} />);
    expect(screen.getByLabelText('ベット額:')).toBeDisabled();
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByLabelText('デッキ数:')).toBeDisabled();
    expect(screen.getByLabelText('CPU人数:')).toBeDisabled();
    expect(screen.getByRole('button', { name: 'ヒント OFF' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'S17' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'カウント OFF' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'DAS ON' })).toBeDisabled();
    expect(screen.getByLabelText('PP:')).toBeDisabled();
    expect(screen.getByLabelText('21+3:')).toBeDisabled();
  });

  it('enables inputs and buttons when loading is false', () => {
    render(<BjBetPhaseControls {...defaultProps({ loading: false })} />);
    expect(screen.getByLabelText('ベット額:')).not.toBeDisabled();
    expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled();
    expect(screen.getByLabelText('デッキ数:')).not.toBeDisabled();
    expect(screen.getByLabelText('CPU人数:')).not.toBeDisabled();
    expect(screen.getByLabelText('PP:')).not.toBeDisabled();
    expect(screen.getByLabelText('21+3:')).not.toBeDisabled();
  });
});
