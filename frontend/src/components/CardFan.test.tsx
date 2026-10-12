import { render, screen } from '@testing-library/react';
import { createElement } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { CardFan } from './CardFan';

let observedWidth = 1008;
class MockResizeObserver {
  private callback: ResizeObserverCallback;
  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
  }
  observe(target: Element) {
    Object.defineProperty(target, 'clientWidth', { configurable: true, value: observedWidth });
    this.callback([], this as unknown as ResizeObserver);
  }
  disconnect() {}
  unobserve() {}
}
vi.stubGlobal('ResizeObserver', MockResizeObserver);
const cards = () =>
  Array.from({ length: 13 }, (_, i) => (
    <button key={i} type="button" data-testid={`card-${i}`}>
      card
    </button>
  ));

describe('CardFan', () => {
  it('uses legacy fallback markup before width measurement', () => {
    vi.stubGlobal('ResizeObserver', undefined);
    const { container } = render(<CardFan itemWidth={106}>{cards()}</CardFan>);
    expect(container.firstElementChild).toHaveClass('flex-wrap', 'gap-2');
    vi.stubGlobal('ResizeObserver', MockResizeObserver);
  });
  it('adds the data-tutorial attribute when dataTutorial is provided', () => {
    vi.stubGlobal('ResizeObserver', undefined);
    render(
      <CardFan itemWidth={106} dataTutorial="ca-player-hand">
        {cards()}
      </CardFan>,
    );
    expect(document.querySelector('[data-tutorial="ca-player-hand"]')).toBeInTheDocument();
    vi.stubGlobal('ResizeObserver', MockResizeObserver);
  });
  it('overlaps one wide row and preserves card elements', () => {
    observedWidth = 1008;
    render(
      <CardFan itemWidth={106} data-testid="fan">
        {cards()}
      </CardFan>,
    );
    expect(screen.getAllByTestId(/^card-/)).toHaveLength(13);
    expect(screen.getByTestId('fan').children).toHaveLength(1);
    expect(screen.getByTestId('card-0').parentElement).toHaveStyle({ marginLeft: '0px' });
    expect(screen.getByTestId('card-1').parentElement?.getAttribute('style')).toContain('margin-left: -');
  });
  it('splits narrow width into multiple rows', () => {
    observedWidth = 300;
    render(<CardFan itemWidth={106}>{cards()}</CardFan>);
    expect(document.querySelector('.flex.flex-nowrap')?.parentElement?.children.length).toBeGreaterThan(1);
  });
  it('uses touch exposure on mobile and desktop exposure at the desktop breakpoint', () => {
    observedWidth = 400;
    window.innerWidth = 375;
    const { unmount } = render(
      <CardFan itemWidth={46} data-testid="mobile-fan">
        {cards()}
      </CardFan>,
    );
    expect(screen.getByTestId('mobile-fan').children).toHaveLength(2);

    unmount();
    window.innerWidth = 800;
    render(
      <CardFan itemWidth={46} data-testid="desktop-fan">
        {cards()}
      </CardFan>,
    );
    expect(screen.getByTestId('desktop-fan').children).toHaveLength(1);
  });
  it('renders children without keys after measuring width', () => {
    observedWidth = 1008;
    const unkeyedCards = Array.from({ length: 2 }, () => createElement('button', { type: 'button' }, 'card'));
    render(<CardFan itemWidth={106}>{unkeyedCards}</CardFan>);
    expect(screen.getAllByRole('button', { name: 'card' })).toHaveLength(2);
  });
});
