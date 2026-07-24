import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { Modal } from './Modal';

describe('Modal', () => {
  it('renders nothing when closed', () => {
    render(
      <Modal open={false} onClose={() => {}}>
        <button type="button">inside</button>
      </Modal>,
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('portals an aria-modal dialog and moves focus inside on open', () => {
    render(
      <Modal open onClose={() => {}} ariaLabel="picker">
        <button type="button">first</button>
        <button type="button">second</button>
      </Modal>,
    );
    const dialog = screen.getByRole('dialog');
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    // Portaled to body, not nested in the render container.
    expect(dialog.closest('[data-testid="rtl-container"]')).toBeNull();
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'first' }));
  });

  it('closes on Escape', () => {
    const onClose = vi.fn();
    render(
      <Modal open onClose={onClose}>
        <button type="button">x</button>
      </Modal>,
    );
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('closes on backdrop click by default but not when dismissOnBackdrop is false', () => {
    const onClose = vi.fn();
    const { rerender } = render(
      <Modal open onClose={onClose}>
        <button type="button">x</button>
      </Modal>,
    );
    fireEvent.click(screen.getByRole('presentation'));
    expect(onClose).toHaveBeenCalledTimes(1);

    onClose.mockClear();
    rerender(
      <Modal open onClose={onClose} dismissOnBackdrop={false}>
        <button type="button">x</button>
      </Modal>,
    );
    fireEvent.click(screen.getByRole('presentation'));
    expect(onClose).not.toHaveBeenCalled();
  });

  it('does not close when the panel itself is clicked', () => {
    const onClose = vi.fn();
    render(
      <Modal open onClose={onClose}>
        <button type="button">x</button>
      </Modal>,
    );
    fireEvent.click(screen.getByRole('dialog'));
    expect(onClose).not.toHaveBeenCalled();
  });

  it('restores focus to the trigger element on close', () => {
    const trigger = document.createElement('button');
    document.body.appendChild(trigger);
    trigger.focus();
    expect(document.activeElement).toBe(trigger);

    const { rerender } = render(
      <Modal open onClose={() => {}}>
        <button type="button">x</button>
      </Modal>,
    );
    // focus moved into the dialog
    expect(document.activeElement).not.toBe(trigger);

    rerender(
      <Modal open={false} onClose={() => {}}>
        <button type="button">x</button>
      </Modal>,
    );
    expect(document.activeElement).toBe(trigger);
    document.body.removeChild(trigger);
  });
});
