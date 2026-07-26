import { useCallback, useState } from 'react';

/** Options for the solitaire drag-and-drop hook. */
export interface UseSolitaireDragDropOptions<T> {
  /** Callback invoked when a card is dropped on a valid target. */
  onMove: (source: T, target: T) => void;
  /** Whether the game is in a playing state. */
  isPlaying: boolean;
  /** Whether interactions are disabled (e.g., loading, auto-completing). */
  disabled: boolean;
}

/** Return type for the solitaire drag-and-drop hook. */
export interface UseSolitaireDragDropReturn<T> {
  /** The zone currently being dragged, or null. */
  dragSource: T | null;
  /** Whether a drag operation is in progress. */
  isDragging: boolean;
  /** The zone currently hovered as a drop target, or null. */
  dropTargetZone: T | null;
  /** Returns a dragStart handler for the given source zone. */
  handleDragStart: (zone: T) => (e: React.DragEvent) => void;
  /** Returns a dragOver handler for the given target zone. */
  handleDragOver: (zone: T) => (e: React.DragEvent) => void;
  /** Clears the current drop target highlight. */
  handleDragLeave: () => void;
  /** Returns a drop handler for the given target zone. */
  handleDrop: (zone: T) => (e: React.DragEvent) => void;
  /** Clears all drag state (for cancelled drags). */
  handleDragEnd: () => void;
  /** Checks whether the given zone matches the current drop target. */
  isDropTarget: (zone: T) => boolean;
  /** Checks whether the given zone matches the current drag source. */
  isDragSource: (zone: T) => boolean;
}

const DND_DATA_TYPE = 'application/json';

/**
 * Compares two move zones by their structural fields. All solitaire move
 * zones share `zone`, `col`, `cardIndex`, and (for FreeCell) `cell`.
 * Avoids `JSON.stringify` cost on hot paths like dragOver and per-card
 * highlight checks.
 */
function zonesEqual<T extends { zone: string }>(a: T, b: T): boolean {
  const ax = a as { zone: string; col?: number; cardIndex?: number; cell?: number };
  const bx = b as { zone: string; col?: number; cardIndex?: number; cell?: number };
  return ax.zone === bx.zone && ax.col === bx.col && ax.cardIndex === bx.cardIndex && ax.cell === bx.cell;
}

/** Shared hook for HTML5 drag-and-drop across solitaire card games. */
export function useSolitaireDragDrop<T extends { zone: string }>({
  onMove,
  isPlaying,
  disabled,
}: UseSolitaireDragDropOptions<T>): UseSolitaireDragDropReturn<T> {
  const [dragSource, setDragSource] = useState<T | null>(null);
  const [dropTargetZone, setDropTargetZone] = useState<T | null>(null);

  const isDragging = dragSource !== null;

  const clearState = useCallback(() => {
    setDragSource(null);
    setDropTargetZone(null);
  }, []);

  const handleDragStart = useCallback(
    (zone: T) => (e: React.DragEvent) => {
      if (!isPlaying || disabled) return;
      e.dataTransfer.setData(DND_DATA_TYPE, JSON.stringify(zone));
      e.dataTransfer.effectAllowed = 'move';
      setDragSource(zone);
    },
    [isPlaying, disabled],
  );

  const handleDragOver = useCallback(
    (zone: T) => (e: React.DragEvent) => {
      if (!isDragging) return;
      e.preventDefault();
      e.dataTransfer.dropEffect = 'move';
      // Bail out when the hover target hasn't changed to avoid re-rendering
      // the entire game on every dragover event (~60fps).
      setDropTargetZone((prev) => (prev && zonesEqual(prev, zone) ? prev : zone));
    },
    [isDragging],
  );

  const handleDragLeave = useCallback(() => {
    setDropTargetZone(null);
  }, []);

  const handleDrop = useCallback(
    (zone: T) => (e: React.DragEvent) => {
      e.preventDefault();
      if (!isPlaying || disabled) {
        clearState();
        return;
      }
      const raw = e.dataTransfer.getData(DND_DATA_TYPE);
      if (!raw) {
        clearState();
        return;
      }
      try {
        const source = JSON.parse(raw) as T;
        onMove(source, zone);
      } finally {
        clearState();
      }
    },
    [onMove, clearState, isPlaying, disabled],
  );

  const handleDragEnd = useCallback(() => {
    clearState();
  }, [clearState]);

  const isDropTarget = useCallback(
    (zone: T): boolean => {
      if (!dropTargetZone) return false;
      return zonesEqual(dropTargetZone, zone);
    },
    [dropTargetZone],
  );

  const isDragSourceFn = useCallback(
    (zone: T): boolean => {
      if (!dragSource) return false;
      return zonesEqual(dragSource, zone);
    },
    [dragSource],
  );

  return {
    dragSource,
    isDragging,
    dropTargetZone,
    handleDragStart,
    handleDragOver,
    handleDragLeave,
    handleDrop,
    handleDragEnd,
    isDropTarget,
    isDragSource: isDragSourceFn,
  };
}
