import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { describe, expect, it, vi } from 'vitest';
import { BjBetPhaseControls, type BjBetPhaseControlsProps } from './BjBetPhaseControls';

function defaultProps(overrides?: Partial<BjBetPhaseControlsProps>): BjBetPhaseControlsProps {
  return {
    betAmount: 10,
    onBetAmountChange: vi.fn(),
    playerChips: 1000,
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
    countingSystem: 0,
    onCountingSystemChange: vi.fn(),
    deckPenetration: 75,
    onDeckPenetrationChange: vi.fn(),
    handCount: 1,
    onHandCountChange: vi.fn(),
    loading: false,
    onBet: vi.fn(),
    perfectPairsBet: 0,
    onPerfectPairsBetChange: vi.fn(),
    twentyOnePlus3Bet: 0,
    onTwentyOnePlus3BetChange: vi.fn(),
    surrenderRule: 0,
    onSurrenderRuleChange: vi.fn(),
    ...overrides,
  };
}

describe('BjBetPhaseControls', () => {
  it('renders bet amount input with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ betAmount: 50 })} />);
    // ChipBetInput now uses type=text + inputMode=numeric (#1615) — value is a string.
    expect(screen.getByLabelText('ベット額:')).toHaveValue('50');
  });

  it('calls onBetAmountChange when bet input changes', () => {
    const onBetAmountChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onBetAmountChange })} />);
    fireEvent.change(screen.getByLabelText('ベット額:'), { target: { value: '100' } });
    expect(onBetAmountChange).toHaveBeenCalledWith(100);
  });

  it('renders quick-bet chips and applies chip-derived amounts (clamped to chips)', () => {
    const onBetAmountChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onBetAmountChange, playerChips: 1000 })} />);
    const row = screen.getByTestId('bj-quick-bet');
    expect(row).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '最小' }));
    expect(onBetAmountChange).toHaveBeenLastCalledWith(10);
    fireEvent.click(screen.getByRole('button', { name: '半額' }));
    expect(onBetAmountChange).toHaveBeenLastCalledWith(500);
    fireEvent.click(screen.getByRole('button', { name: '全額' }));
    expect(onBetAmountChange).toHaveBeenLastCalledWith(1000);
  });

  it('renders additive quick-chip buttons and a clear button', () => {
    render(<BjBetPhaseControls {...defaultProps()} />);
    const row = screen.getByTestId('bj-chip-add');
    expect(row).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '+10' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '+25' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '+100' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'クリア' })).toBeInTheDocument();
  });

  it('adds the chip value to the current bet (clamped to chips) when a chip button is clicked', () => {
    const onBetAmountChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onBetAmountChange, betAmount: 10, playerChips: 1000 })} />);
    fireEvent.click(screen.getByRole('button', { name: '+25' }));
    expect(onBetAmountChange).toHaveBeenLastCalledWith(35);
    fireEvent.click(screen.getByRole('button', { name: '+100' }));
    expect(onBetAmountChange).toHaveBeenLastCalledWith(110);
  });

  it('disables a chip button when adding it would exceed the chip balance', () => {
    render(<BjBetPhaseControls {...defaultProps({ betAmount: 90, playerChips: 100 })} />);
    expect(screen.getByRole('button', { name: '+10' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: '+25' })).toBeDisabled();
    expect(screen.getByRole('button', { name: '+100' })).toBeDisabled();
  });

  it('resets the bet to the table minimum when the clear button is clicked', () => {
    const onBetAmountChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onBetAmountChange, betAmount: 500 })} />);
    fireEvent.click(screen.getByRole('button', { name: 'クリア' }));
    expect(onBetAmountChange).toHaveBeenLastCalledWith(10);
  });

  it('disables the additive chip and clear buttons when loading', () => {
    render(<BjBetPhaseControls {...defaultProps({ loading: true })} />);
    expect(screen.getByRole('button', { name: '+10' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'クリア' })).toBeDisabled();
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
    expect(screen.getByLabelText('PP (ペアベット):')).toHaveValue('20');
  });

  it('calls onPerfectPairsBetChange when PP input changes', () => {
    const onPerfectPairsBetChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onPerfectPairsBetChange })} />);
    fireEvent.change(screen.getByLabelText('PP (ペアベット):'), { target: { value: '30' } });
    expect(onPerfectPairsBetChange).toHaveBeenCalledWith(30);
  });

  it('renders 21+3 input with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ twentyOnePlus3Bet: 40 })} />);
    expect(screen.getByLabelText('21+3:')).toHaveValue('40');
  });

  it('calls onTwentyOnePlus3BetChange when 21+3 input changes', () => {
    const onTwentyOnePlus3BetChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onTwentyOnePlus3BetChange })} />);
    fireEvent.change(screen.getByLabelText('21+3:'), { target: { value: '50' } });
    expect(onTwentyOnePlus3BetChange).toHaveBeenCalledWith(50);
  });

  it('renders counting system selector with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ countingSystem: 1 })} />);
    expect(screen.getByLabelText('カウンティング方式')).toHaveValue('1');
  });

  it('calls onCountingSystemChange when counting system selector changes', () => {
    const onCountingSystemChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onCountingSystemChange })} />);
    fireEvent.change(screen.getByLabelText('カウンティング方式'), { target: { value: '2' } });
    expect(onCountingSystemChange).toHaveBeenCalledWith(2);
  });

  it('disables counting system selector when counting is off', () => {
    render(<BjBetPhaseControls {...defaultProps({ countingEnabled: false })} />);
    expect(screen.getByLabelText('カウンティング方式')).toBeDisabled();
  });

  it('enables counting system selector when counting is on', () => {
    render(<BjBetPhaseControls {...defaultProps({ countingEnabled: true })} />);
    expect(screen.getByLabelText('カウンティング方式')).not.toBeDisabled();
  });

  it('disables counting system selector when loading', () => {
    render(<BjBetPhaseControls {...defaultProps({ countingEnabled: true, loading: true })} />);
    expect(screen.getByLabelText('カウンティング方式')).toBeDisabled();
  });

  it('renders penetration selector with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ deckPenetration: 50 })} />);
    expect(screen.getByLabelText('ペネトレーション:')).toHaveValue('50');
  });

  it('calls onDeckPenetrationChange when penetration selector changes', () => {
    const onDeckPenetrationChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onDeckPenetrationChange })} />);
    fireEvent.change(screen.getByLabelText('ペネトレーション:'), { target: { value: '50' } });
    expect(onDeckPenetrationChange).toHaveBeenCalledWith(50);
  });

  it('disables penetration selector when loading', () => {
    render(<BjBetPhaseControls {...defaultProps({ loading: true })} />);
    expect(screen.getByLabelText('ペネトレーション:')).toBeDisabled();
  });

  it('enables penetration selector when not loading', () => {
    render(<BjBetPhaseControls {...defaultProps({ loading: false })} />);
    expect(screen.getByLabelText('ペネトレーション:')).not.toBeDisabled();
  });

  it('renders hand count selector with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ handCount: 2 })} />);
    expect(screen.getByLabelText('ハンド数:')).toHaveValue('2');
  });

  it('calls onHandCountChange when hand count selector changes', () => {
    const onHandCountChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onHandCountChange })} />);
    fireEvent.change(screen.getByLabelText('ハンド数:'), { target: { value: '3' } });
    expect(onHandCountChange).toHaveBeenCalledWith(3);
  });

  it('disables hand count selector when loading', () => {
    render(<BjBetPhaseControls {...defaultProps({ loading: true })} />);
    expect(screen.getByLabelText('ハンド数:')).toBeDisabled();
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
    expect(screen.getByLabelText('カウンティング方式')).toBeDisabled();
    expect(screen.getByRole('button', { name: 'DAS ON' })).toBeDisabled();
    expect(screen.getByLabelText('PP (ペアベット):')).toBeDisabled();
    expect(screen.getByLabelText('21+3:')).toBeDisabled();
    expect(screen.getByLabelText('ハンド数:')).toBeDisabled();
    expect(screen.getByLabelText('サレンダー:')).toBeDisabled();
  });

  it('enables inputs and buttons when loading is false', () => {
    render(<BjBetPhaseControls {...defaultProps({ loading: false })} />);
    expect(screen.getByLabelText('ベット額:')).not.toBeDisabled();
    expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled();
    expect(screen.getByLabelText('デッキ数:')).not.toBeDisabled();
    expect(screen.getByLabelText('CPU人数:')).not.toBeDisabled();
    expect(screen.getByLabelText('PP (ペアベット):')).not.toBeDisabled();
    expect(screen.getByLabelText('21+3:')).not.toBeDisabled();
    expect(screen.getByLabelText('ハンド数:')).not.toBeDisabled();
    expect(screen.getByLabelText('サレンダー:')).not.toBeDisabled();
  });

  it('renders surrender rule selector with provided value', () => {
    render(<BjBetPhaseControls {...defaultProps({ surrenderRule: 1 })} />);
    expect(screen.getByLabelText('サレンダー:')).toHaveValue('1');
  });

  it('calls onSurrenderRuleChange when value changes', () => {
    const onSurrenderRuleChange = vi.fn();
    render(<BjBetPhaseControls {...defaultProps({ onSurrenderRuleChange })} />);
    fireEvent.change(screen.getByLabelText('サレンダー:'), { target: { value: '2' } });
    expect(onSurrenderRuleChange).toHaveBeenCalledWith(2);
  });

  it('shows side bet badge when perfectPairsBet > 0', () => {
    render(<BjBetPhaseControls {...defaultProps({ perfectPairsBet: 100 })} />);
    expect(screen.getByText('サイドベットあり')).toBeInTheDocument();
  });

  it('shows side bet badge when twentyOnePlus3Bet > 0', () => {
    render(<BjBetPhaseControls {...defaultProps({ twentyOnePlus3Bet: 50 })} />);
    expect(screen.getByText('サイドベットあり')).toBeInTheDocument();
  });

  it('does not show side bet badge when both bets are 0', () => {
    render(<BjBetPhaseControls {...defaultProps({ perfectPairsBet: 0, twentyOnePlus3Bet: 0 })} />);
    expect(screen.queryByText('サイドベットあり')).not.toBeInTheDocument();
  });

  it('sets details open attribute when autoExpandAdvanced is true', () => {
    const { container } = render(<BjBetPhaseControls {...defaultProps({ autoExpandAdvanced: true })} />);
    const details = container.querySelector('details');
    expect(details).toHaveAttribute('open');
  });

  it('does not set details open attribute when autoExpandAdvanced is false', () => {
    const { container } = render(<BjBetPhaseControls {...defaultProps({ autoExpandAdvanced: false })} />);
    const details = container.querySelector('details');
    expect(details).not.toHaveAttribute('open');
  });

  it('renders Perfect Pairs help text inline so touch users can see it', () => {
    render(<BjBetPhaseControls {...defaultProps()} />);
    expect(screen.getByText('Perfect Pairs: 最初の2枚がペアなら配当')).toBeInTheDocument();
  });

  it('renders 21+3 help text inline so touch users can see it', () => {
    render(<BjBetPhaseControls {...defaultProps()} />);
    expect(screen.getByText('21+3: 最初の2枚+ディーラー1枚でポーカー役が成立すれば配当')).toBeInTheDocument();
  });

  it('renders soft 17 help text inline so touch users can see it', () => {
    render(<BjBetPhaseControls {...defaultProps()} />);
    expect(screen.getByText('H17=ソフト17でヒット / S17=ソフト17でスタンド')).toBeInTheDocument();
  });

  it('renders DAS help text inline so touch users can see it', () => {
    render(<BjBetPhaseControls {...defaultProps()} />);
    expect(screen.getByText('DAS=スプリット後にダブルダウン可')).toBeInTheDocument();
  });

  it('does not rely on title attribute tooltips for side bets and rule toggles', () => {
    render(<BjBetPhaseControls {...defaultProps()} />);
    expect(screen.getByLabelText('PP (ペアベット):')).not.toHaveAttribute('title');
    expect(screen.getByLabelText('21+3:')).not.toHaveAttribute('title');
    expect(screen.getByRole('button', { name: 'S17' })).not.toHaveAttribute('title');
    expect(screen.getByRole('button', { name: 'DAS ON' })).not.toHaveAttribute('title');
  });

  it('wires aria-describedby on side bets and rule toggles to their help text ids', () => {
    render(<BjBetPhaseControls {...defaultProps()} />);
    expect(screen.getByLabelText('PP (ペアベット):')).toHaveAttribute('aria-describedby', 'bj-pp-help');
    expect(screen.getByLabelText('21+3:')).toHaveAttribute('aria-describedby', 'bj-t3-help');
    expect(screen.getByRole('button', { name: 'S17' })).toHaveAttribute('aria-describedby', 'bj-soft17-help');
    expect(screen.getByRole('button', { name: 'DAS ON' })).toHaveAttribute('aria-describedby', 'bj-das-help');
  });

  it('renders the four help text paragraphs with the matching ids so aria-describedby resolves', () => {
    const { container } = render(<BjBetPhaseControls {...defaultProps()} />);
    expect(container.querySelector('#bj-pp-help')).toHaveTextContent('Perfect Pairs: 最初の2枚がペアなら配当');
    expect(container.querySelector('#bj-t3-help')).toHaveTextContent(
      '21+3: 最初の2枚+ディーラー1枚でポーカー役が成立すれば配当',
    );
    expect(container.querySelector('#bj-soft17-help')).toHaveTextContent(
      'H17=ソフト17でヒット / S17=ソフト17でスタンド',
    );
    expect(container.querySelector('#bj-das-help')).toHaveTextContent('DAS=スプリット後にダブルダウン可');
  });
});

describe('blackjack.deckUnit i18n pluralization (#1901)', () => {
  // Regression: ISSUE-006 — EN renders "Deck: 1deck(s)" because the source
  // string was a literal "deck(s)" template instead of a real plural form,
  // and there was no space between the count and the unit. Found by /qa on
  // 2026-05-20 (mobile English BlackJack header).
  it('en renders "1 deck" (singular) and "4 decks" (plural) with a leading space', () => {
    const t = i18n.getFixedT('en', 'blackjack');
    expect(t('deckUnit', { count: 1 })).toBe('1 deck');
    expect(t('deckUnit', { count: 2 })).toBe('2 decks');
    expect(t('deckUnit', { count: 4 })).toBe('4 decks');
    expect(t('deckUnit', { count: 1 })).not.toContain('(s)');
  });

  it('ja renders "Nデッキ" with no plural variation', () => {
    const t = i18n.getFixedT('ja', 'blackjack');
    expect(t('deckUnit', { count: 1 })).toBe('1デッキ');
    expect(t('deckUnit', { count: 4 })).toBe('4デッキ');
  });
});
