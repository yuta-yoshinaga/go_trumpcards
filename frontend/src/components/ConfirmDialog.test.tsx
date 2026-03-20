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

  it('calls onCancel when Escape key is pressed on dialog', () => {
    const onCancel = vi.fn();
    render(<ConfirmDialog {...defaultProps({ onCancel })} />);
    fireEvent.keyDown(screen.getByRole('alertdialog'), { key: 'Escape' });
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('does not call onCancel for non-Escape key on dialog', () => {
    const onCancel = vi.fn();
    render(<ConfirmDialog {...defaultProps({ onCancel })} />);
    fireEvent.keyDown(screen.getByRole('alertdialog'), { key: 'Enter' });
    expect(onCancel).not.toHaveBeenCalled();
  });

  it('uses aria-labelledby referencing the title element', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    expect(screen.getByRole('alertdialog')).toHaveAttribute('aria-labelledby', 'confirm-dialog-title');
    expect(screen.getByText('リセット確認')).toHaveAttribute('id', 'confirm-dialog-title');
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

  it('does not trap Tab when not on last element', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    const cancelBtn = screen.getByRole('button', { name: 'キャンセル' });
    cancelBtn.focus();
    fireEvent.keyDown(screen.getByRole('alertdialog'), { key: 'Tab' });
    // Focus stays on cancel (no wrapping needed, browser handles normal tab)
    expect(document.activeElement).toBe(cancelBtn);
  });

  it('does not trap Shift+Tab when not on first element', () => {
    render(<ConfirmDialog {...defaultProps()} />);
    const confirmBtn = screen.getByRole('button', { name: '確認' });
    confirmBtn.focus();
    fireEvent.keyDown(screen.getByRole('alertdialog'), { key: 'Tab', shiftKey: true });
    // Focus stays on confirm (no wrapping needed)
    expect(document.activeElement).toBe(confirmBtn);
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
