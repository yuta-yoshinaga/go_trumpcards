import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { SevensBoard } from './SevensBoard';

// tablePlaced with 7 placed for all suits
const tablePlacedWith7 = [0, 1 << 7, 1 << 7, 1 << 7, 1 << 7];

const defaultProps = {
  tablePlaced: tablePlacedWith7,
  tunnelEnabled: false,
  tunnelSkipWidth: 0,
  endStopEnabled: false,
  jokerSelecting: false,
};

describe('SevensBoard', () => {
  const originalInnerWidth = window.innerWidth;

  afterEach(() => {
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: originalInnerWidth,
    });
  });

  it('renders board title', () => {
    render(<SevensBoard {...defaultProps} />);
    expect(screen.getByText('ボード')).toBeInTheDocument();
  });

  it('renders all 4 suit rows', () => {
    render(<SevensBoard {...defaultProps} />);
    expect(screen.getByText('♠')).toBeInTheDocument();
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders 13 cells per suit (52 total)', () => {
    const { container } = render(<SevensBoard {...defaultProps} />);
    const allCells = container.querySelectorAll('span[style*="width"]');
    expect(allCells).toHaveLength(52);
  });

  it('uses desktop cell dimensions on wide viewport', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    const { container } = render(<SevensBoard {...defaultProps} />);
    const cell = container.querySelector('span[style*="width"]') as HTMLElement;
    expect(cell.style.width).toBe('26px');
    expect(cell.style.height).toBe('26px');
    expect(cell.style.fontSize).toBe('0.75em');
  });

  it('uses mobile cell dimensions on narrow viewport', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    const { container } = render(<SevensBoard {...defaultProps} />);
    const cell = container.querySelector('span[style*="width"]') as HTMLElement;
    expect(cell.style.width).toBe('20px');
    expect(cell.style.height).toBe('20px');
    expect(cell.style.fontSize).toBe('0.6em');
  });

  it('renders responsive grid classes', () => {
    const { container } = render(<SevensBoard {...defaultProps} />);
    const grid = container.querySelector('.grid');
    expect(grid?.className).toContain('grid-cols-1');
    expect(grid?.className).toContain('sm:grid-cols-2');
  });

  it('shows tunnel tag when tunnelEnabled', () => {
    render(<SevensBoard {...defaultProps} tunnelEnabled={true} />);
    expect(screen.getByText('[トンネル]')).toBeInTheDocument();
  });

  it('does not show tunnel tag when tunnelEnabled is false', () => {
    render(<SevensBoard {...defaultProps} tunnelEnabled={false} />);
    expect(screen.queryByText('[トンネル]')).not.toBeInTheDocument();
  });

  it('shows joker select hint when jokerSelecting', () => {
    render(<SevensBoard {...defaultProps} jokerSelecting={true} />);
    expect(screen.getByText('配置先を選択してください')).toBeInTheDocument();
  });

  it('renders playable cells as buttons when jokerSelecting', () => {
    render(<SevensBoard {...defaultProps} jokerSelecting={true} />);
    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThan(0);
  });

  it('calls onJokerPlace when clicking a playable cell', () => {
    const onJokerPlace = vi.fn();
    render(<SevensBoard {...defaultProps} jokerSelecting={true} onJokerPlace={onJokerPlace} />);
    const buttons = screen.getAllByRole('button');
    buttons[0].click();
    expect(onJokerPlace).toHaveBeenCalledTimes(1);
  });

  it('shows tunnel connection icon for each suit when tunnelEnabled', () => {
    render(<SevensBoard {...defaultProps} tunnelEnabled={true} />);
    const icons = screen.getAllByLabelText('トンネル接続');
    expect(icons).toHaveLength(4);
  });

  it('does not show tunnel connection icon when tunnelEnabled is false', () => {
    render(<SevensBoard {...defaultProps} tunnelEnabled={false} />);
    expect(screen.queryByLabelText('トンネル接続')).not.toBeInTheDocument();
  });

  it('highlights tunnel wrap cells when A placed and tunnelEnabled', () => {
    const tablePlaced = [0, (1 << 1) | (1 << 7), 1 << 7, 1 << 7, 1 << 7];
    const { container } = render(<SevensBoard {...defaultProps} tablePlaced={tablePlaced} tunnelEnabled={true} />);
    const highlighted = container.querySelectorAll('.border-amber-400');
    expect(highlighted.length).toBeGreaterThanOrEqual(1);
  });

  it('joker playable button uses desktop dimensions on wide viewport', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1024 });
    render(<SevensBoard {...defaultProps} jokerSelecting={true} />);
    const button = screen.getAllByRole('button')[0];
    expect(button.style.width).toBe('26px');
    expect(button.style.height).toBe('26px');
  });

  it('does not throw when clicking playable cell without onJokerPlace', () => {
    render(<SevensBoard {...defaultProps} jokerSelecting={true} />);
    const buttons = screen.getAllByRole('button');
    expect(() => buttons[0].click()).not.toThrow();
  });

  it('highlights A cell when K is placed and tunnelEnabled', () => {
    const tablePlaced = [0, (1 << 13) | (1 << 7), 1 << 7, 1 << 7, 1 << 7];
    const { container } = render(<SevensBoard {...defaultProps} tablePlaced={tablePlaced} tunnelEnabled={true} />);
    const highlighted = container.querySelectorAll('.border-amber-400');
    expect(highlighted.length).toBeGreaterThanOrEqual(1);
  });

  it('does not highlight tunnel wrap cells when tunnelEnabled is false', () => {
    const tablePlaced = [0, (1 << 1) | (1 << 7), 1 << 7, 1 << 7, 1 << 7];
    const { container } = render(<SevensBoard {...defaultProps} tablePlaced={tablePlaced} tunnelEnabled={false} />);
    const highlighted = container.querySelectorAll('.border-amber-400');
    expect(highlighted).toHaveLength(0);
  });

  it('center cell (7) has bold font-weight when placed', () => {
    const { container } = render(<SevensBoard {...defaultProps} />);
    const cells = container.querySelectorAll('span[style*="width"]');
    const centerCell = cells[6] as HTMLElement;
    expect(centerCell.style.fontWeight).toBe('bold');
  });

  it('non-center placed cell has normal font-weight', () => {
    const tablePlaced = [0, (1 << 6) | (1 << 7), 1 << 7, 1 << 7, 1 << 7];
    const { container } = render(<SevensBoard {...defaultProps} tablePlaced={tablePlaced} />);
    const cells = container.querySelectorAll('span[style*="width"]');
    const nonCenterCell = cells[5] as HTMLElement;
    expect(nonCenterCell.style.fontWeight).toBe('normal');
  });
});
