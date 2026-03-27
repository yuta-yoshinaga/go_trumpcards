import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { gameCategories, gameRoutes } from '../constants/gameRoutes';
import { DesktopSidebar } from './DesktopSidebar';

function renderSidebar(initialPath = '/', props?: { soundMuted?: boolean; onSoundToggle?: () => void }) {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <DesktopSidebar {...props} />
    </MemoryRouter>,
  );
}

function labelFor(labelKey: string): string {
  return i18n.t(labelKey);
}

afterEach(() => {
  i18n.changeLanguage('ja');
  localStorage.removeItem('trumpcards-favorite-games');
});

describe('DesktopSidebar', () => {
  it('renders navigation landmark with aria-label', () => {
    renderSidebar();
    const sidebar = screen.getByRole('complementary');
    expect(sidebar).toBeInTheDocument();
  });

  it('renders site name link pointing to home', () => {
    renderSidebar();
    const brandLink = screen.getByRole('link', { name: 'Trump Cards' });
    expect(brandLink).toHaveAttribute('href', '/');
  });

  it('renders all game links', () => {
    renderSidebar();
    for (const { labelKey } of gameRoutes) {
      expect(screen.getByRole('link', { name: labelFor(labelKey) })).toBeInTheDocument();
    }
  });

  it('renders all category labels', () => {
    renderSidebar();
    for (const { labelKey } of gameCategories) {
      expect(screen.getByText(labelFor(labelKey))).toBeInTheDocument();
    }
  });

  it('marks active game with aria-current="page"', () => {
    renderSidebar('/poker');
    expect(screen.getByRole('link', { name: labelFor('nav.poker') })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: labelFor('nav.hearts') })).not.toHaveAttribute('aria-current');
  });

  it('links point to correct hrefs', () => {
    renderSidebar();
    for (const { path, labelKey } of gameRoutes) {
      expect(screen.getByRole('link', { name: labelFor(labelKey) })).toHaveAttribute('href', path);
    }
  });

  describe('search', () => {
    it('renders search input', () => {
      renderSidebar();
      expect(screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'))).toBeInTheDocument();
    });

    it('filters games by Japanese name', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'ブラック' } });
      expect(screen.getByRole('link', { name: labelFor('nav.blackjack') })).toBeInTheDocument();
      expect(screen.queryByRole('link', { name: labelFor('nav.poker') })).not.toBeInTheDocument();
    });

    it('filters games by English name', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'Blackjack' } });
      expect(screen.getByRole('link', { name: labelFor('nav.blackjack') })).toBeInTheDocument();
    });

    it('shows no results message', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'xyznonexistent' } });
      expect(screen.getByText(i18n.t('nav.noResults'))).toBeInTheDocument();
    });

    it('hides categories during search and restores on clear', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'ブラック' } });
      expect(screen.queryByText(labelFor('nav.category.poker'))).not.toBeInTheDocument();
      fireEvent.change(input, { target: { value: '' } });
      expect(screen.getByText(labelFor('nav.category.poker'))).toBeInTheDocument();
    });

    it('shows clear button and clears on click', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      expect(screen.queryByRole('button', { name: i18n.t('nav.searchClear') })).not.toBeInTheDocument();
      fireEvent.change(input, { target: { value: 'test' } });
      const clearBtn = screen.getByRole('button', { name: i18n.t('nav.searchClear') });
      fireEvent.click(clearBtn);
      expect(screen.getByText(labelFor('nav.category.poker'))).toBeInTheDocument();
    });

    it('clears search on Escape key', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'test' } });
      fireEvent.keyDown(input, { key: 'Escape' });
      expect(screen.getByText(labelFor('nav.category.poker'))).toBeInTheDocument();
    });

    it('does not clear search on non-Escape key', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'ブラック' } });
      fireEvent.keyDown(input, { key: 'Tab' });
      expect(screen.queryByText(labelFor('nav.category.poker'))).not.toBeInTheDocument();
    });
  });

  describe('favorites', () => {
    it('renders favorite toggle buttons for all games', () => {
      renderSidebar();
      const starButtons = screen.getAllByRole('button', { name: i18n.t('nav.addFavorite') });
      expect(starButtons.length).toBe(gameRoutes.length);
    });

    it('toggling star adds game to favorites section', () => {
      renderSidebar();
      expect(screen.queryByText(i18n.t('nav.favoriteGames'))).not.toBeInTheDocument();
      const starButtons = screen.getAllByRole('button', { name: i18n.t('nav.addFavorite') });
      fireEvent.click(starButtons[0]);
      expect(screen.getByText(i18n.t('nav.favoriteGames'))).toBeInTheDocument();
    });

    it('shows favorites section when favorites exist', () => {
      localStorage.setItem('trumpcards-favorite-games', JSON.stringify(['/poker']));
      renderSidebar();
      expect(screen.getByText(i18n.t('nav.favoriteGames'))).toBeInTheDocument();
    });

    it('hides favorites section during search', () => {
      localStorage.setItem('trumpcards-favorite-games', JSON.stringify(['/poker']));
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'test' } });
      expect(screen.queryByText(i18n.t('nav.favoriteGames'))).not.toBeInTheDocument();
    });
  });

  describe('language toggle', () => {
    it('renders JA and EN buttons with aria-pressed', () => {
      renderSidebar();
      const jaBtn = screen.getByRole('button', { name: i18n.t('nav.switchToJa') });
      const enBtn = screen.getByRole('button', { name: i18n.t('nav.switchToEn') });
      expect(jaBtn).toHaveAttribute('aria-pressed', 'true');
      expect(enBtn).toHaveAttribute('aria-pressed', 'false');
    });

    it('switches language to EN', () => {
      renderSidebar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.switchToEn') }));
      expect(i18n.language).toBe('en');
    });
  });

  describe('SoundToggle', () => {
    it('does not render when props not provided', () => {
      renderSidebar();
      expect(screen.queryByRole('button', { name: i18n.t('sound.mute') })).not.toBeInTheDocument();
    });

    it('renders when soundMuted and onSoundToggle are provided', () => {
      renderSidebar('/', { soundMuted: false, onSoundToggle: vi.fn() });
      expect(screen.getByRole('button', { name: i18n.t('sound.mute') })).toBeInTheDocument();
    });

    it('calls onSoundToggle when clicked', () => {
      const onSoundToggle = vi.fn();
      renderSidebar('/', { soundMuted: false, onSoundToggle });
      fireEvent.click(screen.getByRole('button', { name: i18n.t('sound.mute') }));
      expect(onSoundToggle).toHaveBeenCalledTimes(1);
    });
  });

  describe('tutorial progress', () => {
    it('renders tutorial progress panel', () => {
      renderSidebar();
      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
  });
});
