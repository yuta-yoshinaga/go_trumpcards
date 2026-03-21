import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { BaccaratSkeleton } from './BaccaratSkeleton';

describe('BaccaratSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<BaccaratSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('aria-busy')).toBe('true');
  });
});
