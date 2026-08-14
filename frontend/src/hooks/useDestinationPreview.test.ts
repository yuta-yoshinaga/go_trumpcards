import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { useDestinationPreview } from './useDestinationPreview';

interface Src {
  col: number;
}

describe('useDestinationPreview', () => {
  it('has no source until something is hovered or selected', () => {
    const { result } = renderHook(() => useDestinationPreview<Src>(null));
    expect(result.current.source).toBeNull();
    expect(result.current.isPreview).toBe(false);
  });

  it('reports the hovered card as a preview source', () => {
    const { result } = renderHook(() => useDestinationPreview<Src>(null));
    act(() => result.current.previewProps({ col: 3 }).onMouseEnter());
    expect(result.current.source).toEqual({ col: 3 });
    expect(result.current.isPreview).toBe(true);
  });

  it('drops the preview when the pointer leaves', () => {
    const { result } = renderHook(() => useDestinationPreview<Src>(null));
    act(() => result.current.previewProps({ col: 3 }).onMouseEnter());
    act(() => result.current.previewProps({ col: 3 }).onMouseLeave());
    expect(result.current.source).toBeNull();
    expect(result.current.isPreview).toBe(false);
  });

  // **キーボードでも同じものが見えるべき。** hover だけだとフォーカス移動では
  // 何も起きない。
  it('treats focus like hover', () => {
    const { result } = renderHook(() => useDestinationPreview<Src>(null));
    act(() => result.current.previewProps({ col: 1 }).onFocus());
    expect(result.current.source).toEqual({ col: 1 });
    expect(result.current.isPreview).toBe(true);
    act(() => result.current.previewProps({ col: 1 }).onBlur());
    expect(result.current.source).toBeNull();
  });

  // **選択が hover に勝つ。** 勝たないと、選んだ札の移動先がカーソルを追って
  // 消え、狙っている最中に消える。
  it('keeps the selection while the pointer moves over other cards', () => {
    const { result, rerender } = renderHook(({ sel }: { sel: Src | null }) => useDestinationPreview<Src>(sel), {
      initialProps: { sel: null as Src | null },
    });
    rerender({ sel: { col: 0 } });
    expect(result.current.source).toEqual({ col: 0 });
    expect(result.current.isPreview).toBe(false);

    act(() => result.current.previewProps({ col: 5 }).onMouseEnter());
    expect(result.current.source).toEqual({ col: 0 });
    expect(result.current.isPreview).toBe(false);
  });

  // 選択が解けたら、その時点で指している札のプレビューに戻る。
  it('falls back to the hovered card once the selection is cleared', () => {
    const { result, rerender } = renderHook(({ sel }: { sel: Src | null }) => useDestinationPreview<Src>(sel), {
      initialProps: { sel: { col: 0 } as Src | null },
    });
    act(() => result.current.previewProps({ col: 5 }).onMouseEnter());
    rerender({ sel: null });
    expect(result.current.source).toEqual({ col: 5 });
    expect(result.current.isPreview).toBe(true);
  });

  it('clears on demand', () => {
    const { result } = renderHook(() => useDestinationPreview<Src>(null));
    act(() => result.current.previewProps({ col: 2 }).onMouseEnter());
    act(() => result.current.clear());
    expect(result.current.source).toBeNull();
  });

  it('keeps the handler factory stable across renders', () => {
    const { result, rerender } = renderHook(() => useDestinationPreview<Src>(null));
    const first = result.current.previewProps;
    rerender();
    expect(result.current.previewProps).toBe(first);
  });
});
