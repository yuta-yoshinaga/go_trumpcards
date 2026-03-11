import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it } from 'vitest';
import { gameRoutes } from '../constants/gameRoutes';
import { NavBar } from './NavBar';

function renderNavBar(initialPath = '/') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <NavBar />
    </MemoryRouter>,
  );
}

function labelFor(labelKey: string): string {
  return i18n.t(labelKey);
}

afterEach(() => {
  i18n.changeLanguage('ja');
});

describe('NavBar', () => {
  it('renders navigation links for all game routes', () => {
    renderNavBar();
    const links = screen.getAllByRole('link');
    expect(links).toHaveLength(gameRoutes.length);
  });

  for (const { path, labelKey } of gameRoutes) {
    it(`renders ${labelKey} link`, () => {
      renderNavBar();
      expect(screen.getByText(labelFor(labelKey))).toBeInTheDocument();
    });

    it(`marks ${labelKey} link as active when on ${path}`, () => {
      renderNavBar(path);
      expect(screen.getByText(labelFor(labelKey))).toHaveAttribute('aria-current', 'page');
      for (const other of gameRoutes) {
        if (other.path !== path) {
          expect(screen.getByText(labelFor(other.labelKey))).not.toHaveAttribute('aria-current');
        }
      }
    });
  }

  it('links point to correct hrefs', () => {
    renderNavBar();
    for (const { path, labelKey } of gameRoutes) {
      expect(screen.getByRole('link', { name: labelFor(labelKey) })).toHaveAttribute('href', path);
    }
  });

  it('renders JA and EN language toggle buttons', () => {
    renderNavBar();
    expect(screen.getByRole('button', { name: '日本語に切り替え' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Switch to English' })).toBeInTheDocument();
  });

  it('sets aria-pressed=true on JA button when language is ja', () => {
    renderNavBar();
    expect(screen.getByRole('button', { name: '日本語に切り替え' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'Switch to English' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('switches language to EN when EN button is clicked', () => {
    renderNavBar();
    fireEvent.click(screen.getByRole('button', { name: 'Switch to English' }));
    expect(i18n.language).toBe('en');
  });

  it('sets aria-pressed=true on EN button when language is en', () => {
    renderNavBar();
    fireEvent.click(screen.getByRole('button', { name: 'Switch to English' }));
    expect(screen.getByRole('button', { name: 'Switch to English' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: '日本語に切り替え' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('switches language back to JA when JA button is clicked', () => {
    renderNavBar();
    fireEvent.click(screen.getByRole('button', { name: 'Switch to English' }));
    fireEvent.click(screen.getByRole('button', { name: '日本語に切り替え' }));
    expect(i18n.language).toBe('ja');
  });
});
