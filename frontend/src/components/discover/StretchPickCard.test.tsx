/**
 * @vitest-environment jsdom
 */
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { gameRoutes } from '../../constants/gameRoutes';
import { StretchPickCard } from './StretchPickCard';

const briscola = gameRoutes.find((r) => r.path === '/briscola');
if (!briscola) throw new Error('test fixture: Briscola route missing');

describe('StretchPickCard', () => {
  it('renders the stretch eyebrow + a link to the game', () => {
    render(
      <MemoryRouter>
        <StretchPickCard game={briscola} />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading')).toBeInTheDocument();
    expect(screen.getByRole('link')).toHaveAttribute('href', briscola.path);
  });
});
