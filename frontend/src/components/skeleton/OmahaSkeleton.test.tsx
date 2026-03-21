import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { OmahaSkeleton } from './OmahaSkeleton';

describe('OmahaSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<OmahaSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
