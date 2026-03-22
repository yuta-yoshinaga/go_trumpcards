import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SpiderSkeleton } from './SpiderSkeleton';

describe('SpiderSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<SpiderSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
