import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card } from '../types/card';
import { PlayerHandSection } from './PlayerHandSection';

/** Helper to create N cards for testing. */
function makeCards(n: number): Card[] {
  const suits: Card['design'][] = ['SPADE', 'HEART', 'DIAMOND', 'CLOVER'];
  return Array.from({ length: n }, (_, i) => ({
    design: suits[i % 4],
    value: (i % 13) + 1,
  }));
}

const baseProps = {
  humanPlayer: { cards: makeCards(5) },
  selectedCardIndices: [] as number[],
  toggleCard: vi.fn(),
  cardWidth: 40,
  dataTutorialPrefix: 'ht',
};

describe('PlayerHandSection (desktop)', () => {
  it('renders one button per card', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} />);
    expect(screen.getAllByRole('button')).toHaveLength(5);
  });

  it('applies data-tutorial attribute from prefix', () => {
    const { container } = render(<PlayerHandSection {...baseProps} isMobile={false} />);
    expect(container.querySelector('[data-tutorial="ht-player-hand"]')).toBeInTheDocument();
  });

  it('marks selected cards with aria-pressed=true', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} selectedCardIndices={[0, 2]} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons[0]).toHaveAttribute('aria-pressed', 'true');
    expect(buttons[1]).toHaveAttribute('aria-pressed', 'false');
    expect(buttons[2]).toHaveAttribute('aria-pressed', 'true');
  });

  it('calls toggleCard with correct index on click', () => {
    const toggleCard = vi.fn();
    render(<PlayerHandSection {...baseProps} isMobile={false} toggleCard={toggleCard} />);
    fireEvent.click(screen.getAllByRole('button')[3]);
    expect(toggleCard).toHaveBeenCalledWith(3);
  });

  it('uses sp prefix for data-tutorial when dataTutorialPrefix is sp', () => {
    const { container } = render(<PlayerHandSection {...baseProps} isMobile={false} dataTutorialPrefix="sp" />);
    expect(container.querySelector('[data-tutorial="sp-player-hand"]')).toBeInTheDocument();
  });

  it('applies lg:flex-nowrap and lg:overflow-x-auto classes for desktop single-row layout', () => {
    const { container } = render(<PlayerHandSection {...baseProps} isMobile={false} />);
    const hand = container.querySelector('[data-tutorial="ht-player-hand"]');
    expect(hand).toHaveClass('lg:flex-nowrap');
    expect(hand).toHaveClass('lg:overflow-x-auto');
  });
});

describe('PlayerHandSection (mobile)', () => {
  it('renders MobileHandGrid with correct data-tutorial', () => {
    const { container } = render(<PlayerHandSection {...baseProps} isMobile={true} />);
    expect(container.querySelector('[data-tutorial="ht-player-hand"]')).toBeInTheDocument();
  });

  it('renders all cards via MobileHandGrid', () => {
    render(<PlayerHandSection {...baseProps} isMobile={true} />);
    expect(screen.getAllByRole('button')).toHaveLength(5);
  });

  it('calls toggleCard when a card is tapped in mobile mode', () => {
    const toggleCard = vi.fn();
    render(<PlayerHandSection {...baseProps} isMobile={true} toggleCard={toggleCard} />);
    fireEvent.click(screen.getAllByRole('button')[1]);
    expect(toggleCard).toHaveBeenCalledWith(1);
  });
});
