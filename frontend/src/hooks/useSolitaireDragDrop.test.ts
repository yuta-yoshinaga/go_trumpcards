import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useSolitaireDragDrop } from './useSolitaireDragDrop';

interface TestZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

function createDragEvent(overrides: Partial<React.DragEvent> = {}): React.DragEvent {
  return {
    preventDefault: vi.fn(),
    dataTransfer: {
      setData: vi.fn(),
      getData: vi.fn(),
      effectAllowed: '',
    },
    ...overrides,
  } as unknown as React.DragEvent;
}

describe('useSolitaireDragDrop', () => {
  const defaultOptions = {
    onMove: vi.fn(),
    isPlaying: true,
    disabled: false,
  };

  it('returns isDragging false initially', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>(defaultOptions));
    expect(result.current.isDragging).toBe(false);
    expect(result.current.dragSource).toBeNull();
    expect(result.current.dropTargetZone).toBeNull();
  });

  it('handleDragStart sets dragSource and isDragging', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>(defaultOptions));
    const zone: TestZone = { zone: 'tableau', col: 2, cardIndex: 3 };
    const event = createDragEvent();

    act(() => {
      result.current.handleDragStart(zone)(event);
    });

    expect(result.current.isDragging).toBe(true);
    expect(result.current.dragSource).toEqual(zone);
    expect(event.dataTransfer.setData).toHaveBeenCalledWith('application/json', JSON.stringify(zone));
    expect(event.dataTransfer.effectAllowed).toBe('move');
  });

  it('handleDragStart is no-op when not playing', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>({ ...defaultOptions, isPlaying: false }));
    const event = createDragEvent();

    act(() => {
      result.current.handleDragStart({ zone: 'tableau' })(event);
    });

    expect(result.current.isDragging).toBe(false);
    expect(event.dataTransfer.setData).not.toHaveBeenCalled();
  });

  it('handleDragStart is no-op when disabled', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>({ ...defaultOptions, disabled: true }));
    const event = createDragEvent();

    act(() => {
      result.current.handleDragStart({ zone: 'tableau' })(event);
    });

    expect(result.current.isDragging).toBe(false);
  });

  it('handleDragOver calls preventDefault and sets dropTargetZone', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>(defaultOptions));
    const zone: TestZone = { zone: 'foundation', col: 0 };
    const event = createDragEvent();

    // Start a drag first
    act(() => {
      result.current.handleDragStart({ zone: 'waste' })(createDragEvent());
    });

    act(() => {
      result.current.handleDragOver(zone)(event);
    });

    expect(event.preventDefault).toHaveBeenCalled();
    expect(result.current.dropTargetZone).toEqual(zone);
  });

  it('handleDragOver is no-op when not dragging', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>(defaultOptions));
    const event = createDragEvent();

    act(() => {
      result.current.handleDragOver({ zone: 'foundation' })(event);
    });

    expect(event.preventDefault).not.toHaveBeenCalled();
    expect(result.current.dropTargetZone).toBeNull();
  });

  it('handleDragLeave clears dropTargetZone', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>(defaultOptions));

    act(() => {
      result.current.handleDragStart({ zone: 'waste' })(createDragEvent());
    });
    act(() => {
      result.current.handleDragOver({ zone: 'foundation', col: 0 })(createDragEvent());
    });

    expect(result.current.dropTargetZone).not.toBeNull();

    act(() => {
      result.current.handleDragLeave();
    });

    expect(result.current.dropTargetZone).toBeNull();
  });

  it('handleDrop calls onMove with source and target, then clears state', () => {
    const onMove = vi.fn();
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>({ ...defaultOptions, onMove }));
    const source: TestZone = { zone: 'waste' };
    const target: TestZone = { zone: 'foundation', col: 1 };

    act(() => {
      result.current.handleDragStart(source)(createDragEvent());
    });

    const dropEvent = createDragEvent({
      dataTransfer: {
        setData: vi.fn(),
        getData: vi.fn().mockReturnValue(JSON.stringify(source)),
        effectAllowed: '',
      },
    } as unknown as Partial<React.DragEvent>);

    act(() => {
      result.current.handleDrop(target)(dropEvent);
    });

    expect(dropEvent.preventDefault).toHaveBeenCalled();
    expect(onMove).toHaveBeenCalledWith(source, target);
    expect(result.current.isDragging).toBe(false);
    expect(result.current.dragSource).toBeNull();
    expect(result.current.dropTargetZone).toBeNull();
  });

  it('handleDrop is no-op when disabled', () => {
    const onMove = vi.fn();
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>({ ...defaultOptions, onMove, disabled: true }));

    act(() => {
      result.current.handleDragStart({ zone: 'waste' })(createDragEvent());
    });

    const dropEvent = createDragEvent({
      dataTransfer: {
        setData: vi.fn(),
        getData: vi.fn().mockReturnValue(JSON.stringify({ zone: 'waste' })),
        effectAllowed: '',
      },
    } as unknown as Partial<React.DragEvent>);

    act(() => {
      result.current.handleDrop({ zone: 'foundation' })(dropEvent);
    });

    // disabled prevents drag start, so onMove should not be called
    expect(onMove).not.toHaveBeenCalled();
  });

  it('handleDrop is no-op when dataTransfer has no data', () => {
    const onMove = vi.fn();
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>({ ...defaultOptions, onMove }));

    act(() => {
      result.current.handleDragStart({ zone: 'waste' })(createDragEvent());
    });

    const dropEvent = createDragEvent({
      dataTransfer: {
        setData: vi.fn(),
        getData: vi.fn().mockReturnValue(''),
        effectAllowed: '',
      },
    } as unknown as Partial<React.DragEvent>);

    act(() => {
      result.current.handleDrop({ zone: 'foundation' })(dropEvent);
    });

    expect(onMove).not.toHaveBeenCalled();
  });

  it('handleDragEnd clears all drag state', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>(defaultOptions));

    act(() => {
      result.current.handleDragStart({ zone: 'waste' })(createDragEvent());
    });

    expect(result.current.isDragging).toBe(true);

    act(() => {
      result.current.handleDragEnd();
    });

    expect(result.current.isDragging).toBe(false);
    expect(result.current.dragSource).toBeNull();
    expect(result.current.dropTargetZone).toBeNull();
  });

  it('isDropTarget helper correctly identifies matching zone', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>(defaultOptions));

    act(() => {
      result.current.handleDragStart({ zone: 'waste' })(createDragEvent());
    });
    act(() => {
      result.current.handleDragOver({ zone: 'foundation', col: 2 })(createDragEvent());
    });

    expect(result.current.isDropTarget({ zone: 'foundation', col: 2 })).toBe(true);
    expect(result.current.isDropTarget({ zone: 'foundation', col: 3 })).toBe(false);
    expect(result.current.isDropTarget({ zone: 'tableau', col: 2 })).toBe(false);
  });

  it('isDragSource helper correctly identifies matching zone', () => {
    const { result } = renderHook(() => useSolitaireDragDrop<TestZone>(defaultOptions));

    act(() => {
      result.current.handleDragStart({ zone: 'tableau', col: 1, cardIndex: 3 })(createDragEvent());
    });

    expect(result.current.isDragSource({ zone: 'tableau', col: 1, cardIndex: 3 })).toBe(true);
    expect(result.current.isDragSource({ zone: 'tableau', col: 1, cardIndex: 4 })).toBe(false);
    expect(result.current.isDragSource({ zone: 'waste' })).toBe(false);
  });
});
