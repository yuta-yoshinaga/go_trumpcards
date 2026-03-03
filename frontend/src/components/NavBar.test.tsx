import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { gameRoutes } from '../constants/gameRoutes';
import { NavBar } from './NavBar';

function renderNavBar(initialPath = '/') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <NavBar />
    </MemoryRouter>,
  );
}

describe('NavBar', () => {
  it('renders navigation links for all game routes', () => {
    renderNavBar();
    const links = screen.getAllByRole('link');
    expect(links).toHaveLength(gameRoutes.length);
  });

  for (const { path, label } of gameRoutes) {
    it(`renders ${label} link`, () => {
      renderNavBar();
      expect(screen.getByText(label)).toBeInTheDocument();
    });

    it(`marks ${label} link as active when on ${path}`, () => {
      renderNavBar(path);
      expect(screen.getByText(label)).toHaveAttribute('aria-current', 'page');
      for (const other of gameRoutes) {
        if (other.path !== path) {
          expect(screen.getByText(other.label)).not.toHaveAttribute('aria-current');
        }
      }
    });
  }

  it('links point to correct hrefs', () => {
    renderNavBar();
    const links = screen.getAllByRole('link');
    for (let i = 0; i < gameRoutes.length; i++) {
      expect(links[i]).toHaveAttribute('href', gameRoutes[i].path);
    }
  });
});
