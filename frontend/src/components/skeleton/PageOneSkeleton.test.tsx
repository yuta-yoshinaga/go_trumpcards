import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { PageOneSkeleton } from './PageOneSkeleton';

describe('PageOneSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<PageOneSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
