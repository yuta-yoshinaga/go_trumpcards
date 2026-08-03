import { screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it } from 'vitest';
import { DesktopSidebar } from '../components/DesktopSidebar';
import { NavBar } from '../components/NavBar';
import { renderWithProviders } from '../test/renderWithProviders';
import { LegalPage } from './LegalPage';

function renderPage() {
  return renderWithProviders(
    <MemoryRouter initialEntries={['/legal']}>
      <LegalPage />
    </MemoryRouter>,
  );
}

describe('LegalPage', () => {
  it('states that the project is unaffiliated with the trademark owners', () => {
    renderPage();
    // The non-affiliation sentence is the whole point of the page — a rights
    // holder's complaint is that users believe the game is licensed. Assert on
    // the substance, not merely that some heading rendered.
    expect(screen.getByText(/一切関係がなく/)).toBeInTheDocument();
    expect(screen.getByText(/承認・後援・提携を受けてもいません/)).toBeInTheDocument();
  });

  it('explains that rules are implemented but no publisher material is reproduced', () => {
    renderPage();
    expect(screen.getByText(/ルール自体は著作権の保護対象ではありません/)).toBeInTheDocument();
    expect(
      screen.getByText(/アートワーク、ロゴ、書体、パッケージ、トレードドレスは一切使用していません/),
    ).toBeInTheDocument();
  });

  it('credits the card art and sound assets with their licenses', () => {
    renderPage();
    expect(screen.getByText(/Dmitry Fomin/)).toBeInTheDocument();
    expect(screen.getByText(/Byron Knoll/)).toBeInTheDocument();
    expect(screen.getByText(/効果音はすべて本プロジェクトで手続き的に生成/)).toBeInTheDocument();
  });

  it('links to the full trademark inventory, the licence and the issue tracker', () => {
    renderPage();
    const inventory = screen.getByRole('link', { name: /TRADEMARKS\.md/ });
    expect(inventory).toHaveAttribute('href', expect.stringContaining('TRADEMARKS.md'));
    expect(inventory).toHaveAttribute('rel', 'noopener noreferrer');
    expect(screen.getByRole('link', { name: /LICENSE/ })).toHaveAttribute(
      'href',
      expect.stringContaining('/blob/develop/LICENSE'),
    );
    expect(screen.getByRole('link', { name: /GitHub Issues/ })).toHaveAttribute(
      'href',
      expect.stringContaining('/issues'),
    );
  });

  it('offers a way back to the game', () => {
    renderPage();
    expect(screen.getByRole('link', { name: 'ホームに戻る' })).toHaveAttribute('href', '/');
  });
});

// The notice is worthless if it cannot be reached, and the two pieces of
// persistent chrome cover different widths: NavBar is `lg:hidden`, DesktopSidebar
// is `hidden lg:flex`. An earlier revision gated the NavBar link on `isMobile`
// (< 640px), which left every width from 640px to 1023px — tablets and small
// desktops — with no route to the page at all.
//
// This cannot be asserted as "exactly one link in the DOM": jsdom loads no
// Tailwind stylesheet, so `lg:hidden` and `hidden lg:flex` do nothing here and
// both components always render. What is testable is the JS condition that
// decides whether NavBar contributes the link at all, which is where the bug
// was, so that is what these pin.
describe('legal notice reachability across breakpoints', () => {
  const LINK_NAME = '商標とクレジット';
  const originalWidth = window.innerWidth;

  afterEach(() => {
    Object.defineProperty(window, 'innerWidth', { value: originalWidth, configurable: true });
  });

  function renderNavBarAt(width: number) {
    Object.defineProperty(window, 'innerWidth', { value: width, configurable: true });
    renderWithProviders(
      <MemoryRouter initialEntries={['/']}>
        <NavBar />
      </MemoryRouter>,
    );
  }

  it.each([
    ['mobile', 375],
    ['tablet / small desktop', 800],
  ])('NavBar carries the link at %s width, where the sidebar is hidden', (_label, width) => {
    renderNavBarAt(width);
    expect(screen.getByRole('link', { name: LINK_NAME })).toHaveAttribute('href', '/legal');
  });

  it('NavBar drops the link at large-desktop width, where the sidebar takes over', () => {
    renderNavBarAt(1280);
    expect(screen.queryByRole('link', { name: LINK_NAME })).not.toBeInTheDocument();
  });

  it('DesktopSidebar carries the link, covering large-desktop width', () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/']}>
        <DesktopSidebar />
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: LINK_NAME })).toHaveAttribute('href', '/legal');
  });

  // Both links compute `aria-current` from the route, and the tests above only
  // ever render away from /legal — so only the `undefined` branch was taken and
  // the "you are here" state was never asserted at all. Screen-reader users are
  // the ones who lose if it regresses.
  it('NavBar marks its link as the current page while on /legal', () => {
    Object.defineProperty(window, 'innerWidth', { value: 375, configurable: true });
    renderWithProviders(
      <MemoryRouter initialEntries={['/legal']}>
        <NavBar />
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: LINK_NAME })).toHaveAttribute('aria-current', 'page');
  });

  it('DesktopSidebar marks its link as the current page while on /legal', () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/legal']}>
        <DesktopSidebar />
      </MemoryRouter>,
    );
    expect(screen.getByRole('link', { name: LINK_NAME })).toHaveAttribute('aria-current', 'page');
  });
});
