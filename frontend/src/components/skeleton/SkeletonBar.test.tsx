import { describe, expect, it } from 'bun:test';
import { render } from '@testing-library/react';
import { SkeletonBar } from './SkeletonBar';

describe('SkeletonBar', () => {
  it('renders with default height', () => {
    const { container } = render(<SkeletonBar />);
    const el = container.firstChild as HTMLElement;
    expect(el).toBeInTheDocument();
    expect(el.className).toContain('h-9');
    expect(el.className).toContain('animate-pulse');
  });

  it('renders with custom height', () => {
    const { container } = render(<SkeletonBar height="h-12" />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).toContain('h-12');
  });

  it('applies additional className', () => {
    const { container } = render(<SkeletonBar className="extra" />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).toContain('extra');
  });

  it('renders without additional className', () => {
    const { container } = render(<SkeletonBar />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).not.toContain('undefined');
  });
});
