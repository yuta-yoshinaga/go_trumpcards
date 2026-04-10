import type { ReactNode } from 'react';

/** Props for the DropZone wrapper component. */
interface DropZoneProps {
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
}

/**
 * Thin wrapper that receives HTML5 drag events and applies a visual highlight
 * when it is the active drop target. Used to wrap cards and empty pile
 * placeholders in solitaire games.
 */
export function DropZone({ isDropTarget, onDragOver, onDrop, onDragLeave, className, children }: DropZoneProps) {
  const highlightClass = isDropTarget ? 'ring-2 ring-blue-400 rounded' : '';
  const combinedClass = [highlightClass, className].filter(Boolean).join(' ');
  return (
    // biome-ignore lint/a11y/noStaticElementInteractions: drag-and-drop is a progressive enhancement; click-based interaction on inner buttons is the accessible path.
    <div
      className={combinedClass}
      onDragOver={onDragOver}
      onDrop={onDrop}
      onDragLeave={onDragLeave}
      role="presentation"
    >
      {children}
    </div>
  );
}
