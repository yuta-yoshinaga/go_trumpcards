import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { KlondikeSkeleton } from './KlondikeSkeleton';

describe('KlondikeSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<KlondikeSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
