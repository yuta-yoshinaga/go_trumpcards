import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { gameCategories, gameRoutes } from '../constants/gameRoutes';
import { DesktopSidebar } from './DesktopSidebar';

// **この 2 ファイルは全ゲームぶんのリンクを描いて数える。** ゲームが 1 つ
// 増えるたびに重くなり、313 ゲームの時点で `links point to correct hrefs` は
// ローカルで 2.1 秒 —— 既定の 10 秒に対して 2 割強を使う。CI の runner は
// 負荷時にこの数倍かかるので、既定のままだと**中身と無関係にタイムアウトで
// 落ちる**（実際に NavBar と DesktopSidebar の両方が CI で落ちた）。
//
// テストが遅いのは検査の性質であって欠陥ではない（全ルートを網羅する検査を
// 速くする方法は「検査を減らす」しかない）ので、このファイルだけ上限を上げる。
vi.setConfig({ testTimeout: 30_000 });

vi.mock('../providers/SoundProvider', () => ({
  useSound: vi.fn(() => ({
    muted: false,
    toggleMute: vi.fn(),
    playSound: vi.fn(),
    claimExecSound: vi.fn(),
    consumeExecClaim: vi.fn(() => false),
  })),
}));

// Default to a non-game path so useRecentGames does not record the current page
// and produce a duplicate "Recent" link alongside the categorized one. Tests
// that need a specific game path (or `/`) pass it explicitly.
function renderSidebar(initialPath = '/nonexistent') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <DesktopSidebar />
    </MemoryRouter>,
  );
}

function labelFor(labelKey: string): string {
  return i18n.t(labelKey);
}

afterEach(() => {
  i18n.changeLanguage('ja');
  localStorage.removeItem('trumpcards-favorite-games');
  localStorage.removeItem('trumpcards-recent-games');
  // Accordion expansion state bleeds across tests otherwise — the
  // "toggles category open/closed on click" test writes a collapse, and
  // the "expands every category by default" assertions then fail.
  localStorage.removeItem('trumpcards-category-expansion');
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

  it('renders the AI Game Concierge CTA pointing to /discover', () => {
    renderSidebar();
    const cta = screen.getByRole('link', { name: /おすすめを探す/ });
    expect(cta).toHaveAttribute('href', '/discover');
  });

  it('marks the AI Game Concierge CTA as current page when on /discover', () => {
    renderSidebar('/discover');
    const cta = screen.getByRole('link', { name: /おすすめを探す/ });
    expect(cta).toHaveAttribute('aria-current', 'page');
  });

  it('marks the AI Game Concierge CTA as current page on /discover/result too', () => {
    renderSidebar('/discover/result?m=0,0&s=0,0&so=0,0&t=0,0');
    const cta = screen.getByRole('link', { name: /おすすめを探す/ });
    expect(cta).toHaveAttribute('aria-current', 'page');
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
    // /poker also gets recorded in recent games, producing a duplicate link.
    const pokerLinks = screen.getAllByRole('link', { name: labelFor('nav.poker') });
    for (const link of pokerLinks) {
      expect(link).toHaveAttribute('aria-current', 'page');
    }
    expect(screen.getByRole('link', { name: labelFor('nav.hearts') })).not.toHaveAttribute('aria-current');
  });

  it('links point to correct hrefs', () => {
    renderSidebar();
    for (const { path, labelKey } of gameRoutes) {
      expect(screen.getByRole('link', { name: labelFor(labelKey) })).toHaveAttribute('href', path);
    }
  });

  describe('recent games', () => {
    it('does not render section when no recent games stored', () => {
      renderSidebar();
      expect(screen.queryByText(i18n.t('nav.recentGames'))).not.toBeInTheDocument();
    });

    it('renders section when recent games are stored', () => {
      localStorage.setItem('trumpcards-recent-games', JSON.stringify(['/poker', '/hearts']));
      renderSidebar();
      expect(screen.getByText(i18n.t('nav.recentGames'))).toBeInTheDocument();
    });

    it('hides section during search', () => {
      localStorage.setItem('trumpcards-recent-games', JSON.stringify(['/poker']));
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'test' } });
      expect(screen.queryByText(i18n.t('nav.recentGames'))).not.toBeInTheDocument();
    });
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
      const noResultsElements = screen.getAllByText(i18n.t('nav.noResults'));
      expect(noResultsElements.length).toBeGreaterThanOrEqual(1);
    });

    it('hides categories during search and restores on clear', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'ブラック' } });
      expect(screen.queryByText(labelFor('nav.category.poker'))).not.toBeInTheDocument();
      fireEvent.change(input, { target: { value: '' } });
      expect(screen.getByText(labelFor('nav.category.poker'))).toBeInTheDocument();
    });

    it('applies min-w-[44px] and min-h-[44px] touch target to clear button', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'test' } });
      const clearBtn = screen.getByRole('button', { name: i18n.t('nav.searchClear') });
      expect(clearBtn.className).toContain('min-w-[44px]');
      expect(clearBtn.className).toContain('min-h-[44px]');
    });

    it('applies focus ring to clear button', () => {
      renderSidebar();
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'test' } });
      const clearBtn = screen.getByRole('button', { name: i18n.t('nav.searchClear') });
      expect(clearBtn.className).toContain('focus-visible:ring-2');
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
      const starButtons = screen.getAllByRole('button', { name: i18n.t('nav.favoriteGames') });
      expect(starButtons.length).toBe(gameRoutes.length);
    });

    it('applies min-w-[44px] and min-h-[44px] touch target to favorite buttons', () => {
      renderSidebar();
      const starButtons = screen.getAllByRole('button', { name: i18n.t('nav.favoriteGames') });
      expect(starButtons[0].className).toContain('min-w-[44px]');
      expect(starButtons[0].className).toContain('min-h-[44px]');
    });

    it('applies focus ring to favorite buttons', () => {
      renderSidebar();
      const starButtons = screen.getAllByRole('button', { name: i18n.t('nav.favoriteGames') });
      expect(starButtons[0].className).toContain('focus-visible:ring-2');
    });

    it('toggling star adds game to favorites section', () => {
      renderSidebar();
      expect(screen.queryByText(i18n.t('nav.favoriteGames'))).not.toBeInTheDocument();
      const starButtons = screen.getAllByRole('button', { name: i18n.t('nav.favoriteGames') });
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
    it('renders SoundToggle from context', () => {
      renderSidebar();
      expect(screen.getByRole('button', { name: i18n.t('sound.mute') })).toBeInTheDocument();
    });
  });

  describe('accordion categories', () => {
    it('renders categories as details elements', () => {
      renderSidebar();
      for (const { labelKey } of gameCategories) {
        const summary = screen.getByText(labelFor(labelKey));
        expect(summary.closest('details')).toBeInTheDocument();
      }
    });

    it('expands every category by default on first visit', () => {
      // #1698: every category must be discoverable on first load so the
      // 77-game catalog is visible without 5 extra clicks. Across all 6
      // categories, every <details> should start with the `open` attribute.
      renderSidebar('/poker');
      const allCategories = ['nav.category.table', 'nav.category.poker', 'nav.category.solitaire'];
      for (const labelKey of allCategories) {
        const heading = screen.getByText(labelFor(labelKey));
        expect(heading.closest('details')).toHaveAttribute('open');
      }
    });

    it('expands every category by default on the home page', () => {
      renderSidebar('/');
      for (const labelKey of ['nav.category.table', 'nav.category.poker', 'nav.category.solitaire']) {
        const heading = screen.getByText(labelFor(labelKey));
        expect(heading.closest('details')).toHaveAttribute('open');
      }
    });

    it('toggles category open/closed on click', () => {
      renderSidebar('/poker');
      const tableCategory = screen.getByText(labelFor('nav.category.table'));
      expect(tableCategory.closest('details')).toHaveAttribute('open');
      fireEvent.click(tableCategory);
      expect(tableCategory.closest('details')).not.toHaveAttribute('open');
      fireEvent.click(tableCategory);
      expect(tableCategory.closest('details')).toHaveAttribute('open');
    });
  });

  describe('search icon', () => {
    it('renders a search icon in the search area', () => {
      renderSidebar();
      expect(screen.getByTestId('search-icon')).toBeInTheDocument();
    });
  });

  describe('tutorial progress', () => {
    it('renders tutorial progress panel', () => {
      renderSidebar();
      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
  });
});
