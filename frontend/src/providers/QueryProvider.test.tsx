import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { QueryProvider } from './QueryProvider';

describe('QueryProvider', () => {
  it('renders children inside QueryClientProvider', () => {
    render(
      <QueryProvider>
        <span>hello</span>
      </QueryProvider>,
    );
    expect(screen.getByText('hello')).toBeInTheDocument();
  });
});
