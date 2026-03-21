import { describe, expect, it } from 'bun:test';
import { render } from '@testing-library/react';
import { SkeletonHand } from './SkeletonHand';

describe('SkeletonHand', () => {
  it('renders 5 skeleton cards by default', () => {
    const { container } = render(<SkeletonHand cardWidth={60} cardHeight={84} />);
    const cards = container.querySelectorAll('.animate-pulse');
    expect(cards).toHaveLength(5);
  });

  it('renders specified count of skeleton cards', () => {
    const { container } = render(<SkeletonHand cardWidth={60} cardHeight={84} count={3} />);
    const cards = container.querySelectorAll('.animate-pulse');
    expect(cards).toHaveLength(3);
  });

  it('applies additional className', () => {
    const { container } = render(<SkeletonHand cardWidth={60} cardHeight={84} className="extra" />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).toContain('extra');
  });

  it('renders without additional className', () => {
    const { container } = render(<SkeletonHand cardWidth={60} cardHeight={84} />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).not.toContain('undefined');
  });
});
