/**
 * @vitest-environment jsdom
 */
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { gameRoutes } from '../../constants/gameRoutes';
import { RecommendationCard } from './RecommendationCard';

const blackJack = gameRoutes.find((r) => r.path === '/');
if (!blackJack) throw new Error('test fixture: BlackJack route missing');

describe('RecommendationCard', () => {
  it('renders a hero with a heading-level title', () => {
    render(
      <MemoryRouter>
        <RecommendationCard game={blackJack} variant="hero" topAxis="mood" />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading')).toBeInTheDocument();
    // Play link points to the game path.
    expect(screen.getByRole('link')).toHaveAttribute('href', blackJack.path);
  });

  it('renders a row variant as a link without a heading', () => {
    render(
      <MemoryRouter>
        <RecommendationCard game={blackJack} variant="row" topAxis="mood" />
      </MemoryRouter>,
    );
    expect(screen.queryByRole('heading')).not.toBeInTheDocument();
    expect(screen.getByRole('link')).toHaveAttribute('href', blackJack.path);
  });
});
