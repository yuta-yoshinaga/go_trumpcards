import { fireEvent, render, screen } from '@testing-library/react';
import i18n from 'i18next';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { gameCategories, gameRoutes } from '../constants/gameRoutes';
import { NavBar } from './NavBar';

// **この 2 ファイルは全ゲームぶんのリンクを描いて数える。** ゲームが 1 つ
// 増えるたびに重くなり、313 ゲームの時点で `links point to correct hrefs` は
// ローカルで 2.1 秒 —— 既定の 10 秒に対して 2 割強を使う。CI の runner は
// 負荷時にこの数倍かかるので、既定のままだと**中身と無関係にタイムアウトで
// 落ちる**（実際に NavBar と DesktopSidebar の両方が CI で落ちた）。
//
// テストが遅いのは検査の性質であって欠陥ではない（全ルートを網羅する検査を
// 速くする方法は「検査を減らす」しかない）ので、このファイルだけ上限を上げる。
// hookTimeout も同じ理由で上げる。**片方だけでは足りない** —— 316 ゲームぶんの
// DOM を畳む RTL の自動 cleanup は afterEach フックで走るので、既定の 10 秒だと
// 「Hook timed out」で落ちる（CI で実際に落ちた）。
vi.setConfig({ testTimeout: 30_000, hookTimeout: 30_000 });

vi.mock('../providers/SoundProvider', () => ({
  useSound: vi.fn(() => ({
    muted: false,
    toggleMute: vi.fn(),
    playSound: vi.fn(),
    claimExecSound: vi.fn(),
    consumeExecClaim: vi.fn(() => false),
  })),
}));

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
    it(`marks ${labelKey} link as active when on ${path}`, () => {
      renderNavBar(path);
      // Exactly one link may carry aria-current, which also proves no other
      // route's link has it.
      const current = screen.getAllByRole('link').filter((link) => link.hasAttribute('aria-current'));
      expect(current).toHaveLength(1);
      expect(current[0]).toHaveAttribute('aria-current', 'page');
      expect(current[0]).toHaveTextContent(labelFor(labelKey));
    });
  }

  it('links point to correct hrefs', () => {
    renderNavBar();
    // Index the links once. Querying by accessible name recomputes it for every
    // element in the tree, so doing that per route is quadratic.
    // A path can be rendered by more than one link (category list, sidebar),
    // so collect every label per href rather than keeping the last one.
    const labelsByHref = new Map<string, string[]>();
    for (const link of screen.getAllByRole('link')) {
      const href = link.getAttribute('href') ?? '';
      const labels = labelsByHref.get(href) ?? [];
      labels.push(link.textContent ?? '');
      labelsByHref.set(href, labels);
    }
    for (const { path, labelKey } of gameRoutes) {
      expect(labelsByHref.get(path)?.some((label) => label.includes(labelFor(labelKey)))).toBe(true);
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

    it('moves focus to first interactive element when menu opens', () => {
      renderNavBar();
      const btn = screen.getByRole('button', { name: i18n.t('nav.openMenu') });
      fireEvent.click(btn);
      const nav = screen.getByRole('navigation');
      const firstInteractive = nav.querySelector('input, a');
      expect(document.activeElement).toBe(firstInteractive);
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

    describe('focus trap (mobile)', () => {
      const originalInnerWidth = window.innerWidth;

      beforeEach(() => {
        Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
        window.dispatchEvent(new Event('resize'));
      });

      afterEach(() => {
        Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalInnerWidth });
        window.dispatchEvent(new Event('resize'));
      });

      it('wraps Tab from the last focusable element back to the first', () => {
        renderNavBar();
        fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
        const nav = screen.getByRole('navigation');
        const focusable = nav.querySelectorAll<HTMLElement>('a[href], button, input');
        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        last.focus();
        expect(document.activeElement).toBe(last);
        const tabEvent = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true });
        document.dispatchEvent(tabEvent);
        expect(tabEvent.defaultPrevented).toBe(true);
        expect(document.activeElement).toBe(first);
      });

      it('wraps Shift+Tab from the first focusable element to the last', () => {
        renderNavBar();
        fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
        const nav = screen.getByRole('navigation');
        const focusable = nav.querySelectorAll<HTMLElement>('a[href], button, input');
        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        first.focus();
        const tabEvent = new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true });
        document.dispatchEvent(tabEvent);
        expect(tabEvent.defaultPrevented).toBe(true);
        expect(document.activeElement).toBe(last);
      });

      it('pulls focus back into the nav when Tab is pressed outside', () => {
        renderNavBar();
        fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
        const nav = screen.getByRole('navigation');
        const outside = document.createElement('button');
        document.body.appendChild(outside);
        outside.focus();
        expect(document.activeElement).toBe(outside);
        const tabEvent = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true });
        document.dispatchEvent(tabEvent);
        expect(tabEvent.defaultPrevented).toBe(true);
        const focusable = nav.querySelectorAll<HTMLElement>('a[href], button, input');
        expect(document.activeElement).toBe(focusable[0]);
        document.body.removeChild(outside);
      });

      it('does not reset focus when the viewport resizes while the menu is open', () => {
        renderNavBar();
        fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
        const nav = screen.getByRole('navigation');
        const focusable = nav.querySelectorAll<HTMLElement>('a[href], button, input');
        const mid = focusable[Math.floor(focusable.length / 2)];
        mid.focus();
        expect(document.activeElement).toBe(mid);
        // Cross the mobile breakpoint while the menu is still open — the
        // effect re-runs because isMobile flips, but focus must be preserved.
        Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
        window.dispatchEvent(new Event('resize'));
        expect(document.activeElement).toBe(mid);
      });
    });
  });

  describe('collapsible categories', () => {
    it('expands all categories on mobile viewport', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar('/poker');
      const categoryDetails = document.querySelectorAll('.nav-category');
      expect(categoryDetails.length).toBe(gameCategories.length);
      for (const details of categoryDetails) {
        expect(details).toHaveAttribute('open');
      }
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('forces details open on mobile when toggled closed', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar('/poker');
      const pokerDetails = screen.getByText(labelFor('nav.category.poker')).closest('details') as HTMLDetailsElement;
      expect(pokerDetails).toHaveAttribute('open');
      // Simulate toggle to close — onToggle handler should force it back open
      pokerDetails.open = false;
      fireEvent(pokerDetails, new Event('toggle'));
      expect(pokerDetails).toHaveAttribute('open');
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('auto-opens details for the active category', () => {
      renderNavBar('/poker');
      const pokerDetails = screen.getByText(labelFor('nav.category.poker')).closest('details');
      expect(pokerDetails).toHaveAttribute('open');
    });

    it('auto-opens home category when on root path', () => {
      renderNavBar('/');
      const tableDetails = screen.getByText(labelFor('nav.category.table')).closest('details');
      expect(tableDetails).toHaveAttribute('open');
    });

    it('forces details open on medium desktop (sm-lg) when toggled closed', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar('/');
      const tableDetails = screen.getByText(labelFor('nav.category.table')).closest('details') as HTMLDetailsElement;
      expect(tableDetails).toHaveAttribute('open');

      // Simulate toggle to close — the onToggle handler should force it back open
      tableDetails.open = false;
      fireEvent(tableDetails, new Event('toggle'));
      expect(tableDetails).toHaveAttribute('open');

      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('does not close other categories on mousedown inside a category on mobile', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar('/');
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));

      const tableDetails = screen.getByText(labelFor('nav.category.table')).closest('details') as HTMLDetailsElement;
      const pokerDetails = screen.getByText(labelFor('nav.category.poker')).closest('details') as HTMLDetailsElement;
      expect(tableDetails).toHaveAttribute('open');
      expect(pokerDetails).toHaveAttribute('open');

      // mousedown on a Poker link — Table category must stay open to prevent layout shift
      const pokerLink = pokerDetails.querySelector('a') as HTMLElement;
      fireEvent.mouseDown(pokerLink);

      expect(tableDetails).toHaveAttribute('open');
      expect(pokerDetails).toHaveAttribute('open');
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('does not close other categories on mousedown inside a category on medium desktop', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar('/');

      const tableDetails = screen.getByText(labelFor('nav.category.table')).closest('details') as HTMLDetailsElement;
      const pokerDetails = screen.getByText(labelFor('nav.category.poker')).closest('details') as HTMLDetailsElement;
      // Force both open to simulate tablet state
      tableDetails.open = true;
      pokerDetails.open = true;

      const pokerLink = pokerDetails.querySelector('a') as HTMLElement;
      fireEvent.mouseDown(pokerLink);

      expect(tableDetails).toHaveAttribute('open');
      expect(pokerDetails).toHaveAttribute('open');
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('does not close dropdown when clicking inside it', () => {
      renderNavBar('/poker');
      const pokerDetails = screen.getByText(labelFor('nav.category.poker')).closest('details') as HTMLDetailsElement;
      expect(pokerDetails).toHaveAttribute('open');

      // Click inside the open dropdown — should stay open
      const pokerLink = pokerDetails.querySelector('a') as HTMLElement;
      fireEvent.mouseDown(pokerLink);

      expect(pokerDetails).toHaveAttribute('open');
    });

    it('removes outside click listener on unmount', () => {
      vi.spyOn(document, 'removeEventListener');
      const { unmount } = renderNavBar();
      unmount();
      expect(document.removeEventListener).toHaveBeenCalledWith('mousedown', expect.any(Function));
      vi.restoreAllMocks();
    });
  });

  describe('favorites', () => {
    it('shows favorite star buttons on mobile', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const starButtons = screen.getAllByRole('button', { name: i18n.t('nav.favoriteGames') });
      expect(starButtons.length).toBeGreaterThan(0);
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('hides favorite star buttons on desktop', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      expect(screen.queryByRole('button', { name: i18n.t('nav.favoriteGames') })).not.toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('toggling star adds game to favorites section', () => {
      localStorage.removeItem('trumpcards-favorite-games');
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      // No favorites section initially
      expect(screen.queryByText(i18n.t('nav.favoriteGames'))).not.toBeInTheDocument();
      // Click the first star button
      const starButtons = screen.getAllByRole('button', { name: i18n.t('nav.favoriteGames') });
      fireEvent.click(starButtons[0]);
      // Favorites section should appear
      expect(screen.getByText(i18n.t('nav.favoriteGames'))).toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
      localStorage.removeItem('trumpcards-favorite-games');
    });

    it('shows favorites section when favorites exist on mobile', () => {
      localStorage.setItem('trumpcards-favorite-games', JSON.stringify(['/poker']));
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      expect(screen.getByText(i18n.t('nav.favoriteGames'))).toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
      localStorage.removeItem('trumpcards-favorite-games');
    });

    it('hides favorites section on desktop', () => {
      localStorage.setItem('trumpcards-favorite-games', JSON.stringify(['/poker']));
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      expect(screen.queryByText(i18n.t('nav.favoriteGames'))).not.toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
      localStorage.removeItem('trumpcards-favorite-games');
    });

    it('exposes aria-pressed on the favorite toggle and flips color classes on toggle', () => {
      localStorage.removeItem('trumpcards-favorite-games');
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const [firstStar] = screen.getAllByRole('button', { name: i18n.t('nav.favoriteGames') });
      expect(firstStar).toHaveAttribute('aria-pressed', 'false');
      expect(firstStar.className).toContain('text-ds-text-muted');
      fireEvent.click(firstStar);
      const toggled = screen.getAllByRole('button', { name: i18n.t('nav.favoriteGames') })[0];
      expect(toggled).toHaveAttribute('aria-pressed', 'true');
      expect(toggled.className).toContain('text-ds-accent');
      expect(toggled.className).not.toContain('text-ds-text-muted');
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
      localStorage.removeItem('trumpcards-favorite-games');
    });
  });

  describe('recently played', () => {
    it('does not show section when no recent games', () => {
      localStorage.removeItem('trumpcards-recent-games');
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      // Use a non-game path so useRecentGames doesn't record anything
      renderNavBar('/nonexistent');
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      expect(screen.queryByText(i18n.t('nav.recentGames'))).not.toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('shows recently played section with stored games on mobile', () => {
      localStorage.setItem('trumpcards-recent-games', JSON.stringify(['/poker', '/hearts']));
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      expect(screen.getByText(i18n.t('nav.recentGames'))).toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
      localStorage.removeItem('trumpcards-recent-games');
    });

    it('hides recently played section on desktop', () => {
      localStorage.setItem('trumpcards-recent-games', JSON.stringify(['/poker']));
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      expect(screen.queryByText(i18n.t('nav.recentGames'))).not.toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
      localStorage.removeItem('trumpcards-recent-games');
    });

    it('hides recently played section when search is active', () => {
      localStorage.setItem('trumpcards-recent-games', JSON.stringify(['/poker']));
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'ブラック' } });
      expect(screen.queryByText(i18n.t('nav.recentGames'))).not.toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
      localStorage.removeItem('trumpcards-recent-games');
    });
  });

  describe('game search', () => {
    it('shows search input on mobile when menu is open', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      expect(screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'))).toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('hides search input on desktop', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1280 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      expect(screen.queryByPlaceholderText(i18n.t('nav.searchPlaceholder'))).not.toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('filters games by name when typing', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'ブラック' } });
      // Should show blackjack, hide poker
      expect(screen.getByRole('link', { name: labelFor('nav.blackjack') })).toBeInTheDocument();
      expect(screen.queryByRole('link', { name: labelFor('nav.poker') })).not.toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('matches English game names when UI is Japanese', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'Blackjack' } });
      expect(screen.getByRole('link', { name: labelFor('nav.blackjack') })).toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('restores categories when search is cleared', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'ブラック' } });
      // Categories should be hidden during search
      expect(screen.queryByText(labelFor('nav.category.poker'))).not.toBeInTheDocument();
      // Clear search
      fireEvent.change(input, { target: { value: '' } });
      // Categories should reappear
      expect(screen.getByText(labelFor('nav.category.poker'))).toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('search input has aria-label and type="search"', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const input = screen.getByRole('searchbox');
      expect(input).toHaveAttribute('aria-label', i18n.t('nav.searchPlaceholder'));
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('shows clear button when search has text', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      // No clear button initially
      expect(screen.queryByRole('button', { name: i18n.t('nav.searchClear') })).not.toBeInTheDocument();
      fireEvent.change(input, { target: { value: 'test' } });
      // Clear button appears
      const clearBtn = screen.getByRole('button', { name: i18n.t('nav.searchClear') });
      expect(clearBtn).toBeInTheDocument();
      // Clicking clear restores categories
      fireEvent.click(clearBtn);
      expect(screen.getByText(labelFor('nav.category.poker'))).toBeInTheDocument();
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('shows no results message when search matches nothing', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'xyznonexistent' } });
      const noResultsElements = screen.getAllByText(i18n.t('nav.noResults'));
      expect(noResultsElements.length).toBeGreaterThanOrEqual(1);
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });

    it('clears search when a game link is clicked', () => {
      const original = window.innerWidth;
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      renderNavBar();
      fireEvent.click(screen.getByRole('button', { name: i18n.t('nav.openMenu') }));
      const input = screen.getByPlaceholderText(i18n.t('nav.searchPlaceholder'));
      fireEvent.change(input, { target: { value: 'ブラック' } });
      fireEvent.click(screen.getByRole('link', { name: labelFor('nav.blackjack') }));
      // Menu should be closed (search cleared implicitly)
      const nav = screen.getByRole('navigation');
      expect(nav).toHaveClass('hidden');
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    });
  });

  describe('SoundToggle', () => {
    it('renders SoundToggle from context', () => {
      renderNavBar();
      expect(screen.getAllByRole('button', { name: 'サウンドをオフにする' }).length).toBeGreaterThan(0);
    });
  });
});
