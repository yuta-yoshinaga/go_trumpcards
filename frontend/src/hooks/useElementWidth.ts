import { useCallback, useRef, useState } from 'react';

/** Measures an attached element and tracks its rendered width with ResizeObserver. */
export function useElementWidth<T extends HTMLElement>(): [ref: (node: T | null) => void, width: number] {
  const [width, setWidth] = useState(0);
  const observerRef = useRef<ResizeObserver | null>(null);

  const ref = useCallback((node: T | null) => {
    observerRef.current?.disconnect();
    observerRef.current = null;
    // A detached or replaced node must not keep reporting the previous element's width.
    setWidth(0);
    if (!node || typeof ResizeObserver === 'undefined') return;

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const measuredWidth = Math.floor(entry.contentRect.width);
        setWidth((currentWidth) => (currentWidth === measuredWidth ? currentWidth : measuredWidth));
      }
    });
    observerRef.current = observer;
    observer.observe(node);
  }, []);

  return [ref, width];
}
