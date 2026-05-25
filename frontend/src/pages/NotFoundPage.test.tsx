/**
 * @vitest-environment jsdom
 */
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { NotFoundPage } from './NotFoundPage';

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('NotFoundPage (#1902)', () => {
  it('renders the not-found heading instead of silently redirecting', () => {
    renderAt('/nonexistent');
    expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
  });

  it('shows the requested path so the user can spot a typo', () => {
    renderAt('/games/xyz');
    expect(screen.getByText(/\/games\/xyz/)).toBeInTheDocument();
  });

  it('offers a Discover CTA linking to /discover', () => {
    renderAt('/foo/bar/baz');
    const link = screen.getAllByRole('link').find((a) => a.getAttribute('href') === '/discover');
    expect(link).toBeDefined();
  });

  it('offers a Home CTA linking to /', () => {
    renderAt('/foo/bar/baz');
    const link = screen.getAllByRole('link').find((a) => a.getAttribute('href') === '/');
    expect(link).toBeDefined();
  });
});
