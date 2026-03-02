import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CpuActionLog } from './CpuActionLog';

const names: Record<number, string> = { 0: 'フォールド', 1: 'チェック', 2: 'コール', 3: 'ベット' };

describe('CpuActionLog', () => {
  it('returns null when actions is undefined', () => {
    const { container } = render(<CpuActionLog actions={undefined} actionNames={names} />);
    expect(container.firstChild).toBeNull();
  });

  it('returns null when actions is empty', () => {
    const { container } = render(<CpuActionLog actions={[]} actionNames={names} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders actions with names', () => {
    const actions = [
      { playerIdx: 1, action: 2, amount: 50 },
      { playerIdx: 2, action: 0, amount: 0 },
    ];
    render(<CpuActionLog actions={actions} actionNames={names} />);
    expect(screen.getByText('CPU行動:')).toBeInTheDocument();
    expect(screen.getByText('Player 1: コール (50)')).toBeInTheDocument();
    expect(screen.getByText('Player 2: フォールド')).toBeInTheDocument();
  });

  it('renders unknown action as 不明', () => {
    const actions = [{ playerIdx: 1, action: 99, amount: 0 }];
    render(<CpuActionLog actions={actions} actionNames={names} />);
    expect(screen.getByText('Player 1: 不明')).toBeInTheDocument();
  });
});
