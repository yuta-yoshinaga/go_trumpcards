import { useCardDimensions } from '../../hooks/useCardDimensions';
import { type GameKey, gameTheme } from '../../styles/gameTheme';
import { SkeletonCard } from './SkeletonCard';
import { SkeletonGrid } from './SkeletonGrid';
import { SkeletonHand } from './SkeletonHand';
import { SkeletonShell } from './SkeletonShell';

/** Centered hand sections (BlackJack/Baccarat-style table games). */
export interface CasinoTableLayout {
  kind: 'casino-table';
  /** Number of cards per centered hand section (one section per entry). */
  sections: number[];
  /**
   * Footer style.
   * - `bet` (default): two stacked button-shaped placeholders.
   * - `hand`: labelled mini hand + button (BlackJack).
   */
  footerStyle?: 'bet' | 'hand';
  /** Cards in the footer hand when footerStyle === 'hand'. */
  footerHandSize?: number;
}

/** Optional community cards + opponent boxes + footer hand (Hold'em-style poker). */
export interface CommunityPokerLayout {
  kind: 'community-poker';
  /** Cards in the community-card row. Omit to hide the row. */
  community?: number;
  /** Number of opponent boxes. */
  opponents: number;
  /** Cards per opponent box. */
  opponentCards: number;
  /** Cards in the footer hand (player). */
  footerHandSize: number;
}

/** Title bar + opponent rows + optional trick area + footer hand (most trick-taking & matching games). */
export interface TrickTakingLayout {
  kind: 'trick-taking';
  /** Number of opponent rows (default 3). */
  opponents?: number;
  /** Opponent row style: `bar` (text placeholder) or `hand` (mini hand). Default `bar`. */
  opponentStyle?: 'bar' | 'hand';
  /** Cards per opponent when opponentStyle === 'hand' (default 4). */
  opponentHandSize?: number;
  /** Render a center-card placeholder above opponents. */
  centerCard?: boolean;
  /** Render a trick-area placeholder below opponents. */
  trickArea?: boolean;
  /** Render the title bar at the top (default true). */
  titleBar?: boolean;
  /** Cards in the footer hand. */
  footerHandSize: number;
  /** Footer button width: `standard` (w-32) or `wide` (w-48). Default `standard`. */
  footerButton?: 'standard' | 'wide';
}

/** Foundation-style top row + tableau columns (Klondike-style solitaire). */
export interface TableauLayout {
  kind: 'tableau';
  /** Cards in the top (foundation/stock/freecells) row. */
  topRow: number;
  /** Columns in the tableau row. */
  tableau: number;
}

/** Triangle/columns of cards + optional stock-waste row (Pyramid/TriPeaks/Golf). */
export interface TieredRowsLayout {
  kind: 'tiered-rows';
  /** Card counts per row. */
  rows: number[];
  /** Render stock+waste pair below. */
  stockWaste?: boolean;
  /** Use column layout (Golf) instead of pyramid rows. */
  columns?: boolean;
}

/** Full grid of cards (Memory, Sevens). */
export interface CardGridLayout {
  kind: 'card-grid';
  count: number;
  cols: string;
  aspectRatio?: string;
  gridClassName?: string;
  /** Optional row of N pill placeholders above the grid (player-score row). */
  topPills?: number;
  /** When set, footer renders a SkeletonHand of this size + a button (Sevens). */
  footerHandSize?: number;
}

/** Centered placeholder for War/PigsTail/FiftyOne. */
export interface CenteredLayout {
  kind: 'centered';
  /** Card counts per row (centered). */
  rows: number[];
  /** Card shape (default `card`). `circle` is used by PigsTail. */
  shape?: 'card' | 'circle';
  /**
   * Inter-item gap. `narrow` (default) uses `gap-2` (compact rows like
   * FiftyOne). `wide` uses `gap-8` (spread out, used by War's two-card duel
   * and PigsTail's two-circle arrangement).
   */
  gap?: 'narrow' | 'wide';
  /** Number of small bar placeholders to render below the cards. */
  bars?: number;
}

/** Discriminated union of all supported skeleton layouts. */
export type SkeletonLayout =
  | CasinoTableLayout
  | CommunityPokerLayout
  | TrickTakingLayout
  | TableauLayout
  | TieredRowsLayout
  | CardGridLayout
  | CenteredLayout;

/** Props for the unified game skeleton. */
export interface GameSkeletonProps {
  /** Game key for theme resolution from `gameTheme`. */
  gameKey: GameKey;
  /** Layout descriptor. */
  layout: SkeletonLayout;
}

/**
 * Renders a loading skeleton placeholder for any game page using a data-driven
 * layout. The background and footer styling are auto-resolved from
 * `gameTheme[gameKey]`, and the body+footer content come from the `layout`
 * descriptor. Replaces the previous per-game `*Skeleton.tsx` files.
 */
export function GameSkeleton({ gameKey, layout }: GameSkeletonProps) {
  const theme = gameTheme[gameKey];
  const { cardWidth, cardHeight } = useCardDimensions();
  const { bodyClassName, footerPadding, body, footer } = renderLayout(layout, cardWidth, cardHeight);
  return (
    <SkeletonShell
      bgClass={theme.bg}
      bodyClassName={bodyClassName}
      footerClassName={`${theme.footer} ${footerPadding}`}
      footer={footer}
    >
      {body}
    </SkeletonShell>
  );
}

interface RenderedLayout {
  bodyClassName?: string;
  footerPadding: string;
  body: React.ReactNode;
  footer: React.ReactNode;
}

function renderLayout(layout: SkeletonLayout, cardWidth: number, cardHeight: number): RenderedLayout {
  switch (layout.kind) {
    case 'casino-table':
      return renderCasinoTable(layout, cardWidth, cardHeight);
    case 'community-poker':
      return renderCommunityPoker(layout, cardWidth, cardHeight);
    case 'trick-taking':
      return renderTrickTaking(layout, cardWidth, cardHeight);
    case 'tableau':
      return renderTableau(layout, cardWidth, cardHeight);
    case 'tiered-rows':
      return renderTieredRows(layout, cardWidth, cardHeight);
    case 'card-grid':
      return renderCardGrid(layout, cardWidth, cardHeight);
    case 'centered':
      return renderCentered(layout);
    default: {
      const _exhaustive: never = layout;
      throw new Error(`Unhandled SkeletonLayout: ${JSON.stringify(_exhaustive)}`);
    }
  }
}

function renderCasinoTable(layout: CasinoTableLayout, cardWidth: number, cardHeight: number): RenderedLayout {
  const isHand = layout.footerStyle === 'hand';
  const handSize = layout.footerHandSize ?? 5;
  const body = layout.sections.map((cards, i) => (
    <div key={i} data-skeleton-section="casino-table-section" className="mb-4">
      <div className="h-5 w-24 rounded bg-white/10 animate-pulse mx-auto mb-1" />
      <div className="flex justify-center gap-2">
        <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={cards} />
      </div>
    </div>
  ));
  const footer = isHand ? (
    <>
      <div className="mb-2">
        <div className="h-5 w-20 rounded bg-white/10 animate-pulse mb-1" />
        <div data-skeleton-section="footer-hand">
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={handSize} />
        </div>
      </div>
      <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
    </>
  ) : (
    <div className="flex flex-col items-center gap-2 pb-2">
      <div className="h-8 w-32 rounded bg-white/10 animate-pulse" />
      <div className="h-8 w-24 rounded bg-white/10 animate-pulse" />
    </div>
  );
  return {
    bodyClassName: isHand ? 'p-4' : undefined,
    footerPadding: isHand ? 'px-4 py-3' : 'px-4 pt-3',
    body,
    footer,
  };
}

function renderCommunityPoker(layout: CommunityPokerLayout, cardWidth: number, cardHeight: number): RenderedLayout {
  const body = (
    <>
      {layout.community !== undefined && (
        <div data-skeleton-section="community" className="mb-4">
          <div className="h-5 w-32 rounded bg-white/10 animate-pulse mb-1.5" />
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={layout.community} />
        </div>
      )}
      {Array.from({ length: layout.opponents }, (_, i) => (
        <div key={i} data-skeleton-section="opponent" className="mb-3 p-2 rounded bg-black/30">
          <div className="h-4 w-20 rounded bg-white/10 animate-pulse mb-2" />
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={layout.opponentCards} />
        </div>
      ))}
    </>
  );
  const footer = (
    <>
      <div className="mb-2">
        <div className="h-5 w-20 rounded bg-white/10 animate-pulse mb-1" />
        <div data-skeleton-section="footer-hand">
          <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={layout.footerHandSize} />
        </div>
      </div>
      <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
    </>
  );
  return {
    bodyClassName: 'pt-4 px-5',
    footerPadding: 'px-5 py-3',
    body,
    footer,
  };
}

function renderTrickTaking(layout: TrickTakingLayout, cardWidth: number, cardHeight: number): RenderedLayout {
  const opponents = layout.opponents ?? 3;
  const opponentStyle = layout.opponentStyle ?? 'bar';
  const opponentHandSize = layout.opponentHandSize ?? 4;
  const titleBar = layout.titleBar ?? true;
  const footerButtonWidth = layout.footerButton === 'wide' ? 'w-48' : 'w-32';
  const buttonExtraClass = layout.footerButton === 'wide' ? 'mx-auto' : '';
  const body = (
    <>
      {titleBar && (
        <div data-skeleton-section="title-bar" className="h-5 w-48 rounded bg-white/10 animate-pulse mx-auto mb-2" />
      )}
      {layout.centerCard && (
        <div data-skeleton-section="center-card" className="my-3 p-2 rounded bg-black/30 flex items-center gap-3">
          <div className="h-16 w-12 rounded bg-white/10 animate-pulse" />
          <div className="h-4 w-24 rounded bg-white/10 animate-pulse" />
        </div>
      )}
      {opponentStyle === 'bar'
        ? Array.from({ length: opponents }, (_, i) => (
            <div key={i} data-skeleton-section="opponent" className="mb-2 p-2 rounded bg-black/30">
              <div className="h-4 w-32 rounded bg-white/10 animate-pulse" />
            </div>
          ))
        : Array.from({ length: opponents }, (_, i) => (
            <div key={i} data-skeleton-section="opponent" className="mb-2 p-2 rounded bg-black/30">
              <div className="h-4 w-16 rounded bg-white/10 animate-pulse mb-2" />
              <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={opponentHandSize} />
            </div>
          ))}
      {layout.trickArea && (
        <div data-skeleton-section="trick-area" className="my-3 p-2 rounded bg-black/30">
          <div className="h-4 w-16 rounded bg-white/10 animate-pulse mb-1" />
          <div className="h-20 w-full rounded bg-white/10 animate-pulse" />
        </div>
      )}
    </>
  );
  const footer = (
    <>
      <div data-skeleton-section="footer-hand" className="mb-2">
        <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={layout.footerHandSize} />
      </div>
      <div className={`h-8 ${footerButtonWidth} rounded bg-white/10 animate-pulse ${buttonExtraClass}`.trim()} />
    </>
  );
  return {
    footerPadding: 'px-4 py-2.5',
    body,
    footer,
  };
}

function renderTableau(layout: TableauLayout, cardWidth: number, cardHeight: number): RenderedLayout {
  const body = (
    <>
      <div data-skeleton-section="tableau-top" className="flex gap-2 mb-3 items-start flex-wrap">
        {Array.from({ length: layout.topRow }, (_, i) => (
          <SkeletonCard key={i} width={cardWidth} height={cardHeight} />
        ))}
      </div>
      <div data-skeleton-section="tableau-cols" className="flex gap-2 mb-3">
        {Array.from({ length: layout.tableau }, (_, i) => (
          <div key={i} className="flex-1 min-w-0">
            <SkeletonCard width={cardWidth} height={cardHeight} />
          </div>
        ))}
      </div>
    </>
  );
  const footer = <div className="h-8 w-48 rounded bg-white/10 animate-pulse mx-auto" />;
  return {
    footerPadding: 'px-4 py-2.5',
    body,
    footer,
  };
}

function renderTieredRows(layout: TieredRowsLayout, cardWidth: number, cardHeight: number): RenderedLayout {
  const body = (
    <>
      {layout.columns ? (
        <div className="flex gap-1 justify-center">
          {layout.rows.map((rowCount, col) => (
            <div key={col} data-skeleton-section="tiered-row" className="flex flex-col gap-1">
              {Array.from({ length: rowCount }, (_, row) => (
                <SkeletonCard key={row} width={cardWidth} height={cardHeight} />
              ))}
            </div>
          ))}
        </div>
      ) : (
        layout.rows.map((rowCount, row) => (
          <div key={row} data-skeleton-section="tiered-row" className="flex justify-center gap-1 mb-1">
            {Array.from({ length: rowCount }, (_, col) => (
              <SkeletonCard key={col} width={cardWidth} height={cardHeight} />
            ))}
          </div>
        ))
      )}
      {layout.stockWaste && (
        <div data-skeleton-section="stock-waste" className="flex gap-2 mt-3">
          <SkeletonCard width={cardWidth} height={cardHeight} />
          <SkeletonCard width={cardWidth} height={cardHeight} />
        </div>
      )}
    </>
  );
  const footer = <div className="h-8 w-48 rounded bg-white/10 animate-pulse mx-auto" />;
  return {
    footerPadding: 'px-4 py-2.5',
    body,
    footer,
  };
}

function renderCardGrid(layout: CardGridLayout, cardWidth: number, cardHeight: number): RenderedLayout {
  const hasHandFooter = layout.footerHandSize !== undefined;
  const body = (
    <>
      {layout.topPills !== undefined && (
        <div
          data-skeleton-section="top-pills"
          className={
            hasHandFooter
              ? 'flex gap-2.5 flex-wrap mb-2.5'
              : 'my-1 px-2 py-1 rounded bg-black/30 flex gap-3 lg:shrink-0'
          }
        >
          {Array.from({ length: layout.topPills }, (_, i) =>
            hasHandFooter ? (
              <div key={i} className="p-2 rounded bg-black/30">
                <div className="h-4 w-16 rounded bg-white/10 animate-pulse" />
              </div>
            ) : (
              <div key={i} className="h-4 w-20 rounded bg-white/10 animate-pulse" />
            ),
          )}
        </div>
      )}
      <div
        data-skeleton-section="card-grid"
        className={
          hasHandFooter
            ? 'bg-black/30 rounded p-2 my-2'
            : 'my-3 lg:my-1 p-2 lg:p-1 rounded bg-black/40 lg:flex-1 lg:min-h-0 lg:overflow-hidden'
        }
      >
        <SkeletonGrid
          count={layout.count}
          cols={layout.cols}
          aspectRatio={layout.aspectRatio}
          gridClassName={layout.gridClassName}
        />
      </div>
    </>
  );
  const footer = hasHandFooter ? (
    <>
      <div data-skeleton-section="footer-hand" className="mb-2">
        <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={layout.footerHandSize ?? 5} />
      </div>
      <div className="h-8 w-32 rounded bg-white/10 animate-pulse" />
    </>
  ) : (
    <div className="h-8 w-24 rounded bg-white/10 animate-pulse" />
  );
  return {
    bodyClassName: hasHandFooter ? undefined : 'pt-3 lg:pt-1 px-4 lg:px-8 lg:flex lg:flex-col lg:overflow-hidden',
    footerPadding: 'px-4 py-2.5',
    body,
    footer,
  };
}

function renderCentered(layout: CenteredLayout): RenderedLayout {
  const shapeClass = layout.shape === 'circle' ? 'h-20 w-20 rounded-full' : 'h-24 w-16 rounded';
  // `gap` defaults to `wide` for circles (PigsTail) and `narrow` otherwise.
  // Static strings are required so Tailwind's JIT scanner can detect the classes.
  const gap = (layout.gap ?? (layout.shape === 'circle' ? 'wide' : 'narrow')) === 'wide' ? 'gap-8' : 'gap-2';
  const body = (
    <>
      {layout.rows.map((cardCount, rowIdx) => (
        <div
          key={rowIdx}
          data-skeleton-section="centered-row"
          className={`flex justify-center ${gap} ${rowIdx === 0 ? (layout.bars ? 'mb-4' : 'my-8') : 'mt-2'}`}
        >
          {Array.from({ length: cardCount }, (_, i) => (
            <div key={i} className={`${shapeClass} bg-white/10 animate-pulse`} />
          ))}
        </div>
      ))}
      {layout.bars !== undefined && (
        <div data-skeleton-section="centered-bars">
          {Array.from({ length: layout.bars }, (_, i) => (
            <div key={i} className="mb-2 p-2 rounded bg-black/30">
              <div className="h-4 w-32 rounded bg-white/10 animate-pulse" />
            </div>
          ))}
        </div>
      )}
    </>
  );
  const footer = <div className="h-10 w-48 rounded bg-white/10 animate-pulse mx-auto" />;
  return {
    footerPadding: 'px-4 py-2.5',
    body,
    footer,
  };
}
