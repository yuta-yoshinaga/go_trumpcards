import type { ReactNode } from 'react';

/** Props for the DropZone wrapper component. */
export interface DropZoneProps {
  /** Whether this zone is the currently active drop target (for highlighting). */
  isDropTarget: boolean;
  /** Handler invoked on dragOver (should call preventDefault to allow drop). */
  onDragOver: (e: React.DragEvent) => void;
  /** Handler invoked on drop. */
  onDrop: (e: React.DragEvent) => void;
  /** Optional handler invoked on dragLeave (to clear hover state). */
  onDragLeave?: () => void;
  /** Additional class names for the wrapper. */
  className?: string;
  /** Content rendered inside the drop zone. */
  children: ReactNode;
  /**
   * Accessible label describing what this zone represents (e.g., "ファウンデーション: スペード").
   * When provided together with `onKeyboardDrop`, a keyboard-accessible "move here" button is
   * rendered so assistive-tech users can drop the currently-selected card without drag events.
   */
  ariaLabel?: string;
  /**
   * Optional click handler that commits a move using the caller's currently-selected card.
   * When provided together with `ariaLabel`, a keyboard-focusable button appears (visually
   * hidden until focused). The wrapper gains `position: relative` automatically so the
   * focus overlay stays contained; callers do not need to add it to `className`.
   */
  onKeyboardDrop?: () => void;
  /** When true, the keyboard-drop button is rendered disabled (e.g., no card selected). */
  keyboardDropDisabled?: boolean;
  /** Label for the keyboard-drop button. Defaults to ariaLabel. */
  keyboardDropLabel?: string;
}

/**
 * Thin wrapper that receives HTML5 drag events and applies a visual highlight
 * when it is the active drop target. Used to wrap cards and empty pile
 * placeholders in solitaire games.
 *
 * When `onKeyboardDrop` is supplied, the zone becomes a `role="region"` and
 * exposes a visually-hidden-until-focused button so keyboard and screen-reader
 * users can drop the selected card without triggering HTML5 drag events.
 */
export function DropZone({
  isDropTarget,
  onDragOver,
  onDrop,
  onDragLeave,
  className,
  children,
  ariaLabel,
  onKeyboardDrop,
  keyboardDropDisabled,
  keyboardDropLabel,
}: DropZoneProps) {
  const keyboardAffordance = onKeyboardDrop !== undefined && !!ariaLabel;
  const highlightClass = isDropTarget ? 'ring-2 ring-ds-info rounded' : '';
  const positionClass = keyboardAffordance ? 'relative' : '';
  const combinedClass = [positionClass, highlightClass, className].filter(Boolean).join(' ');
  const label = keyboardDropLabel ?? ariaLabel;
  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: drag-and-drop is a progressive enhancement; the keyboard affordance below is the accessible path.
    // biome-ignore lint/a11y/useAriaPropsSupportedByRole: aria-label is only set when role is "region", which supports it.
    <div
      className={combinedClass}
      onDragOver={onDragOver}
      onDrop={onDrop}
      onDragLeave={onDragLeave}
      role={keyboardAffordance ? 'region' : 'presentation'}
      aria-label={keyboardAffordance ? ariaLabel : undefined}
    >
      {children}
      {keyboardAffordance && (
        <button
          type="button"
          onClick={onKeyboardDrop}
          disabled={keyboardDropDisabled}
          aria-label={label}
          className="sr-only focus-visible:not-sr-only focus-visible:absolute focus-visible:inset-0 focus-visible:z-10 focus-visible:flex focus-visible:items-center focus-visible:justify-center focus-visible:rounded focus-visible:bg-black/80 focus-visible:px-2 focus-visible:py-1 focus-visible:text-white focus-visible:text-xs focus-visible:ring-2 focus-visible:ring-ds-accent disabled:opacity-50"
        >
          {label}
        </button>
      )}
    </div>
  );
}
