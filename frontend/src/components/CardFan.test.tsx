import { render, screen } from '@testing-library/react';
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
});
