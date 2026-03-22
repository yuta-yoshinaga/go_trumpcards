import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { gameCategories, gameRoutes } from '../constants/gameRoutes';
import { NavBar } from './NavBar';

function renderNavBar(initialPath = '/', props?: { soundMuted?: boolean; onSoundToggle?: () => void }) {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <NavBar {...props} />
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
    // game links + brand link
    expect(links.length).toBeGreaterThanOrEqual(gameRoutes.length);
  });

  it('renders category labels for all categories', () => {
    renderNavBar();
    for (const { labelKey } of gameCategories) {
      expect(screen.getAllByText(labelFor(labelKey)).length).toBeGreaterThanOrEqual(1);
    }
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
    const jaBtns = screen.getAllByRole('button', { name: i18n.t('nav.switchToJa') });
    const enBtns = screen.getAllByRole('button', { name: i18n.t('nav.switchToEn') });
    expect(jaBtns.length).toBeGreaterThanOrEqual(1);
    expect(enBtns.length).toBeGreaterThanOrEqual(1);
    expect(jaBtns[0]).toHaveAttribute('aria-pressed', 'true');
    expect(enBtns[0]).toHaveAttribute('aria-pressed', 'false');
  });

  it('switches language to EN and updates aria-pressed when EN button is clicked', () => {
    renderNavBar();
    fireEvent.click(screen.getAllByRole('button', { name: i18n.t('nav.switchToEn') })[0]);
    expect(i18n.language).toBe('en');
    expect(screen.getAllByRole('button', { name: i18n.t('nav.switchToEn') })[0]).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    expect(screen.getAllByRole('button', { name: i18n.t('nav.switchToJa') })[0]).toHaveAttribute(
      'aria-pressed',
      'false',
    );
  });

  it('switches language back to JA and updates aria-pressed when JA button is clicked', () => {
    renderNavBar();
    fireEvent.click(screen.getAllByRole('button', { name: i18n.t('nav.switchToEn') })[0]);
    fireEvent.click(screen.getAllByRole('button', { name: i18n.t('nav.switchToJa') })[0]);
    expect(i18n.language).toBe('ja');
    expect(screen.getAllByRole('button', { name: i18n.t('nav.switchToJa') })[0]).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    expect(screen.getAllByRole('button', { name: i18n.t('nav.switchToEn') })[0]).toHaveAttribute(
      'aria-pressed',
      'false',
    );
  });

  describe('hamburger menu', () => {
    it('renders hamburger button with openMenu aria-label and aria-controls', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      expect(btn).toBeInTheDocument();
      expect(btn).toHaveAttribute('aria-expanded', 'false');
      expect(btn).toHaveAttribute('aria-controls', 'main-nav');
      expect(btn.querySelector('svg')).toBeInTheDocument();
    });

    it('nav has id matching aria-controls and hidden class by default', () => {
      renderNavBar();
      const nav = screen.getByRole('navigation');
      expect(nav).toHaveAttribute('id', 'main-nav');
      expect(nav).toHaveClass('hidden');
    });

    it('clicking hamburger toggles menu open', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      fireEvent.click(btn);
      expect(btn).toHaveAttribute('aria-expanded', 'true');
      expect(btn.querySelector('svg')).toBeInTheDocument();
      const nav = screen.getByRole('navigation');
      expect(nav).not.toHaveClass('hidden');
    });

    it('clicking hamburger again toggles menu closed and updates aria-label', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      fireEvent.click(btn);
      const closeBtn = screen.getByRole('button', { name: i18n.t('nav.closeMenu') });
      expect(closeBtn).toBeInTheDocument();
      fireEvent.click(closeBtn);
      expect(closeBtn).toHaveAttribute('aria-expanded', 'false');
      expect(screen.getByRole('button', { name: i18n.t('nav.openMenu') })).toBeInTheDocument();
      const nav = screen.getByRole('navigation');
      expect(nav).toHaveClass('hidden');
    });

    it('clicking a game link closes the menu', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      fireEvent.click(btn);
      // Click a non-home link to cover onClick in a different category iteration
      const pokerLink = screen.getByRole('link', { name: labelFor('nav.poker') });
      fireEvent.click(pokerLink);
      const nav = screen.getByRole('navigation');
      expect(nav).toHaveClass('hidden');
      expect(screen.getByRole('button', { name: i18n.t('nav.openMenu') })).toHaveAttribute('aria-expanded', 'false');
    });

    it('renders brand link pointing to home', () => {
      renderNavBar();
      const brandLink = screen.getByRole('link', { name: 'Trump Cards' });
      expect(brandLink).toHaveAttribute('href', '/');
    });

    it('clicking brand link closes the menu', () => {
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      fireEvent.click(screen.getByRole('link', { name: 'Trump Cards' }));
      const nav = screen.getByRole('navigation');
      expect(nav).toHaveClass('hidden');
    });

    it('language toggle is always accessible in mobile header', () => {
      renderNavBar();
      // Language buttons exist even when nav is hidden (mobile header has its own)
      const jaBtns = screen.getAllByRole('button', { name: i18n.t('nav.switchToJa') });
      expect(jaBtns.length).toBeGreaterThanOrEqual(2);
    });

    it('moves focus to first game link when menu opens', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      fireEvent.click(btn);
      const nav = screen.getByRole('navigation');
      const firstLink = nav.querySelector('a');
      expect(document.activeElement).toBe(firstLink);
    });

    it('returns focus to hamburger button when menu closes', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      fireEvent.click(btn);
      const closeBtn = screen.getByRole('button', { name: i18n.t('nav.closeMenu') });
      fireEvent.click(closeBtn);
      expect(document.activeElement).toBe(closeBtn);
    });

    it('closes menu and returns focus to hamburger button on Escape key', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      fireEvent.click(btn);
      const nav = screen.getByRole('navigation');
      fireEvent.keyDown(nav, { key: 'Escape' });
      expect(nav).toHaveClass('hidden');
      expect(document.activeElement).toBe(btn);
    });

    it('does not close menu on non-Escape key', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      fireEvent.click(btn);
      const nav = screen.getByRole('navigation');
      fireEvent.keyDown(nav, { key: 'Tab' });
      expect(nav).not.toHaveClass('hidden');
    });
  });

  describe('SoundToggle', () => {
    it('does not render SoundToggle when props not provided', () => {
      renderNavBar();
      expect(screen.queryByRole('button', { name: 'サウンドをオフにする' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'サウンドをオンにする' })).not.toBeInTheDocument();
    });

    it('renders SoundToggle when soundMuted and onSoundToggle are provided', () => {
      renderNavBar('/', { soundMuted: false, onSoundToggle: vi.fn() });
      expect(screen.getAllByRole('button', { name: 'サウンドをオフにする' }).length).toBeGreaterThan(0);
    });

    it('calls onSoundToggle when SoundToggle is clicked', () => {
      const onSoundToggle = vi.fn();
      renderNavBar('/', { soundMuted: false, onSoundToggle });
      const buttons = screen.getAllByRole('button', { name: 'サウンドをオフにする' });
      fireEvent.click(buttons[0]);
      expect(onSoundToggle).toHaveBeenCalledTimes(1);
    });
  });
});
