import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SkeletonGrid } from './SkeletonGrid';

describe('SkeletonGrid', () => {
  it('renders correct number of cells', () => {
    const { container } = render(<SkeletonGrid count={12} cols="grid-cols-4" />);
    const cells = container.querySelectorAll('.animate-pulse');
    expect(cells).toHaveLength(12);
  });

  it('applies cols class to grid container', () => {
    const { container } = render(<SkeletonGrid count={4} cols="grid-cols-4 sm:grid-cols-6" />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).toContain('grid-cols-4');
    expect(el.className).toContain('sm:grid-cols-6');
  });

  it('uses default aspect ratio', () => {
    const { container } = render(<SkeletonGrid count={1} cols="grid-cols-1" />);
    const cell = container.querySelector('.animate-pulse') as HTMLElement;
    expect(cell.className).toContain('aspect-[2/3]');
  });

  it('uses custom aspect ratio', () => {
    const { container } = render(<SkeletonGrid count={1} cols="grid-cols-1" aspectRatio="aspect-square" />);
    const cell = container.querySelector('.animate-pulse') as HTMLElement;
    expect(cell.className).toContain('aspect-square');
  });

  it('applies gridClassName to grid container', () => {
    const { container } = render(
      <SkeletonGrid count={4} cols="grid-cols-4" gridClassName="lg:grid-rows-4 lg:h-full" />,
    );
    const el = container.firstChild as HTMLElement;
    expect(el.className).toContain('lg:grid-rows-4');
    expect(el.className).toContain('lg:h-full');
  });

  it('omits trailing space when gridClassName is empty', () => {
    const { container } = render(<SkeletonGrid count={1} cols="grid-cols-1" />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).not.toMatch(/\s$/);
  });
});
