import type { RefObject } from 'react';
import { useEffect, useState } from 'react';

/** Measures an element's rendered width while observation is enabled. */
export function useElementWidth(ref: RefObject<HTMLElement | null>, enabled: boolean): number {
  const [width, setWidth] = useState(0);

  useEffect(() => {
    const element = ref.current;
    if (!enabled || !element || typeof window.ResizeObserver === 'undefined') return;

    const observer = new window.ResizeObserver((entries) => {
      for (const entry of entries) {
        const measuredWidth = Math.floor(entry.contentRect.width);
        setWidth((currentWidth) => (currentWidth === measuredWidth ? currentWidth : measuredWidth));
      }
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, [enabled, ref]);

  return width;
}
