import { describe, expect, it } from 'bun:test';
import { render } from '@testing-library/react';
import { SkeletonCard } from './SkeletonCard';

describe('SkeletonCard', () => {
  it('renders with correct dimensions and pulse animation', () => {
    const { container } = render(<SkeletonCard width={60} height={84} />);
    const el = container.firstChild as HTMLElement;
    expect(el).toBeInTheDocument();
    expect(el.style.width).toBe('60px');
    expect(el.style.height).toBe('84px');
    expect(el.className).toContain('animate-pulse');
    expect(el.getAttribute('aria-hidden')).toBe('true');
  });

  it('applies additional className', () => {
    const { container } = render(<SkeletonCard width={40} height={60} className="extra" />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).toContain('extra');
  });

  it('renders without additional className', () => {
    const { container } = render(<SkeletonCard width={40} height={60} />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).not.toContain('undefined');
  });
});
