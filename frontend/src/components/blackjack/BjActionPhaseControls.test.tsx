import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { BjActionPhaseControls, type BjActionPhaseControlsProps } from './BjActionPhaseControls';
import {
  BJ_SUGGEST_DOUBLE,
  BJ_SUGGEST_DOUBLE_STAND,
  BJ_SUGGEST_HIT,
  BJ_SUGGEST_SPLIT,
  BJ_SUGGEST_STAND,
  BJ_SUGGEST_SURRENDER,
} from './bjConstants';

function defaultProps(overrides?: Partial<BjActionPhaseControlsProps>): BjActionPhaseControlsProps {
  return {
    loading: false,
    hintEnabled: false,
    suggestedAction: 0,
    showDoubleDown: false,
    showSplit: false,
    showSurrender: false,
    onHit: vi.fn(),
    onStand: vi.fn(),
    onDoubleDown: vi.fn(),
    onSplit: vi.fn(),
    onSurrender: vi.fn(),
    ...overrides,
  };
}

describe('BjActionPhaseControls', () => {
  it('renders hit and stand buttons', () => {
    render(<BjActionPhaseControls {...defaultProps()} />);
    expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();
  });

  it('calls onHit when hit button is clicked', () => {
    const onHit = vi.fn();
    render(<BjActionPhaseControls {...defaultProps({ onHit })} />);
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));
    expect(onHit).toHaveBeenCalled();
  });

  it('calls onStand when stand button is clicked', () => {
    const onStand = vi.fn();
    render(<BjActionPhaseControls {...defaultProps({ onStand })} />);
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    expect(onStand).toHaveBeenCalled();
  });

  it('shows double down button when showDoubleDown is true', () => {
    render(<BjActionPhaseControls {...defaultProps({ showDoubleDown: true })} />);
    expect(screen.getByRole('button', { name: 'ダブルダウン' })).toBeInTheDocument();
  });

  it('hides double down button when showDoubleDown is false', () => {
    render(<BjActionPhaseControls {...defaultProps({ showDoubleDown: false })} />);
    expect(screen.queryByRole('button', { name: 'ダブルダウン' })).not.toBeInTheDocument();
  });

  it('calls onDoubleDown when double down button is clicked', () => {
    const onDoubleDown = vi.fn();
    render(<BjActionPhaseControls {...defaultProps({ showDoubleDown: true, onDoubleDown })} />);
    fireEvent.click(screen.getByRole('button', { name: 'ダブルダウン' }));
    expect(onDoubleDown).toHaveBeenCalled();
  });

  it('shows split button when showSplit is true', () => {
    render(<BjActionPhaseControls {...defaultProps({ showSplit: true })} />);
    expect(screen.getByRole('button', { name: 'スプリット' })).toBeInTheDocument();
  });

  it('hides split button when showSplit is false', () => {
    render(<BjActionPhaseControls {...defaultProps({ showSplit: false })} />);
    expect(screen.queryByRole('button', { name: 'スプリット' })).not.toBeInTheDocument();
  });

  it('calls onSplit when split button is clicked', () => {
    const onSplit = vi.fn();
    render(<BjActionPhaseControls {...defaultProps({ showSplit: true, onSplit })} />);
    fireEvent.click(screen.getByRole('button', { name: 'スプリット' }));
    expect(onSplit).toHaveBeenCalled();
  });

  it('shows surrender button when showSurrender is true', () => {
    render(<BjActionPhaseControls {...defaultProps({ showSurrender: true })} />);
    expect(screen.getByRole('button', { name: 'サレンダー' })).toBeInTheDocument();
  });

  it('hides surrender button when showSurrender is false', () => {
    render(<BjActionPhaseControls {...defaultProps({ showSurrender: false })} />);
    expect(screen.queryByRole('button', { name: 'サレンダー' })).not.toBeInTheDocument();
  });

  it('calls onSurrender when surrender button is clicked', () => {
    const onSurrender = vi.fn();
    render(<BjActionPhaseControls {...defaultProps({ showSurrender: true, onSurrender })} />);
    fireEvent.click(screen.getByRole('button', { name: 'サレンダー' }));
    expect(onSurrender).toHaveBeenCalled();
  });

  it('disables all buttons when loading is true', () => {
    render(
      <BjActionPhaseControls
        {...defaultProps({ loading: true, showDoubleDown: true, showSplit: true, showSurrender: true })}
      />,
    );
    expect(screen.getByRole('button', { name: 'ヒット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'ダブルダウン' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'スプリット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'サレンダー' })).toBeDisabled();
  });

  it('enables buttons when loading is false', () => {
    render(<BjActionPhaseControls {...defaultProps({ loading: false })} />);
    expect(screen.getByRole('button', { name: 'ヒット' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: 'スタンド' })).not.toBeDisabled();
  });

  it('highlights hit button when hint suggests hit', () => {
    render(<BjActionPhaseControls {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_HIT })} />);
    expect(screen.getByRole('button', { name: 'ヒット' })).toHaveClass('ring-2');
  });

  it('highlights stand button when hint suggests stand', () => {
    render(<BjActionPhaseControls {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_STAND })} />);
    expect(screen.getByRole('button', { name: 'スタンド' })).toHaveClass('ring-2');
  });

  it('highlights double down button when hint suggests double', () => {
    render(
      <BjActionPhaseControls
        {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_DOUBLE, showDoubleDown: true })}
      />,
    );
    expect(screen.getByRole('button', { name: 'ダブルダウン' })).toHaveClass('ring-2');
  });

  it('highlights double down button when hint suggests doubleStand', () => {
    render(
      <BjActionPhaseControls
        {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_DOUBLE_STAND, showDoubleDown: true })}
      />,
    );
    expect(screen.getByRole('button', { name: 'ダブルダウン' })).toHaveClass('ring-2');
  });

  it('highlights split button when hint suggests split', () => {
    render(
      <BjActionPhaseControls
        {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_SPLIT, showSplit: true })}
      />,
    );
    expect(screen.getByRole('button', { name: 'スプリット' })).toHaveClass('ring-2');
  });

  it('highlights surrender button when hint suggests surrender', () => {
    render(
      <BjActionPhaseControls
        {...defaultProps({ hintEnabled: true, suggestedAction: BJ_SUGGEST_SURRENDER, showSurrender: true })}
      />,
    );
    expect(screen.getByRole('button', { name: 'サレンダー' })).toHaveClass('ring-2');
  });

  it('does not highlight buttons when hintEnabled is false', () => {
    render(
      <BjActionPhaseControls
        {...defaultProps({
          hintEnabled: false,
          suggestedAction: BJ_SUGGEST_HIT,
          showDoubleDown: true,
          showSplit: true,
          showSurrender: true,
        })}
      />,
    );
    expect(screen.getByRole('button', { name: 'ヒット' })).not.toHaveClass('ring-2');
    expect(screen.getByRole('button', { name: 'スタンド' })).not.toHaveClass('ring-2');
    expect(screen.getByRole('button', { name: 'ダブルダウン' })).not.toHaveClass('ring-2');
    expect(screen.getByRole('button', { name: 'スプリット' })).not.toHaveClass('ring-2');
    expect(screen.getByRole('button', { name: 'サレンダー' })).not.toHaveClass('ring-2');
  });
});
