import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { PhaseIndicator } from './PhaseIndicator';

describe('PhaseIndicator', () => {
  it('renders phase name', () => {
    render(<PhaseIndicator phaseName="プリフロップ" />);
    expect(screen.getByText('プリフロップ')).toBeInTheDocument();
  });

  it('shows your turn indicator when isHumanTurn is true', () => {
    render(<PhaseIndicator phaseName="アクション" isHumanTurn={true} />);
    expect(screen.getByText('あなたのターン')).toBeInTheDocument();
    const turnSpan = screen.getByText('あなたのターン');
    expect(turnSpan).toHaveClass(
      'text-ds-success',
      'motion-safe:animate-pulse',
      'font-bold',
      'text-base',
      'bg-ds-success/20',
      'ring-2',
      'ring-ds-success/40',
      'px-2',
      'py-0.5',
      'rounded-full',
    );
  });

  it('shows waiting indicator when isHumanTurn is false', () => {
    render(<PhaseIndicator phaseName="アクション" isHumanTurn={false} />);
    expect(screen.getByText('待機中')).toBeInTheDocument();
    const turnSpan = screen.getByText('待機中');
    expect(turnSpan).toHaveClass('text-game-text-muted');
  });

  it('does not render turn indicator when isHumanTurn is undefined', () => {
    render(<PhaseIndicator phaseName="ベットフェーズ" />);
    expect(screen.queryByText('あなたのターン')).not.toBeInTheDocument();
    expect(screen.queryByText('待機中')).not.toBeInTheDocument();
  });

  it('renders children', () => {
    render(
      <PhaseIndicator phaseName="フロップ" isHumanTurn={true}>
        <span>ポット: 100</span>
      </PhaseIndicator>,
    );
    expect(screen.getByText('ポット: 100')).toBeInTheDocument();
  });

  it('does not place aria-live on the outer container', () => {
    render(<PhaseIndicator phaseName="テスト" />);
    expect(screen.getByTestId('phase-indicator')).not.toHaveAttribute('aria-live');
  });

  it('exposes a scoped sr-only live region for phase announcements', () => {
    render(<PhaseIndicator phaseName="ベットフェーズ" isHumanTurn={true} />);
    const statusNode = screen.getByTestId('phase-announcement');
    expect(statusNode).toHaveAttribute('aria-live', 'polite');
    expect(statusNode).toHaveAttribute('aria-atomic', 'true');
    expect(statusNode).toHaveClass('sr-only');
    expect(statusNode.textContent).toContain('ベットフェーズ');
    expect(statusNode.textContent).toContain('あなたのターン');
  });
});
