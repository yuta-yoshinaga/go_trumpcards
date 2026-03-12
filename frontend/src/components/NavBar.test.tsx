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

  it('renders JA and EN language toggle buttons with aria-labels and aria-pressed', () => {
    renderNavBar();
    const jaBtn = screen.getByRole('button', { name: i18n.t('nav.switchToJa') });
    const enBtn = screen.getByRole('button', { name: i18n.t('nav.switchToEn') });
    expect(jaBtn).toBeInTheDocument();
    expect(enBtn).toBeInTheDocument();
    expect(jaBtn).toHaveAttribute('aria-pressed', 'true');
    expect(enBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('switches language to EN and updates aria-pressed when EN button is clicked', () => {
    renderNavBar();
    fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.switchToEn') }));
    expect(i18n.language).toBe('en');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToEn') })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToJa') })).toHaveAttribute('aria-pressed', 'false');
  });

  it('switches language back to JA and updates aria-pressed when JA button is clicked', () => {
    renderNavBar();
    fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.switchToEn') }));
    fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.switchToJa') }));
    expect(i18n.language).toBe('ja');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToJa') })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: i18n.t('nav.switchToEn') })).toHaveAttribute('aria-pressed', 'false');
  });
});
