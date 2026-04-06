import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card } from '../../types/card';
import { AnimatedCard } from './AnimatedCard';

const mockCard: Card = { design: 'SPADE', value: 1 };

vi.mock('../../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

import { useReducedMotion } from '../../hooks/useReducedMotion';

describe('AnimatedCard', () => {
  it('renders animated wrapper when motion is enabled', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCard card={mockCard} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('renders plain CardImage when reduced motion is preferred', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCard card={mockCard} />);
    expect(screen.queryByTestId('animated-card')).not.toBeInTheDocument();
    expect(screen.getByRole('img')).toBeInTheDocument();
  });

  it('passes width and style to CardImage', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCard card={mockCard} width={60} style={{ opacity: 0.5 }} />);
    const img = screen.getByRole('img');
    expect(img).toHaveStyle({ width: '60px', opacity: '0.5' });
  });

  it('passes className to CardImage', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    render(<AnimatedCard card={mockCard} className="test-class" />);
    expect(screen.getByRole('img')).toHaveClass('test-class');
  });

  it('applies dealDelay and isSelected defaults', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCard card={mockCard} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('applies custom dealDelay and isSelected', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    render(<AnimatedCard card={mockCard} dealDelay={0.2} isSelected={true} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('passes drag props to CardImage in reduced mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onDragStart = vi.fn();
    render(<AnimatedCard card={mockCard} draggable onDragStart={onDragStart} />);
    expect(screen.getByRole('img')).toHaveAttribute('draggable', 'true');
  });

  it('passes drag props to CardImage in animated mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onDragStart = vi.fn();
    render(<AnimatedCard card={mockCard} draggable onDragStart={onDragStart} />);
    expect(screen.getByRole('img')).toHaveAttribute('draggable', 'true');
  });

  it('accepts onDealComplete callback prop', () => {
    vi.mocked(useReducedMotion).mockReturnValue(false);
    const onDealComplete = vi.fn();
    render(<AnimatedCard card={mockCard} onDealComplete={onDealComplete} />);
    expect(screen.getByTestId('animated-card')).toBeInTheDocument();
  });

  it('does not pass onDealComplete in reduced motion mode', () => {
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const onDealComplete = vi.fn();
    render(<AnimatedCard card={mockCard} onDealComplete={onDealComplete} />);
    expect(screen.queryByTestId('animated-card')).not.toBeInTheDocument();
  });
});
