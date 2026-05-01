import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { gameTheme } from '../../styles/gameTheme';
import { GameSkeleton } from './GameSkeleton';

describe('GameSkeleton (layout-driven)', () => {
  describe('theme resolution', () => {
    it('resolves bgClass from gameTheme[gameKey]', () => {
      render(<GameSkeleton gameKey="hearts" layout={{ kind: 'trick-taking', footerHandSize: 5 }} />);
      expect(screen.getByTestId('skeleton').className).toContain(gameTheme.hearts.bg);
    });

    it('resolves footer base classes from gameTheme[gameKey]', () => {
      const { container } = render(
        <GameSkeleton gameKey="blackjack" layout={{ kind: 'casino-table', sections: [5], footerStyle: 'hand' }} />,
      );
      const footer = container.querySelector('footer');
      expect(footer).not.toBeNull();
      expect(footer?.className).toContain(gameTheme.blackjack.footer);
    });

    it('renders skeleton shell with role=status', () => {
      render(<GameSkeleton gameKey="war" layout={{ kind: 'centered', rows: [2] }} />);
      const el = screen.getByTestId('skeleton');
      expect(el.getAttribute('role')).toBe('status');
    });
  });

  describe('casino-table layout', () => {
    it('renders one centered hand section per entry in sections[]', () => {
      const { container } = render(
        <GameSkeleton gameKey="baccarat" layout={{ kind: 'casino-table', sections: [2, 2] }} />,
      );
      const sections = container.querySelectorAll('[data-skeleton-section="casino-table-section"]');
      expect(sections).toHaveLength(2);
    });

    it('renders 5 cards in a section that requests 5', () => {
      const { container } = render(
        <GameSkeleton gameKey="caribbeanstud" layout={{ kind: 'casino-table', sections: [5, 5] }} />,
      );
      const firstSection = container.querySelector('[data-skeleton-section="casino-table-section"]');
      const cards = firstSection?.querySelectorAll('.animate-pulse');
      expect(cards?.length).toBeGreaterThanOrEqual(5);
    });

    it('renders bet-style footer (two stacked bars) when footerStyle is omitted', () => {
      const { container } = render(
        <GameSkeleton gameKey="baccarat" layout={{ kind: 'casino-table', sections: [2, 2] }} />,
      );
      const footer = container.querySelector('footer');
      const bars = footer?.querySelectorAll('.animate-pulse');
      expect(bars?.length).toBe(2);
    });

    it('renders hand-style footer when footerStyle="hand"', () => {
      const { container } = render(
        <GameSkeleton
          gameKey="blackjack"
          layout={{ kind: 'casino-table', sections: [5], footerStyle: 'hand', footerHandSize: 5 }}
        />,
      );
      const footer = container.querySelector('footer');
      expect(footer?.querySelector('[data-skeleton-section="footer-hand"]')).not.toBeNull();
    });
  });

  describe('community-poker layout', () => {
    it('renders a community-card row when community is set', () => {
      const { container } = render(
        <GameSkeleton
          gameKey="holdem"
          layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 2, footerHandSize: 2 }}
        />,
      );
      expect(container.querySelector('[data-skeleton-section="community"]')).not.toBeNull();
    });

    it('omits the community row when community is undefined', () => {
      const { container } = render(
        <GameSkeleton
          gameKey="poker"
          layout={{ kind: 'community-poker', opponents: 3, opponentCards: 5, footerHandSize: 5 }}
        />,
      );
      expect(container.querySelector('[data-skeleton-section="community"]')).toBeNull();
    });

    it('renders one opponent box per opponents count', () => {
      const { container } = render(
        <GameSkeleton
          gameKey="omaha"
          layout={{ kind: 'community-poker', community: 5, opponents: 3, opponentCards: 4, footerHandSize: 4 }}
        />,
      );
      const opponents = container.querySelectorAll('[data-skeleton-section="opponent"]');
      expect(opponents).toHaveLength(3);
    });
  });

  describe('trick-taking layout', () => {
    it('renders default 3 opponent rows as bars', () => {
      const { container } = render(
        <GameSkeleton gameKey="hearts" layout={{ kind: 'trick-taking', footerHandSize: 5 }} />,
      );
      const opponents = container.querySelectorAll('[data-skeleton-section="opponent"]');
      expect(opponents).toHaveLength(3);
    });

    it('renders specified number of opponents', () => {
      const { container } = render(
        <GameSkeleton gameKey="napoleon" layout={{ kind: 'trick-taking', opponents: 4, footerHandSize: 5 }} />,
      );
      const opponents = container.querySelectorAll('[data-skeleton-section="opponent"]');
      expect(opponents).toHaveLength(4);
    });

    it('renders title bar by default', () => {
      const { container } = render(
        <GameSkeleton gameKey="hearts" layout={{ kind: 'trick-taking', footerHandSize: 5 }} />,
      );
      expect(container.querySelector('[data-skeleton-section="title-bar"]')).not.toBeNull();
    });

    it('omits title bar when titleBar=false', () => {
      const { container } = render(
        <GameSkeleton
          gameKey="oldmaid"
          layout={{ kind: 'trick-taking', titleBar: false, opponents: 0, footerHandSize: 5 }}
        />,
      );
      expect(container.querySelector('[data-skeleton-section="title-bar"]')).toBeNull();
    });

    it('renders standard footer button (w-32) by default', () => {
      const { container } = render(
        <GameSkeleton gameKey="hearts" layout={{ kind: 'trick-taking', footerHandSize: 5 }} />,
      );
      const footer = container.querySelector('footer');
      expect(footer?.querySelector('.w-32')).not.toBeNull();
      expect(footer?.querySelector('.w-48')).toBeNull();
    });

    it('renders wide footer button (w-48 mx-auto) when footerButton="wide"', () => {
      const { container } = render(
        <GameSkeleton gameKey="daifugo" layout={{ kind: 'trick-taking', footerHandSize: 5, footerButton: 'wide' }} />,
      );
      const footer = container.querySelector('footer');
      const wideBtn = footer?.querySelector('.w-48');
      expect(wideBtn).not.toBeNull();
      expect(wideBtn?.className).toContain('mx-auto');
    });

    it('renders trick area when trickArea=true', () => {
      const { container } = render(
        <GameSkeleton gameKey="hearts" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />,
      );
      expect(container.querySelector('[data-skeleton-section="trick-area"]')).not.toBeNull();
    });

    it('renders center card when centerCard=true', () => {
      const { container } = render(
        <GameSkeleton gameKey="crazyeights" layout={{ kind: 'trick-taking', centerCard: true, footerHandSize: 5 }} />,
      );
      expect(container.querySelector('[data-skeleton-section="center-card"]')).not.toBeNull();
    });

    it('renders opponents as hand-style mini hands when opponentStyle="hand"', () => {
      const { container } = render(
        <GameSkeleton
          gameKey="oldmaid"
          layout={{ kind: 'trick-taking', opponentStyle: 'hand', opponentHandSize: 3, footerHandSize: 5 }}
        />,
      );
      const opponents = container.querySelectorAll('[data-skeleton-section="opponent"]');
      // Each hand-style opponent contains its own SkeletonHand cards.
      const cards = opponents[0]?.querySelectorAll('.animate-pulse');
      expect(cards?.length).toBeGreaterThanOrEqual(3);
    });
  });

  describe('tableau layout', () => {
    it('renders top row + tableau columns', () => {
      const { container } = render(
        <GameSkeleton gameKey="klondike" layout={{ kind: 'tableau', topRow: 6, tableau: 7 }} />,
      );
      const top = container.querySelector('[data-skeleton-section="tableau-top"]');
      const bottom = container.querySelector('[data-skeleton-section="tableau-cols"]');
      expect(top).not.toBeNull();
      expect(bottom).not.toBeNull();
      expect(top?.querySelectorAll('.animate-pulse')).toHaveLength(6);
      expect(bottom?.querySelectorAll('.animate-pulse')).toHaveLength(7);
    });
  });

  describe('tiered-rows layout', () => {
    it('renders one row per rows[] entry with the specified number of cards', () => {
      const { container } = render(
        <GameSkeleton gameKey="pyramid" layout={{ kind: 'tiered-rows', rows: [1, 2, 3, 4], stockWaste: true }} />,
      );
      const rows = container.querySelectorAll('[data-skeleton-section="tiered-row"]');
      expect(rows).toHaveLength(4);
      expect(rows[0]?.querySelectorAll('.animate-pulse')).toHaveLength(1);
      expect(rows[3]?.querySelectorAll('.animate-pulse')).toHaveLength(4);
    });

    it('renders stock+waste row when stockWaste=true', () => {
      const { container } = render(
        <GameSkeleton gameKey="tripeaks" layout={{ kind: 'tiered-rows', rows: [3, 6, 9, 10], stockWaste: true }} />,
      );
      expect(container.querySelector('[data-skeleton-section="stock-waste"]')).not.toBeNull();
    });
  });

  describe('card-grid layout', () => {
    it('renders a grid with the requested cell count', () => {
      const { container } = render(
        <GameSkeleton
          gameKey="memory"
          layout={{ kind: 'card-grid', count: 52, cols: 'grid-cols-13', aspectRatio: 'aspect-[2/3]' }}
        />,
      );
      const grid = container.querySelector('[data-skeleton-section="card-grid"]');
      expect(grid).not.toBeNull();
      expect(grid?.querySelectorAll('.animate-pulse')).toHaveLength(52);
    });

    it('renders top pills row when topPills set', () => {
      const { container } = render(
        <GameSkeleton gameKey="memory" layout={{ kind: 'card-grid', count: 52, cols: 'grid-cols-13', topPills: 4 }} />,
      );
      const pills = container.querySelector('[data-skeleton-section="top-pills"]');
      expect(pills).not.toBeNull();
      expect(pills?.querySelectorAll('.animate-pulse')).toHaveLength(4);
    });
  });

  describe('centered layout', () => {
    it('renders one row per rows[] entry', () => {
      const { container } = render(<GameSkeleton gameKey="fiftyone" layout={{ kind: 'centered', rows: [5, 5] }} />);
      const rows = container.querySelectorAll('[data-skeleton-section="centered-row"]');
      expect(rows).toHaveLength(2);
    });

    it('renders bars below cards when bars set', () => {
      const { container } = render(
        <GameSkeleton gameKey="pigtail" layout={{ kind: 'centered', rows: [2], shape: 'circle', bars: 4 }} />,
      );
      expect(container.querySelector('[data-skeleton-section="centered-bars"]')).not.toBeNull();
    });

    it('uses circle shape when shape="circle"', () => {
      const { container } = render(
        <GameSkeleton gameKey="pigtail" layout={{ kind: 'centered', rows: [2], shape: 'circle' }} />,
      );
      const row = container.querySelector('[data-skeleton-section="centered-row"]');
      const firstShape = row?.querySelector('.animate-pulse');
      expect(firstShape?.className).toContain('rounded-full');
    });

    it('uses gap-2 by default for card shape (FiftyOne)', () => {
      const { container } = render(<GameSkeleton gameKey="fiftyone" layout={{ kind: 'centered', rows: [5] }} />);
      const row = container.querySelector('[data-skeleton-section="centered-row"]');
      expect(row?.className).toContain('gap-2');
      expect(row?.className).not.toContain('gap-8');
    });

    it('uses gap-8 when gap="wide" (War)', () => {
      const { container } = render(
        <GameSkeleton gameKey="war" layout={{ kind: 'centered', rows: [2], gap: 'wide' }} />,
      );
      const row = container.querySelector('[data-skeleton-section="centered-row"]');
      expect(row?.className).toContain('gap-8');
    });

    it('uses gap-8 by default for circle shape (PigsTail)', () => {
      const { container } = render(
        <GameSkeleton gameKey="pigtail" layout={{ kind: 'centered', rows: [2], shape: 'circle' }} />,
      );
      const row = container.querySelector('[data-skeleton-section="centered-row"]');
      expect(row?.className).toContain('gap-8');
    });
  });
});
