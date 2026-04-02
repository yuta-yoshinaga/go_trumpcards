import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../test/renderWithProviders';
import type { GoFishBook } from '../../types/card';
import { GoFishBooksDisplay } from './GoFishBooksDisplay';

describe('GoFishBooksDisplay', () => {
  it('renders nothing when books array is empty', () => {
    const { container } = renderWithProviders(<GoFishBooksDisplay books={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders books with rank names', () => {
    const books: GoFishBook[] = [
      { rank: 1, cards: [] },
      { rank: 7, cards: [] },
      { rank: 13, cards: [] },
    ];
    renderWithProviders(<GoFishBooksDisplay books={books} />);
    expect(screen.getByText('A')).toBeInTheDocument();
    expect(screen.getByText('7')).toBeInTheDocument();
    expect(screen.getByText('K')).toBeInTheDocument();
  });

  it('renders book count in title', () => {
    const books: GoFishBook[] = [{ rank: 5, cards: [] }];
    renderWithProviders(<GoFishBooksDisplay books={books} />);
    expect(screen.getByText(/ブック.*1/)).toBeInTheDocument();
  });
});
