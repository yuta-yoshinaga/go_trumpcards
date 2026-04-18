import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ConfirmDialogProps } from './ConfirmDialog';
import { ConfirmDialog } from './ConfirmDialog';

function defaultProps(overrides?: Partial<ConfirmDialogProps>): ConfirmDialogProps {
  return {
    open: true,
    title: 'リセット確認',
    message: '本当にゲームをリセットしますか？',
    confirmLabel: '確認',
    cancelLabel: 'キャンセル',
    onConfirm: vi.fn(),
    onCancel: vi.fn(),
    ...overrides,
  };
}

describe('ConfirmDialog', () => {
  it('renders nothing when open is false', () => {
    const { container } = render(<ConfirmDialog {...defaultProps({ open: false })} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders dialog when open is true', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    expect(screen.getByText('リセット確認')).toBeInTheDocument();
    expect(screen.getByText('本当にゲームをリセットしますか？')).toBeInTheDocument();
  });

  it('has aria-modal attribute', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    expect(screen.getByRole('alertdialog')).toHaveAttribute('aria-modal', 'true');
  });

  it('calls onConfirm when confirm button is clicked', () => {
    const onConfirm = vi.fn();
    render(<ConfirmDialog {...defaultProps({ onConfirm })} />);
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(onConfirm).toHaveBeenCalledOnce();
  });

  it('calls onCancel when cancel button is clicked', () => {
    const onCancel = vi.fn();
    render(<ConfirmDialog {...defaultProps({ onCancel })} />);
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('calls onCancel when overlay is clicked', () => {
    const onCancel = vi.fn();
    render(<ConfirmDialog {...defaultProps({ onCancel })} />);
    // Click the overlay (presentation element)
    fireEvent.click(screen.getByRole('presentation'));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('does not call onCancel when dialog content is clicked', () => {
    const onCancel = vi.fn();
    render(<ConfirmDialog {...defaultProps({ onCancel })} />);
    fireEvent.click(screen.getByRole('alertdialog'));
    expect(onCancel).not.toHaveBeenCalled();
  });

  it('calls onCancel when Escape key is pressed at document level', () => {
    const onCancel = vi.fn();
    render(<ConfirmDialog {...defaultProps({ onCancel })} />);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('does not call onCancel for non-Escape key at document level', () => {
    const onCancel = vi.fn();
    render(<ConfirmDialog {...defaultProps({ onCancel })} />);
    fireEvent.keyDown(document, { key: 'Enter' });
    expect(onCancel).not.toHaveBeenCalled();
  });

  it('exposes the title via the accessible name', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    expect(screen.getByRole('alertdialog', { name: 'リセット確認' })).toBeInTheDocument();
  });

  it('exposes the message via the accessible description', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    expect(screen.getByRole('alertdialog', { description: '本当にゲームをリセットしますか？' })).toBeInTheDocument();
  });

  it('omits aria-describedby and the message paragraph when message is empty', () => {
    render(<ConfirmDialog {...defaultProps({ message: '' })} />);
    const dialog = screen.getByRole('alertdialog');
    expect(dialog).not.toHaveAttribute('aria-describedby');
    expect(screen.queryByText('本当にゲームをリセットしますか？')).not.toBeInTheDocument();
  });

  it('focuses cancel button on open', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'キャンセル' }));
  });

  it('traps focus forward: Tab on last button wraps to first', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    const confirmBtn = screen.getByRole('button', { name: '確認' });
    confirmBtn.focus();
    fireEvent.keyDown(screen.getByRole('alertdialog'), { key: 'Tab' });
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'キャンセル' }));
  });

  it('traps focus backward: Shift+Tab on first button wraps to last', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    const cancelBtn = screen.getByRole('button', { name: 'キャンセル' });
    cancelBtn.focus();
    fireEvent.keyDown(screen.getByRole('alertdialog'), { key: 'Tab', shiftKey: true });
    expect(document.activeElement).toBe(screen.getByRole('button', { name: '確認' }));
  });

  it('does not trap non-Tab keys', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    const cancelBtn = screen.getByRole('button', { name: 'キャンセル' });
    cancelBtn.focus();
    fireEvent.keyDown(screen.getByRole('alertdialog'), { key: 'Enter' });
    expect(document.activeElement).toBe(cancelBtn);
  });

  it('does not preventDefault for Tab on non-last element', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    const cancelBtn = screen.getByRole('button', { name: 'キャンセル' });
    cancelBtn.focus();
    const event = new KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true });
    screen.getByRole('alertdialog').dispatchEvent(event);
    expect(event.defaultPrevented).toBe(false);
  });

  it('does not preventDefault for Shift+Tab on non-first element', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    const confirmBtn = screen.getByRole('button', { name: '確認' });
    confirmBtn.focus();
    const event = new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true });
    screen.getByRole('alertdialog').dispatchEvent(event);
    expect(event.defaultPrevented).toBe(false);
  });

  it('restores focus to previously focused element on close', () => {
    const trigger = document.createElement('button');
    trigger.textContent = 'Trigger';
    document.body.appendChild(trigger);
    trigger.focus();

    const { rerender } = render(<ConfirmDialog {...defaultProps()} />);
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'キャンセル' }));

    rerender(<ConfirmDialog {...defaultProps({ open: false })} />);
    expect(document.activeElement).toBe(trigger);

    document.body.removeChild(trigger);
  });

  it('handles dialogRef being null and no focusable elements', () => {
    const origQuerySelectorAll = HTMLElement.prototype.querySelectorAll;
    let callCount = 0;
    vi.spyOn(HTMLElement.prototype, 'querySelectorAll').mockImplementation(function (
      this: HTMLElement,
      selector: string,
    ) {
      callCount++;
      // First call from useEffect: return empty to cover focusable.length === 0 branch
      if (callCount === 1) {
        return origQuerySelectorAll.call(document.createElement('div'), 'nonexistent');
      }
      return origQuerySelectorAll.call(this, selector);
    });

    render(<ConfirmDialog {...defaultProps()} />);
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    vi.restoreAllMocks();
  });

  it('handles cleanup when triggerRef is not an HTMLElement', () => {
    // Simulate document.activeElement being null by setting it before render
    const originalDescriptor = Object.getOwnPropertyDescriptor(document, 'activeElement');
    Object.defineProperty(document, 'activeElement', {
      get: () => null,
      configurable: true,
    });

    const { rerender } = render(<ConfirmDialog {...defaultProps()} />);

    // Restore real activeElement before rerender so DOM operations work
    if (originalDescriptor) {
      Object.defineProperty(document, 'activeElement', originalDescriptor);
    }

    // Closing dialog: triggerRef.current is null, instanceof HTMLElement is false
    rerender(<ConfirmDialog {...defaultProps({ open: false })} />);
    expect(screen.queryByRole('alertdialog')).toBeNull();
  });
});
