import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { BaccaratSkeleton } from './BaccaratSkeleton';

describe('BaccaratSkeleton', () => {
  it('renders skeleton structure', () => {
    render(<BaccaratSkeleton />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.getByTestId('skeleton').getAttribute('role')).toBe('status');
  });
});
