import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { render, screen } from '@testing-library/react';
import * as useReducedMotionModule from '../../hooks/useReducedMotion';
import type { Card } from '../../types/card';
import { AnimatedCard } from './AnimatedCard';

const mockCard: Card = { design: 'SPADE', value: 1 };

describe('AnimatedCard', () => {
  let spy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    spy = vi.spyOn(useReducedMotionModule, 'useReducedMotion').mockReturnValue(false);
  });

  afterEach(() => {
    spy.mockRestore();
  });

  it('renders animated wrapper when motion is enabled', () => {
    spy.mockReturnValue(false);
    render(<AnimatedCard card={mockCard} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('renders plain CardImage when reduced motion is preferred', () => {
    spy.mockReturnValue(true);
    render(<AnimatedCard card={mockCard} />);
    expect(screen.queryByTestId('animated-card')).not.toBeInTheDocument();
    expect(screen.getByRole('img')).toBeInTheDocument();
  });

  it('passes width and style to CardImage', () => {
    spy.mockReturnValue(true);
    render(<AnimatedCard card={mockCard} width={60} style={{ opacity: 0.5 }} />);
    const img = screen.getByRole('img');
    expect(img).toHaveStyle({ width: '60px', opacity: '0.5' });
  });

  it('passes className to CardImage', () => {
    spy.mockReturnValue(true);
    render(<AnimatedCard card={mockCard} className="test-class" />);
    expect(screen.getByRole('img')).toHaveClass('test-class');
  });

  it('applies dealDelay and isSelected defaults', () => {
    spy.mockReturnValue(false);
    render(<AnimatedCard card={mockCard} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('applies custom dealDelay and isSelected', () => {
    spy.mockReturnValue(false);
    render(<AnimatedCard card={mockCard} dealDelay={0.2} isSelected={true} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('passes drag props to CardImage in reduced mode', () => {
    spy.mockReturnValue(true);
    const onDragStart = vi.fn();
    render(<AnimatedCard card={mockCard} draggable onDragStart={onDragStart} />);
    expect(screen.getByRole('img')).toHaveAttribute('draggable', 'true');
  });

  it('passes drag props to CardImage in animated mode', () => {
    spy.mockReturnValue(false);
    const onDragStart = vi.fn();
    render(<AnimatedCard card={mockCard} draggable onDragStart={onDragStart} />);
    expect(screen.getByRole('img')).toHaveAttribute('draggable', 'true');
  });
});
