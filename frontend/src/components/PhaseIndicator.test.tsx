import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
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
    expect(turnSpan).toHaveClass('text-green-400', 'animate-pulse');
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

  it('has aria-live polite', () => {
    render(<PhaseIndicator phaseName="テスト" />);
    expect(screen.getByTestId('phase-indicator')).toHaveAttribute('aria-live', 'polite');
  });
});
