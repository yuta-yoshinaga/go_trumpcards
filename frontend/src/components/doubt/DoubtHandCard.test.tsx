import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card } from '../../types/card';
import { DoubtHandCard } from './DoubtHandCard';

const card: Card = { design: 'SPADE', value: 1 };

describe('DoubtHandCard', () => {
  it('renders a button with data-testid hand-card', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={true} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card')).toBeInTheDocument();
  });

  it('sets aria-pressed true when selected', () => {
    render(<DoubtHandCard card={card} index={0} selected={true} selectable={true} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card')).toHaveAttribute('aria-pressed', 'true');
  });

  it('sets aria-pressed false when not selected', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={true} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card')).toHaveAttribute('aria-pressed', 'false');
  });

  it('is enabled when selectable', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={true} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card')).toBeEnabled();
  });

  it('is disabled when not selectable', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={false} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card')).toBeDisabled();
  });

  it('has pointer cursor when selectable', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={true} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card').style.cursor).toBe('pointer');
  });

  it('has default cursor when not selectable', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={false} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card').style.cursor).toBe('default');
  });

  it('has opacity 0.5 when not selectable', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={false} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card').style.opacity).toBe('0.5');
  });

  it('has opacity 1 when selectable', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={true} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card').style.opacity).toBe('1');
  });

  it('has selected border when selected', () => {
    render(<DoubtHandCard card={card} index={0} selected={true} selectable={true} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card').style.border).toBe('3px solid var(--color-game-card-selected)');
  });

  it('has transparent border when not selected', () => {
    render(<DoubtHandCard card={card} index={0} selected={false} selectable={true} onToggle={vi.fn()} />);
    expect(screen.getByTestId('hand-card').style.border).toBe('3px solid transparent');
  });

  it('calls onToggle with index when clicked', () => {
    const onToggle = vi.fn();
    render(<DoubtHandCard card={card} index={3} selected={false} selectable={true} onToggle={onToggle} />);
    fireEvent.click(screen.getByTestId('hand-card'));
    expect(onToggle).toHaveBeenCalledWith(3);
  });
});
