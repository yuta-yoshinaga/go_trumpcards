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
    expect(wrapper.className).toContain('ring-blue-400');
  });

  it('does not apply highlight class when isDropTarget is false', () => {
    const { container } = render(
      <DropZone isDropTarget={false} onDragOver={vi.fn()} onDrop={vi.fn()}>
        <span>child</span>
      </DropZone>,
    );
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).not.toContain('ring-blue-400');
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
