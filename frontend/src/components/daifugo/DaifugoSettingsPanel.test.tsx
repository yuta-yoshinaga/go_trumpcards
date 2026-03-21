import { describe, expect, it, vi } from 'bun:test';
import { fireEvent, render, screen } from '@testing-library/react';
import type { DaifugoConfigInput } from '../../types/card';
import { DaifugoSettingsPanel } from './DaifugoSettingsPanel';

const defaultConfig: DaifugoConfigInput = {
  jokerCount: 2,
  eightCutEnabled: true,
  suitLockMode: 2,
  elevenBackEnabled: true,
  sequenceEnabled: true,
  cardExchangeEnabled: true,
  blindExchangeEnabled: false,
  fiveSkipEnabled: false,
  fiveSkipCount: 1,
  sevenPassEnabled: false,
  tenDiscardEnabled: false,
  spadeThreeEnabled: false,
  capitalFallEnabled: false,
  nineReverseEnabled: false,
  coupDetatEnabled: false,
  numberLockEnabled: false,
  sandstormEnabled: false,
  emperorEnabled: false,
  sequenceRevolutionEnabled: false,
  sequenceLockEnabled: false,
  illegalFinishEnabled: false,
  queenBomberEnabled: false,
  cpuDifficulty: 0,
};

describe('DaifugoSettingsPanel', () => {
  it('renders settings title/summary', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    expect(screen.getByText('ルール設定')).toBeInTheDocument();
  });

  it('renders joker count dropdown with current value', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const select = screen.getByLabelText('ジョーカー枚数:') as HTMLSelectElement;
    expect(select.value).toBe('2');
    expect(select.options).toHaveLength(3);
  });

  it('calls onChange when joker count changes', () => {
    const onChange = vi.fn();
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={onChange} />);
    const select = screen.getByLabelText('ジョーカー枚数:');
    fireEvent.change(select, { target: { value: '1' } });
    expect(onChange).toHaveBeenCalledWith('jokerCount', 1);
  });

  it('renders cpu difficulty dropdown', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const select = screen.getByLabelText('CPU難易度:') as HTMLSelectElement;
    expect(select.value).toBe('0');
    expect(screen.getByText('よわい')).toBeInTheDocument();
    expect(screen.getByText('ふつう')).toBeInTheDocument();
    expect(screen.getByText('つよい')).toBeInTheDocument();
  });

  it('calls onChange when cpu difficulty changes', () => {
    const onChange = vi.fn();
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={onChange} />);
    const select = screen.getByLabelText('CPU難易度:');
    fireEvent.change(select, { target: { value: '2' } });
    expect(onChange).toHaveBeenCalledWith('cpuDifficulty', 2);
  });

  it('renders suit lock mode dropdown with options', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const select = screen.getByLabelText('スート縛り:') as HTMLSelectElement;
    expect(select.value).toBe('2');
    expect(screen.getByText('なし')).toBeInTheDocument();
    expect(screen.getByText('片縛り')).toBeInTheDocument();
    expect(screen.getByText('両縛り')).toBeInTheDocument();
  });

  it('calls onChange when suit lock mode changes', () => {
    const onChange = vi.fn();
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={onChange} />);
    const select = screen.getByLabelText('スート縛り:');
    fireEvent.change(select, { target: { value: '1' } });
    expect(onChange).toHaveBeenCalledWith('suitLockMode', 1);
  });

  it('renders five skip count dropdown with values 1-5', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const select = screen.getByLabelText('5飛びスキップ数:') as HTMLSelectElement;
    expect(select.value).toBe('1');
    expect(select.options).toHaveLength(5);
    for (let i = 0; i < 5; i++) {
      expect(select.options[i].value).toBe(String(i + 1));
    }
  });

  it('calls onChange when five skip count changes', () => {
    const onChange = vi.fn();
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={onChange} />);
    const select = screen.getByLabelText('5飛びスキップ数:');
    fireEvent.change(select, { target: { value: '3' } });
    expect(onChange).toHaveBeenCalledWith('fiveSkipCount', 3);
  });

  it('renders all boolean rule checkboxes', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const expectedLabels = [
      '8切り',
      '11バック',
      '階段',
      'カード交換',
      '5飛び',
      '7渡し',
      '10捨て',
      'スペ3返し',
      '都落ち',
      '9リバース',
      'クーデター',
      '数縛り',
      '砂嵐',
      'エンペラー',
      '階段革命',
      '階段縛り',
      '反則上がり',
      '12ボンバー',
      'ブラインド交換',
    ];
    for (const label of expectedLabels) {
      expect(screen.getByLabelText(label)).toBeInTheDocument();
    }
  });

  it('calls onChange when a boolean checkbox is toggled', () => {
    const onChange = vi.fn();
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={onChange} />);
    const checkbox = screen.getByLabelText('8切り');
    fireEvent.click(checkbox);
    expect(onChange).toHaveBeenCalledWith('eightCutEnabled', false);
  });

  it('checkbox checked state reflects config (checked)', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const checkbox = screen.getByLabelText('8切り') as HTMLInputElement;
    expect(checkbox.checked).toBe(true);
  });

  it('checkbox checked state reflects config (unchecked)', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const checkbox = screen.getByLabelText('5飛び') as HTMLInputElement;
    expect(checkbox.checked).toBe(false);
  });

  it('renders queenBomber checkbox', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const checkbox = screen.getByLabelText('12ボンバー') as HTMLInputElement;
    expect(checkbox).toBeInTheDocument();
    expect(checkbox.checked).toBe(false);
  });

  it('renders numberLock checkbox', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    const checkbox = screen.getByLabelText('数縛り') as HTMLInputElement;
    expect(checkbox).toBeInTheDocument();
    expect(checkbox.checked).toBe(false);
  });

  it('disables fiveSkipCount select when fiveSkipEnabled is false', () => {
    render(<DaifugoSettingsPanel config={defaultConfig} onChange={vi.fn()} />);
    expect(screen.getByLabelText('5飛びスキップ数:')).toBeDisabled();
  });

  it('enables fiveSkipCount select when fiveSkipEnabled is true', () => {
    render(<DaifugoSettingsPanel config={{ ...defaultConfig, fiveSkipEnabled: true }} onChange={vi.fn()} />);
    expect(screen.getByLabelText('5飛びスキップ数:')).not.toBeDisabled();
  });

  it('disables sequenceLockEnabled checkbox when sequenceEnabled is false', () => {
    render(<DaifugoSettingsPanel config={{ ...defaultConfig, sequenceEnabled: false }} onChange={vi.fn()} />);
    expect(screen.getByLabelText('階段縛り')).toBeDisabled();
  });

  it('enables sequenceLockEnabled checkbox when sequenceEnabled is true', () => {
    render(<DaifugoSettingsPanel config={{ ...defaultConfig, sequenceEnabled: true }} onChange={vi.fn()} />);
    expect(screen.getByLabelText('階段縛り')).not.toBeDisabled();
  });

  it('disables blindExchangeEnabled checkbox when cardExchangeEnabled is false', () => {
    render(<DaifugoSettingsPanel config={{ ...defaultConfig, cardExchangeEnabled: false }} onChange={vi.fn()} />);
    expect(screen.getByLabelText('ブラインド交換')).toBeDisabled();
  });

  it('enables blindExchangeEnabled checkbox when cardExchangeEnabled is true', () => {
    render(<DaifugoSettingsPanel config={{ ...defaultConfig, cardExchangeEnabled: true }} onChange={vi.fn()} />);
    expect(screen.getByLabelText('ブラインド交換')).not.toBeDisabled();
  });
});
