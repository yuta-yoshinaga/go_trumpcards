import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
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
