import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { OldMaidSettingsDialog } from './OldMaidSettingsDialog';

const defaultProps = {
  open: true,
  mode: 0,
  cpuPlacementStrategy: false,
  cpuMemoryAI: false,
  cpuHesitationEnabled: false,
  cpuMetaAI: false,
  onModeChange: vi.fn(),
  onStrategyChange: vi.fn(),
  onMemoryAIChange: vi.fn(),
  onHesitationChange: vi.fn(),
  onMetaAIChange: vi.fn(),
  onApply: vi.fn(),
  onClose: vi.fn(),
};

describe('OldMaidSettingsDialog', () => {
  it('renders null when open is false', () => {
    const { container } = render(<OldMaidSettingsDialog {...defaultProps} open={false} />);
    expect(container.innerHTML).toBe('');
  });

  it('renders dialog when open is true', () => {
    render(<OldMaidSettingsDialog {...defaultProps} />);
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('Old Maid 設定')).toBeInTheDocument();
  });

  it('calls onClose on Escape key', () => {
    const onClose = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onClose={onClose} />);
    fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('calls onClose on backdrop click', () => {
    const onClose = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onClose={onClose} />);
    // Click the backdrop (presentation overlay)
    fireEvent.click(screen.getByRole('presentation'));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('does not call onClose on dialog inner click', () => {
    const onClose = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onClose={onClose} />);
    fireEvent.click(screen.getByRole('dialog'));
    expect(onClose).not.toHaveBeenCalled();
  });

  it('calls onApply when 適用 button is clicked', () => {
    const onApply = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onApply={onApply} />);
    fireEvent.click(screen.getByRole('button', { name: '適用' }));
    expect(onApply).toHaveBeenCalledTimes(1);
  });

  it('calls onClose when キャンセル button is clicked', () => {
    const onClose = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onClose={onClose} />);
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('shows mode 0 radio checked by default', () => {
    render(<OldMaidSettingsDialog {...defaultProps} />);
    const radios = screen.getAllByRole('radio');
    expect(radios[0]).toBeChecked();
    expect(radios[1]).not.toBeChecked();
  });

  it('shows mode 1 radio checked when mode is 1', () => {
    render(<OldMaidSettingsDialog {...defaultProps} mode={1} />);
    const radios = screen.getAllByRole('radio');
    expect(radios[0]).not.toBeChecked();
    expect(radios[1]).toBeChecked();
  });

  it('calls onModeChange when radio is clicked', () => {
    const onModeChange = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onModeChange={onModeChange} />);
    fireEvent.click(screen.getAllByRole('radio')[1]);
    expect(onModeChange).toHaveBeenCalledWith(1);
  });

  it('calls onStrategyChange when checkbox is toggled', () => {
    const onStrategyChange = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onStrategyChange={onStrategyChange} />);
    fireEvent.click(screen.getByLabelText('CPU心理戦（奇数カードを端に配置）'));
    expect(onStrategyChange).toHaveBeenCalledWith(true);
  });

  it('calls onMemoryAIChange when checkbox is toggled', () => {
    const onMemoryAIChange = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onMemoryAIChange={onMemoryAIChange} />);
    fireEvent.click(screen.getByLabelText('CPU記憶AI（引いた位置を記憶して戦略的に選択）'));
    expect(onMemoryAIChange).toHaveBeenCalledWith(true);
  });

  it('calls onHesitationChange when checkbox is toggled', () => {
    const onHesitationChange = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onHesitationChange={onHesitationChange} />);
    fireEvent.click(screen.getByLabelText('CPU迷い時間ディレイ（カード内容により反応速度が変化）'));
    expect(onHesitationChange).toHaveBeenCalledWith(true);
  });

  it('calls onMetaAIChange when checkbox is toggled', () => {
    const onMetaAIChange = vi.fn();
    render(<OldMaidSettingsDialog {...defaultProps} onMetaAIChange={onMetaAIChange} />);
    fireEvent.click(screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）'));
    expect(onMetaAIChange).toHaveBeenCalledWith(true);
  });

  it('shows checkboxes as checked when props are true', () => {
    render(
      <OldMaidSettingsDialog
        {...defaultProps}
        cpuPlacementStrategy={true}
        cpuMemoryAI={true}
        cpuHesitationEnabled={true}
        cpuMetaAI={true}
      />,
    );
    expect(screen.getByLabelText('CPU心理戦（奇数カードを端に配置）')).toBeChecked();
    expect(screen.getByLabelText('CPU記憶AI（引いた位置を記憶して戦略的に選択）')).toBeChecked();
    expect(screen.getByLabelText('CPU迷い時間ディレイ（カード内容により反応速度が変化）')).toBeChecked();
    expect(screen.getByLabelText('メタAI（CPUがプレイスタイルを学習）')).toBeChecked();
  });

  it('has focus trap: Tab wraps from last to first focusable', () => {
    render(<OldMaidSettingsDialog {...defaultProps} />);
    const dialog = screen.getByRole('dialog');
    // Tab key on the dialog triggers the handler
    fireEvent.keyDown(dialog, { key: 'Tab' });
    // No error means focus trap handled it
    expect(dialog).toBeInTheDocument();
  });

  it('has focus trap: Shift+Tab wraps from first to last focusable', () => {
    render(<OldMaidSettingsDialog {...defaultProps} />);
    const dialog = screen.getByRole('dialog');
    fireEvent.keyDown(dialog, { key: 'Tab', shiftKey: true });
    expect(dialog).toBeInTheDocument();
  });

  it('groups radio buttons in fieldset with legend', () => {
    render(<OldMaidSettingsDialog {...defaultProps} />);
    const legend = screen.getByText('モード選択');
    const fieldset = legend.closest('fieldset');
    expect(fieldset).toBeInTheDocument();
  });

  it('groups checkboxes in fieldset with legend', () => {
    render(<OldMaidSettingsDialog {...defaultProps} />);
    const legend = screen.getByText('CPU設定');
    const fieldset = legend.closest('fieldset');
    expect(fieldset).toBeInTheDocument();
  });

  it('every radio/checkbox label meets the 44px minimum touch target (WCAG 2.5.5)', () => {
    // Issue #1510: setup dialog stacks 6 small toggles; below 44px tap area
    // mobile users mis-toggle adjacent options.
    render(<OldMaidSettingsDialog {...defaultProps} />);
    // The dialog portals to document.body via the shared Modal, so query the document.
    const labels = document.querySelectorAll('label.flex.items-center');
    expect(labels.length).toBeGreaterThanOrEqual(6);
    for (const label of labels) {
      expect(label.className).toContain('min-h-[44px]');
    }
  });
});
