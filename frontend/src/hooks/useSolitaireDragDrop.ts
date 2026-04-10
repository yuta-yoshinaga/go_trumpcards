import { useCallback, useState } from 'react';

/** Options for the solitaire drag-and-drop hook. */
interface UseSolitaireDragDropOptions<T> {
  /** Callback invoked when a card is dropped on a valid target. */
  onMove: (source: T, target: T) => void;
  /** Whether the game is in a playing state. */
  isPlaying: boolean;
  /** Whether interactions are disabled (e.g., loading, auto-completing). */
  disabled: boolean;
}

/** Return type for the solitaire drag-and-drop hook. */
interface UseSolitaireDragDropReturn<T> {
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
      setDropTargetZone(zone);
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
      return JSON.stringify(dropTargetZone) === JSON.stringify(zone);
    },
    [dropTargetZone],
  );

  const isDragSourceFn = useCallback(
    (zone: T): boolean => {
      if (!dragSource) return false;
      return JSON.stringify(dragSource) === JSON.stringify(zone);
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
