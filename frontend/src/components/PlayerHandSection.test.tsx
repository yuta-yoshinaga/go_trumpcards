import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Card } from '../types/card';
import { PlayerHandSection } from './PlayerHandSection';

let observedHandWidth = 640;
const observedTargets: Element[] = [];
class MockResizeObserver {
  private callback: ResizeObserverCallback;
  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
  }
  observe(target: Element) {
    observedTargets.push(target);
    Object.defineProperty(target, 'clientWidth', { configurable: true, value: observedHandWidth });
    this.callback([], this as unknown as ResizeObserver);
  }
  disconnect() {}
  unobserve() {}
}
vi.stubGlobal('ResizeObserver', MockResizeObserver);

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
  it('appends the accessible label only to trump cards', () => {
    const cards: Card[] = [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 2 },
    ];
    render(
      <PlayerHandSection
        {...baseProps}
        humanPlayer={{ cards }}
        isMobile={false}
        trumpIndices={[0]}
        trumpAccessibleLabel="切り札"
      />,
    );
    const buttons = screen.getAllByRole('button');
    expect(buttons[0]).toHaveAccessibleName(/ \(切り札\)$/);
    expect(buttons[1]).not.toHaveAccessibleName(/切り札/);
  });

  it('does not append a suffix to trump cards when trumpAccessibleLabel is omitted', () => {
    const cards: Card[] = [{ design: 'SPADE', value: 1 }];
    render(<PlayerHandSection {...baseProps} humanPlayer={{ cards }} isMobile={false} trumpIndices={[0]} />);
    expect(screen.getByRole('button')).not.toHaveAccessibleName(/ \([^)]*\)$/);
  });

  it('includes the trump label and role badge in the accessible name', () => {
    render(
      <PlayerHandSection
        {...baseProps}
        humanPlayer={{ cards: [{ design: 'SPADE', value: 13 }] }}
        isMobile={false}
        trumpIndices={[0]}
        trumpAccessibleLabel="切り札"
        cardBadgeFor={() => ({ glyph: 'M', title: 'マタドール' })}
      />,
    );
    expect(screen.getByRole('button')).toHaveAccessibleName(/♠ K \(切り札\) \(マタドール\)/);
    expect(screen.getByText('M')).toBeInTheDocument();
  });

  it('renders legal cards with the success ring', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} legalIndices={[0]} />);
    expect(screen.getAllByRole('button')[0]).toHaveClass('ring-2', 'ring-ds-success');
    expect(screen.getAllByRole('button')[1]).not.toHaveClass('ring-ds-success');
  });

  it('uses the flexible wrapped layout below desktop width', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    window.dispatchEvent(new Event('resize'));
    const { container } = render(<PlayerHandSection {...baseProps} isMobile={false} />);
    const hand = container.querySelector('[data-tutorial="ht-player-hand"]');
    expect(hand).toHaveClass('flex', 'flex-wrap');
    expect(hand).not.toHaveClass('flex-col');
    expect(hand?.firstElementChild).toHaveClass('flex-wrap', 'gap-1');
    expect(screen.getAllByRole('button')[0].style.marginLeft).toBe('');
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1200 });
    window.dispatchEvent(new Event('resize'));
  });

  it('renders one button per card', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} />);
    expect(screen.getAllByRole('button')).toHaveLength(5);
  });

  it('applies data-tutorial attribute from prefix', () => {
    const { container } = render(<PlayerHandSection {...baseProps} isMobile={false} />);
    expect(container.querySelector('[data-tutorial="ht-player-hand"]')).toBeInTheDocument();
  });

  it('observes the hand node when cards are added after an empty render', () => {
    observedTargets.length = 0;
    const { rerender, container } = render(
      <PlayerHandSection {...baseProps} humanPlayer={{ cards: [] }} isMobile={true} />,
    );
    rerender(<PlayerHandSection {...baseProps} isMobile={false} />);
    const hand = container.querySelector('[data-tutorial="ht-player-hand"]');
    expect(observedTargets).toContain(hand);
  });

  it('marks selected cards with aria-pressed=true', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} selectedCardIndices={[0, 2]} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons[0]).toHaveAttribute('aria-pressed', 'true');
    expect(buttons[1]).toHaveAttribute('aria-pressed', 'false');
    expect(buttons[2]).toHaveAttribute('aria-pressed', 'true');
  });

  it('does not raise selected cards above neighboring cards', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} selectedCardIndices={[0]} />);

    const buttons = screen.getAllByRole('button');
    expect(buttons[0]).toHaveAttribute('aria-pressed', 'true');
    expect(buttons[0]).not.toHaveStyle({ zIndex: '3' });
    expect(buttons[1]).toHaveAttribute('aria-pressed', 'false');
    expect(buttons[1]).not.toHaveStyle({ zIndex: '3' });
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

  it('overlaps cards only when needed on wide desktop layouts', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1050 });
    window.dispatchEvent(new Event('resize'));
    const { container, rerender } = render(
      <PlayerHandSection {...baseProps} humanPlayer={{ cards: makeCards(30) }} isMobile={false} />,
    );
    const hand = container.querySelector('[data-tutorial="ht-player-hand"]');
    expect(hand).toHaveClass('lg:overflow-x-auto');
    expect(container.querySelectorAll('[data-hand-card-index="0"]')).toHaveLength(1);
    expect(screen.getAllByRole('button')[1].style.marginLeft).toMatch(/-/);
    rerender(<PlayerHandSection {...baseProps} isMobile={false} />);
    expect(screen.getAllByRole('button')[1].style.marginLeft).toBe('4px');
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1200 });
    window.dispatchEvent(new Event('resize'));
  });

  it('keeps overlapped desktop cards out of the hover stacking order but raises focused cards', () => {
    observedHandWidth = 120;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1050 });
    window.dispatchEvent(new Event('resize'));

    render(<PlayerHandSection {...baseProps} isMobile={false} />);

    const card = screen.getAllByRole('button')[0];
    expect(card).not.toHaveClass('hover:z-10');
    expect(card).toHaveClass('focus-visible:z-10');

    observedHandWidth = 640;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1200 });
    window.dispatchEvent(new Event('resize'));
  });

  it('splits into two desktop rows when 28px exposure cannot fit one row', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1050 });
    window.dispatchEvent(new Event('resize'));
    const { container } = render(
      <PlayerHandSection {...baseProps} humanPlayer={{ cards: makeCards(60) }} isMobile={false} />,
    );
    const hand = container.querySelector('[data-tutorial="ht-player-hand"]');
    expect(hand?.children.length).toBeGreaterThan(2);
    const rows = hand?.children;
    expect(rows?.[0].querySelectorAll('button').length).toBeLessThanOrEqual(22);
    expect(rows?.[0].querySelector('button:nth-child(2)')?.getAttribute('style')).toContain('margin-left');
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1200 });
    window.dispatchEvent(new Event('resize'));
  });

  it('puts the remainder card in the first of two rows', () => {
    observedHandWidth = 500;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
    window.dispatchEvent(new Event('resize'));
    const { container } = render(
      <PlayerHandSection {...baseProps} humanPlayer={{ cards: makeCards(19) }} isMobile={false} />,
    );
    const rows = container.querySelectorAll('[data-tutorial="ht-player-hand"] > div');
    expect(rows).toHaveLength(2);
    expect(rows[0].querySelectorAll('button')).toHaveLength(10);
    expect(rows[1].querySelectorAll('button')).toHaveLength(9);
    observedHandWidth = 640;
  });

  it('uses the hand container width for overlap even when the window is wider', () => {
    observedHandWidth = 500;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
    const { container } = render(
      <PlayerHandSection {...baseProps} humanPlayer={{ cards: makeCards(30) }} isMobile={false} />,
    );
    const rows = container.querySelectorAll('[data-tutorial="ht-player-hand"] > div');
    expect(rows.length).toBe(2);
    for (const row of rows) {
      const buttons = Array.from(row.querySelectorAll('button'));
      const totalWidth =
        buttons.length * 46 +
        buttons.slice(1).reduce((sum, button) => sum + Number.parseFloat(button.style.marginLeft), 0);
      expect(totalWidth).toBeLessThanOrEqual(500);
    }
    observedHandWidth = 640;
  });

  it('appends each card status to its accessible name when provided', () => {
    render(
      <PlayerHandSection
        {...baseProps}
        isMobile={false}
        cardStatusFor={(idx) => (idx === 1 ? '使用可能' : undefined)}
      />,
    );
    const buttons = screen.getAllByRole('button');
    expect(buttons[1]).toHaveAccessibleName(/\(使用可能\)$/);
    expect(buttons[0]).not.toHaveAccessibleName(/\(使用可能\)$/);
  });

  it('does not append a status to card accessible names when cardStatusFor is omitted', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} />);
    for (const button of screen.getAllByRole('button')) {
      expect(button).not.toHaveAccessibleName(/\(使用可能\)$/);
    }
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

describe('PlayerHandSection (validIndices)', () => {
  it('marks cards outside validIndices as aria-disabled on desktop (still focusable)', () => {
    const toggleCard = vi.fn();
    render(
      <PlayerHandSection
        {...baseProps}
        isMobile={false}
        toggleCard={toggleCard}
        validIndices={[0, 2]}
        restrictedTooltip="Cannot play"
      />,
    );
    const buttons = screen.getAllByRole('button');
    expect(buttons[0]).not.toHaveAttribute('aria-disabled');
    expect(buttons[1]).toHaveAttribute('aria-disabled', 'true');
    expect(buttons[1]).not.toBeDisabled(); // must stay focusable for the tooltip
    expect(buttons[1]).toHaveAttribute('title', 'Cannot play');
    fireEvent.click(buttons[1]);
    expect(toggleCard).not.toHaveBeenCalled();
    fireEvent.click(buttons[0]);
    expect(toggleCard).toHaveBeenCalledWith(0);
  });

  it('marks cards outside validIndices as aria-disabled on mobile (still focusable)', () => {
    const toggleCard = vi.fn();
    render(
      <PlayerHandSection
        {...baseProps}
        isMobile={true}
        toggleCard={toggleCard}
        validIndices={[0, 2]}
        restrictedTooltip="Cannot play"
      />,
    );
    const buttons = screen.getAllByRole('button');
    expect(buttons[1]).toHaveAttribute('aria-disabled', 'true');
    expect(buttons[1]).not.toBeDisabled();
    expect(buttons[1]).toHaveAttribute('title', 'Cannot play');
    fireEvent.click(buttons[1]);
    expect(toggleCard).not.toHaveBeenCalled();
  });

  it('does not mark any card aria-disabled when validIndices is undefined', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} />);
    const buttons = screen.getAllByRole('button');
    for (const btn of buttons) expect(btn).not.toHaveAttribute('aria-disabled');
  });
});

describe('PlayerHandSection (highlightIndices)', () => {
  // JSDOM drops the CSS var from the `border` shorthand, so assert on the
  // parseable boxShadow glow, which is distinct per state:
  //   highlight = rgba(232, 146, 58, …) (warning), selection = rgba(59, 130, 246, …) (blue).
  it('gives highlighted cards a warning glow and dims the rest on desktop', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} highlightIndices={[1, 3]} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons[1].style.boxShadow).toContain('rgba(232, 146, 58');
    expect(buttons[3].style.boxShadow).toContain('rgba(232, 146, 58');
    expect(buttons[0]).toHaveClass('opacity-60');
    expect(buttons[1]).not.toHaveClass('opacity-60');
  });

  it('keeps the selection glow on a card that is both selected and highlighted', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} highlightIndices={[0]} selectedCardIndices={[0]} />);
    const buttons = screen.getAllByRole('button');
    // Selection wins over highlight, and a selected card is never dimmed.
    expect(buttons[0].style.boxShadow).toContain('rgba(59, 130, 246');
    expect(buttons[0]).not.toHaveClass('opacity-60');
  });

  it('highlights and dims on mobile too', () => {
    render(<PlayerHandSection {...baseProps} isMobile={true} highlightIndices={[2]} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons[2].style.boxShadow).toContain('rgba(232, 146, 58');
    expect(buttons[0]).toHaveClass('opacity-60');
  });

  it('applies no highlight or dimming when highlightIndices is undefined', () => {
    render(<PlayerHandSection {...baseProps} isMobile={false} />);
    for (const btn of screen.getAllByRole('button')) expect(btn).not.toHaveClass('opacity-60');
  });
});
