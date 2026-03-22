import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { NapoleonSkeleton } from './NapoleonSkeleton';

describe('NapoleonSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<NapoleonSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
