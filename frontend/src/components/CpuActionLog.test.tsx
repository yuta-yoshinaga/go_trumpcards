import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CpuActionLog } from './CpuActionLog';

describe('CpuActionLog', () => {
  it('returns null when actions is undefined', () => {
    const { container } = render(<CpuActionLog actions={undefined} />);
    expect(container.firstChild).toBeNull();
  });

  it('returns null when actions is empty', () => {
    const { container } = render(<CpuActionLog actions={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders actions with names', () => {
    const actions = [
      { playerIdx: 1, action: 2, amount: 50 },
      { playerIdx: 2, action: 0, amount: 0 },
    ];
    render(<CpuActionLog actions={actions} />);
    expect(screen.getByText('CPU行動:')).toBeInTheDocument();
    expect(screen.getByText('Player 1: コール (50)')).toBeInTheDocument();
    expect(screen.getByText('Player 2: フォールド')).toBeInTheDocument();
  });

  it('has role="log" and aria-live="polite" for screen readers', () => {
    const actions = [{ playerIdx: 1, action: 0, amount: 0 }];
    render(<CpuActionLog actions={actions} />);
    const el = screen.getByRole('log');
    expect(el).toBeInTheDocument();
    expect(el).toHaveAttribute('aria-live', 'polite');
    expect(el).toHaveAttribute('aria-label', 'CPUプレイヤーの行動ログ');
  });

  it('renders unknown action as 不明', () => {
    const actions = [{ playerIdx: 1, action: 99, amount: 0 }];
    render(<CpuActionLog actions={actions} />);
    expect(screen.getByText('Player 1: 不明')).toBeInTheDocument();
  });
});
