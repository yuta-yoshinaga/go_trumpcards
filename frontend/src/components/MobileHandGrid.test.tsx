import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card } from '../types/card';
import { MobileHandGrid } from './MobileHandGrid';

vi.mock('../hooks/useReducedMotion', () => ({
  useReducedMotion: vi.fn(() => false),
}));

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

  it('applies scroll fallback when many cards on narrow viewport exceed minimum tap target', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      // 15 cards split into 2 rows: ceil(15/2)=8 cards per row.
      // buttonWidth = 40+6 = 46. 8*46=368, available=375-32=343 => overlap needed = 25/7 ≈ 3.6px.
      // But with even more cards in a row this should trigger scroll.
      // Use 20 cards: ceil(20/2)=10 per row. 10*46=460 >> 343 => overlap = 117/9 ≈ 13px.
      // 46 - 13 = 33 visible < 44 minimum → scroll fallback expected.
      const cards = makeCards(20);
      const { container } = render(
        <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
      );
      const rows = container.querySelectorAll('[data-testid="hand-row"]');
      // At least one row should have scroll fallback class
      const hasScroll = Array.from(rows).some((row) => row.classList.contains('overflow-x-auto'));
      expect(hasScroll).toBe(true);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    }
  });

  it('does not apply scroll fallback when cards fit with adequate tap targets', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      // 8 cards: ceil(8/2)=4 per row. 4*46=184, available=343 → cards fit with gap, no scroll.
      const cards = makeCards(8);
      const { container } = render(
        <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
      );
      const rows = container.querySelectorAll('[data-testid="hand-row"]');
      const hasScroll = Array.from(rows).some((row) => row.classList.contains('overflow-x-auto'));
      expect(hasScroll).toBe(false);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    }
  });

  it('card buttons have flex-shrink-0 in scroll fallback mode', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      const cards = makeCards(20);
      const { container } = render(
        <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
      );
      const scrollRow = container.querySelector('.overflow-x-auto');
      expect(scrollRow).not.toBeNull();
      const buttons = scrollRow?.querySelectorAll('button') ?? [];
      // Buttons in scroll row should have flex-shrink: 0
      for (const btn of buttons) {
        expect(btn.style.flexShrink).toBe('0');
      }
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    }
  });

  it('expands margin for neighbors of selected card', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      const cards = makeCards(10);
      // Select card at index 3 (4th card in first row of 5)
      const { container } = render(
        <MobileHandGrid cards={cards} selectedIndices={[3]} onToggle={() => {}} cardWidth={40} />,
      );
      const rows = container.querySelectorAll('[data-testid="hand-row"]');
      const firstRowButtons = rows[0].querySelectorAll('button');
      // Card at index 3 (selected) and 4 (right of selected) both expand — should have wider margin
      const neighborMargin = Number.parseFloat(firstRowButtons[4].style.marginLeft);
      // Card at index 1 is not selected or adjacent-to-selected — should have base margin
      const baseMargin = Number.parseFloat(firstRowButtons[1].style.marginLeft);
      expect(neighborMargin).toBeGreaterThan(baseMargin);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    }
  });

  it('does not expand first card margin even if neighbor is selected', () => {
    const cards = makeCards(6);
    // Select card at index 0 — card at index 1 is neighbor, but card 0 (i=0) always has marginLeft=0
    render(<MobileHandGrid cards={cards} selectedIndices={[0]} onToggle={() => {}} cardWidth={40} />);
    const buttons = screen.getAllByRole('button');
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

  it('renders CardImage instead of AnimatedCard (no animated-card testid)', () => {
    const cards = makeCards(5);
    const { container } = render(
      <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
    );
    expect(container.querySelectorAll('[data-testid="animated-card"]')).toHaveLength(0);
    expect(container.querySelectorAll('img')).toHaveLength(5);
  });

  it('applies deal-in animation class to card images', () => {
    const cards = makeCards(3);
    const { container } = render(
      <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
    );
    const images = container.querySelectorAll('img');
    for (const img of images) {
      expect(img.className).toContain('animate-card-deal-in');
    }
  });

  it('skips deal-in animation when reduced motion is preferred', async () => {
    const { useReducedMotion } = await import('../hooks/useReducedMotion');
    vi.mocked(useReducedMotion).mockReturnValue(true);
    const cards = makeCards(3);
    const { container } = render(
      <MobileHandGrid cards={cards} selectedIndices={[]} onToggle={() => {}} cardWidth={40} />,
    );
    const images = container.querySelectorAll('img');
    for (const img of images) {
      expect(img.className || '').not.toContain('animate-card-deal-in');
      expect(img.style.animationDelay).toBe('');
    }
    vi.mocked(useReducedMotion).mockReturnValue(false);
  });
});
