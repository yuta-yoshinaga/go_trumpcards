import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card } from '../types/card';
import { MobileHandGrid } from './MobileHandGrid';

/** Helper to create N cards for testing. */
function makeCards(n: number): Card[] {
  const suits: Card['design'][] = ['SPADE', 'HEART', 'DIAMOND', 'CLOVER'];
  return Array.from({ length: n }, (_, i) => ({
    design: suits[i % 4],
    value: (i % 13) + 1,
  }));
}

describe('MobileHandGrid', () => {
  it('renders all cards', () => {
    const cards = makeCards(13);
    render(<MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(13);
  });

  it('splits cards into 2 rows', () => {
    const cards = makeCards(13);
    const { container } = render(
      <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
    );
    const rows = container.querySelectorAll('[data-testid="hand-row"]');
    expect(rows).toHaveLength(2);
    // First row: ceil(13/2) = 7, second row: 6
    expect(rows[0].querySelectorAll('button')).toHaveLength(7);
    expect(rows[1].querySelectorAll('button')).toHaveLength(6);
  });

  it('splits even card count evenly', () => {
    const cards = makeCards(10);
    const { container } = render(
      <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
    );
    const rows = container.querySelectorAll('[data-testid="hand-row"]');
    expect(rows).toHaveLength(2);
    expect(rows[0].querySelectorAll('button')).toHaveLength(5);
    expect(rows[1].querySelectorAll('button')).toHaveLength(5);
  });

  it('calls onToggle with correct index when card is clicked', () => {
    const onToggle = vi.fn();
    const cards = makeCards(13);
    render(<MobileHandGrid cards={cards} selectedIndices={[]} onToggle={onToggle} cardWidth={40} />);
    const buttons = screen.getAllByRole('button');
    // Click 3rd card (index 2) in first row
    fireEvent.click(buttons[2]);
    expect(onToggle).toHaveBeenCalledWith(2);
    // Click 1st card (index 7) in second row — ceil(13/2)=7, so second row starts at 7
    fireEvent.click(buttons[7]);
    expect(onToggle).toHaveBeenCalledWith(7);
  });

  it('marks selected cards with aria-pressed', () => {
    const cards = makeCards(13);
    render(<MobileHandGrid cards={cards} selectedIndices={[0, 5, 12]} onToggle={() => {}} cardWidth={40} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons[0]).toHaveAttribute('aria-pressed', 'true');
    expect(buttons[1]).toHaveAttribute('aria-pressed', 'false');
    expect(buttons[5]).toHaveAttribute('aria-pressed', 'true');
    expect(buttons[12]).toHaveAttribute('aria-pressed', 'true');
  });

  it('passes dataTutorial as data-tutorial attribute', () => {
    const cards = makeCards(5);
    const { container } = render(
      <MobileHandGrid
        cards={cards}
        selectedIndices={[]}
        onToggle={() => {}}
        cardWidth={40}
        dataTutorial="ht-player-hand"
      />,
    );
    const wrapper = container.firstElementChild as HTMLElement;
    expect(wrapper).toHaveAttribute('data-tutorial', 'ht-player-hand');
  });

  it('renders single row when 3 or fewer cards', () => {
    const cards = makeCards(3);
    const { container } = render(
      <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
    );
    const rows = container.querySelectorAll('[data-testid="hand-row"]');
    expect(rows).toHaveLength(1);
    expect(rows[0].querySelectorAll('button')).toHaveLength(3);
  });

  it('renders no card buttons when cards array is empty', () => {
    render(<MobileHandGrid cards={[]} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />);
    expect(screen.queryAllByRole('button')).toHaveLength(0);
  });

  it('renders single card without overlap (computeOverlap cardCount <= 1)', () => {
    const cards = makeCards(1);
    render(<MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(1);
    // Single card should have no marginLeft
    expect(buttons[0].style.marginLeft).toBe('0px');
  });

  it('uses positive gap when viewport is wide enough for all cards', () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    window.dispatchEvent(new Event('resize'));
    try {
      const cards = makeCards(3);
      render(<MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />);
      const buttons = screen.getAllByRole('button');
      // Second card should have positive marginLeft (DEFAULT_CARD_GAP = 2px)
      expect(Number.parseFloat(buttons[1].style.marginLeft)).toBeGreaterThan(0);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });
});
