import { afterEach, describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import type { Card } from '../../types/card';
import { OldMaidDiscardedArea } from './OldMaidDiscardedArea';

function makeCard(design: Card['design'], value: number): Card {
  return { design, value };
}

describe('OldMaidDiscardedArea', () => {
  const originalInnerWidth = window.innerWidth;

  afterEach(() => {
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: originalInnerWidth,
    });
  });

  it('shows placeholder when cards is undefined', () => {
    render(<OldMaidDiscardedArea cards={undefined} />);
    expect(screen.getByText('捨て札エリア')).toBeInTheDocument();
  });

  it('shows placeholder when cards is empty', () => {
    render(<OldMaidDiscardedArea cards={[]} />);
    expect(screen.getByText('捨て札エリア')).toBeInTheDocument();
  });

  it('renders paired cards for even number of cards', () => {
    const cards = [makeCard('SPADE', 1), makeCard('HEART', 1), makeCard('DIAMOND', 5), makeCard('CLOVER', 5)];
    const { container } = render(<OldMaidDiscardedArea cards={cards} />);
    const pairContainers = container.querySelectorAll('div[style*="position: relative"]');
    expect(pairContainers).toHaveLength(2);
    expect(screen.getAllByRole('img')).toHaveLength(4);
  });

  it('renders pairs plus a remainder card for odd number of cards', () => {
    const cards = [makeCard('SPADE', 1), makeCard('HEART', 1), makeCard('DIAMOND', 5)];
    const { container } = render(<OldMaidDiscardedArea cards={cards} />);
    const pairContainers = container.querySelectorAll('div[style*="position: relative"]');
    expect(pairContainers).toHaveLength(1);
    expect(screen.getAllByRole('img')).toHaveLength(3);
  });

  it('uses desktop card dimensions on wide viewport', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    const cards = [makeCard('SPADE', 1), makeCard('HEART', 1)];
    const { container } = render(<OldMaidDiscardedArea cards={cards} />);
    const pairContainer = container.querySelector('div[style*="position: relative"]') as HTMLElement;
    // desktop: cardWidth=60, cardHeight=84, overlapLeft=11, overlapTop=7
    expect(pairContainer.style.width).toBe('71px');
    expect(pairContainer.style.height).toBe('91px');
  });

  it('uses mobile card dimensions on narrow viewport', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    const cards = [makeCard('SPADE', 1), makeCard('HEART', 1)];
    const { container } = render(<OldMaidDiscardedArea cards={cards} />);
    const pairContainer = container.querySelector('div[style*="position: relative"]') as HTMLElement;
    // mobile: cardWidth=40, cardHeight=60, overlapLeft=7, overlapTop=4
    expect(pairContainer.style.width).toBe('47px');
    expect(pairContainer.style.height).toBe('64px');
  });

  it('scales overlap offsets proportionally on desktop', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    const cards = [makeCard('SPADE', 1), makeCard('HEART', 1)];
    const { container } = render(<OldMaidDiscardedArea cards={cards} />);
    const overlappedImg = container.querySelectorAll('img')[1] as HTMLElement;
    // desktop cardWidth=60: left=Math.round(60*0.18)=11, top=Math.round(60*0.11)=7
    expect(overlappedImg.style.left).toBe('11px');
    expect(overlappedImg.style.top).toBe('7px');
  });

  it('scales overlap offsets proportionally on mobile', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    const cards = [makeCard('SPADE', 1), makeCard('HEART', 1)];
    const { container } = render(<OldMaidDiscardedArea cards={cards} />);
    const overlappedImg = container.querySelectorAll('img')[1] as HTMLElement;
    // mobile cardWidth=40: left=Math.round(40*0.18)=7, top=Math.round(40*0.11)=4
    expect(overlappedImg.style.left).toBe('7px');
    expect(overlappedImg.style.top).toBe('4px');
  });
});
