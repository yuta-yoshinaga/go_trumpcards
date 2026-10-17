import type { ReactElement } from 'react';
import { useCallback, useRef, useState } from 'react';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { DESKTOP_MIN_CARD_EXPOSURE_PX, fanOverlap, fanRows } from '../utils/cardFan';
import { MIN_CARD_EXPOSURE_PX } from '../utils/handOverlap';
import { splitBalanced } from '../utils/splitBalanced';

/**
 * Render keyed card elements in balanced rows that overlap to fit the measured width.
 * The width is the root's clientWidth, so callers must not give the root horizontal padding.
 */
export function CardFan({
  itemWidth,
  children,
  className,
  dataTutorial,
  'data-testid': testId,
}: {
  itemWidth: number;
  children: ReactElement[];
  className?: string;
  dataTutorial?: string;
  'data-testid'?: string;
}) {
  const [width, setWidth] = useState(0);
  const observerRef = useRef<ResizeObserver | null>(null);
  const ref = useCallback((element: HTMLDivElement | null) => {
    observerRef.current?.disconnect();
    observerRef.current = null;
    if (!element || typeof ResizeObserver === 'undefined') return;
    const observer = new ResizeObserver(() => setWidth(element.clientWidth));
    observerRef.current = observer;
    observer.observe(element);
    setWidth(element.clientWidth);
  }, []);
  const { isMobile } = useCardDimensions();
  const rootProps = { ref, className, 'data-testid': testId, 'data-tutorial': dataTutorial };
  if (width === 0)
    return (
      <div {...rootProps} className={`flex flex-wrap justify-center gap-2 ${className ?? ''}`}>
        {children}
      </div>
    );
  const exposure = isMobile ? MIN_CARD_EXPOSURE_PX : DESKTOP_MIN_CARD_EXPOSURE_PX;
  const rows = splitBalanced(children, fanRows(children.length, itemWidth, width, exposure));
  return (
    <div {...rootProps}>
      {rows.map(({ items, start }, rowIndex) => {
        const overlap = fanOverlap(items.length, itemWidth, width, exposure, 8);
        return (
          <div key={`fan-row-${rowIndex}`} className="flex flex-nowrap justify-center-safe">
            {items.map((child, index) => (
              <div
                key={child.key ?? start + index}
                className="shrink-0"
                style={{ marginLeft: index === 0 ? 0 : overlap }}
              >
                {child}
              </div>
            ))}
          </div>
        );
      })}
    </div>
  );
}
