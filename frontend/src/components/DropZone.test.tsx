import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { DropZone } from './DropZone';

describe('DropZone', () => {
  it('renders children', () => {
    render(
      <DropZone isDropTarget={false} onDragOver={vi.fn()} onDrop={vi.fn()}>
        <span>child content</span>
      </DropZone>,
    );
    expect(screen.getByText('child content')).toBeInTheDocument();
  });

  it('applies highlight class when isDropTarget is true', () => {
    const { container } = render(
      <DropZone isDropTarget={true} onDragOver={vi.fn()} onDrop={vi.fn()}>
        <span>child</span>
      </DropZone>,
    );
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).toContain('ring-2');
    expect(wrapper.className).toContain('ring-ds-info');
  });

  it('does not apply highlight class when isDropTarget is false', () => {
    const { container } = render(
      <DropZone isDropTarget={false} onDragOver={vi.fn()} onDrop={vi.fn()}>
        <span>child</span>
      </DropZone>,
    );
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).not.toContain('ring-ds-info');
  });

  it('calls onDragOver when drag-over event fires', () => {
    const onDragOver = vi.fn();
    const { container } = render(
      <DropZone isDropTarget={false} onDragOver={onDragOver} onDrop={vi.fn()}>
        <span>child</span>
      </DropZone>,
    );
    const wrapper = container.firstChild as HTMLElement;
    fireEvent.dragOver(wrapper);
    expect(onDragOver).toHaveBeenCalled();
  });

  it('calls onDrop when drop event fires', () => {
    const onDrop = vi.fn();
    const { container } = render(
      <DropZone isDropTarget={false} onDragOver={vi.fn()} onDrop={onDrop}>
        <span>child</span>
      </DropZone>,
    );
    const wrapper = container.firstChild as HTMLElement;
    fireEvent.drop(wrapper);
    expect(onDrop).toHaveBeenCalled();
  });

  it('renders a keyboard-drop button when onKeyboardDrop and ariaLabel are provided', () => {
    const onKeyboardDrop = vi.fn();
    render(
      <DropZone
        isDropTarget={false}
        onDragOver={vi.fn()}
        onDrop={vi.fn()}
        ariaLabel="ファウンデーション: スペード"
        onKeyboardDrop={onKeyboardDrop}
      >
        <span>child</span>
      </DropZone>,
    );
    const button = screen.getByRole('button', { name: 'ファウンデーション: スペード' });
    fireEvent.click(button);
    expect(onKeyboardDrop).toHaveBeenCalled();
  });

  it('disables the keyboard-drop button when keyboardDropDisabled is true', () => {
    render(
      <DropZone
        isDropTarget={false}
        onDragOver={vi.fn()}
        onDrop={vi.fn()}
        ariaLabel="foundation"
        onKeyboardDrop={vi.fn()}
        keyboardDropDisabled={true}
      >
        <span>child</span>
      </DropZone>,
    );
    expect(screen.getByRole('button', { name: 'foundation' })).toBeDisabled();
  });

  it('uses role="region" with aria-label when keyboard drop is enabled', () => {
    const { container } = render(
      <DropZone
        isDropTarget={false}
        onDragOver={vi.fn()}
        onDrop={vi.fn()}
        ariaLabel="ファウンデーション"
        onKeyboardDrop={vi.fn()}
      >
        <span>child</span>
      </DropZone>,
    );
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.getAttribute('role')).toBe('region');
    expect(wrapper.getAttribute('aria-label')).toBe('ファウンデーション');
  });

  it('does not use role="region" when ariaLabel is absent even if onKeyboardDrop is provided', () => {
    const { container } = render(
      <DropZone isDropTarget={false} onDragOver={vi.fn()} onDrop={vi.fn()} onKeyboardDrop={vi.fn()}>
        <span>child</span>
      </DropZone>,
    );
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.getAttribute('role')).toBe('presentation');
    expect(wrapper.getAttribute('aria-label')).toBeNull();
    expect(screen.queryByRole('button')).toBeNull();
  });

  it('adds position:relative when keyboard affordance is enabled', () => {
    const { container } = render(
      <DropZone
        isDropTarget={false}
        onDragOver={vi.fn()}
        onDrop={vi.fn()}
        ariaLabel="foundation"
        onKeyboardDrop={vi.fn()}
      >
        <span>child</span>
      </DropZone>,
    );
    expect((container.firstChild as HTMLElement).className).toContain('relative');
  });

  it('does not render a keyboard-drop button when onKeyboardDrop is omitted', () => {
    render(
      <DropZone isDropTarget={false} onDragOver={vi.fn()} onDrop={vi.fn()} ariaLabel="foundation">
        <span>child</span>
      </DropZone>,
    );
    expect(screen.queryByRole('button')).toBeNull();
  });

  it('calls onDragLeave when provided', () => {
    const onDragLeave = vi.fn();
    const { container } = render(
      <DropZone isDropTarget={false} onDragOver={vi.fn()} onDrop={vi.fn()} onDragLeave={onDragLeave}>
        <span>child</span>
      </DropZone>,
    );
    const wrapper = container.firstChild as HTMLElement;
    fireEvent.dragLeave(wrapper);
    expect(onDragLeave).toHaveBeenCalled();
  });
});
