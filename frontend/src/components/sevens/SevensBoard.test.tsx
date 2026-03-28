import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
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

  it('renders 52 board cells (13 per suit)', () => {
    render(<SevensBoard {...defaultProps} />);
    const cells = screen.getAllByTestId('board-cell');
    expect(cells).toHaveLength(52);
  });

  it('uses CSS grid with 14-column layout (suit label + 13 values)', () => {
    render(<SevensBoard {...defaultProps} />);
    const grid = screen.getByTestId('sevens-grid') as HTMLElement;
    expect(grid.style.gridTemplateColumns).toBe('auto repeat(13, 1fr)');
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
    screen.getByLabelText('SPADE 6 に配置').click();
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

  it('does not throw when clicking playable cell without onJokerPlace', () => {
    render(<SevensBoard {...defaultProps} jokerSelecting={true} />);
    expect(() => screen.getByLabelText('SPADE 6 に配置').click()).not.toThrow();
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

  it('center cell (7) has bold class', () => {
    render(<SevensBoard {...defaultProps} />);
    const cells = screen.getAllByTestId('board-cell');
    // 7th cell in first suit row (index 6)
    expect(cells[6].className).toContain('font-bold');
  });

  it('non-center placed cell does not have bold class', () => {
    const tablePlaced = [0, (1 << 6) | (1 << 7), 1 << 7, 1 << 7, 1 << 7];
    render(<SevensBoard {...defaultProps} tablePlaced={tablePlaced} />);
    const cells = screen.getAllByTestId('board-cell');
    // 6th cell in first suit row (index 5)
    expect(cells[5].className).not.toContain('font-bold');
  });

  it('placed cell has active background color and black text', () => {
    const tablePlaced = [0, (1 << 6) | (1 << 7), 1 << 7, 1 << 7, 1 << 7];
    render(<SevensBoard {...defaultProps} tablePlaced={tablePlaced} />);
    const cells = screen.getAllByTestId('board-cell');
    // Value 6 (index 5) is placed and non-center
    expect(cells[5].style.background).toBe('var(--color-game-status-active)');
    expect(cells[5].style.color).toBe('black');
  });

  it('center cell (7) has waiting background color and black text', () => {
    render(<SevensBoard {...defaultProps} />);
    const cells = screen.getAllByTestId('board-cell');
    expect(cells[6].style.background).toBe('var(--color-game-status-waiting)');
    expect(cells[6].style.color).toBe('black');
  });

  it('empty cell has transparent background', () => {
    render(<SevensBoard {...defaultProps} />);
    const cells = screen.getAllByTestId('board-cell');
    // Value 1 (index 0) is not placed — JSDOM normalizes rgba spaces
    expect(cells[0].style.background).toMatch(/rgba\(255,\s*255,\s*255,\s*0\.1\)/);
  });

  it('joker playable button renders with white text color', () => {
    render(<SevensBoard {...defaultProps} jokerSelecting={true} />);
    const button = screen.getByLabelText('SPADE 6 に配置') as HTMLElement;
    expect(button.style.color).toBe('white');
  });

  it('grid has horizontal scroll wrapper', () => {
    const { container } = render(<SevensBoard {...defaultProps} />);
    const scrollWrapper = container.querySelector('.overflow-x-auto');
    expect(scrollWrapper).toBeInTheDocument();
  });

  it('grid has minimum width for readability on mobile', () => {
    render(<SevensBoard {...defaultProps} />);
    const grid = screen.getByTestId('sevens-grid') as HTMLElement;
    expect(grid).toHaveStyle({ minWidth: '480px' });
  });

  it('renders ScrollFadeHint on mobile', () => {
    const innerWidthSpy = vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(375);
    try {
      const { container } = render(<SevensBoard {...defaultProps} />);
      const fadeHint = container.querySelector('.bg-gradient-to-l');
      expect(fadeHint).toBeInTheDocument();
    } finally {
      innerWidthSpy.mockRestore();
    }
  });

  it('does not render ScrollFadeHint on desktop', () => {
    const innerWidthSpy = vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(1024);
    try {
      const { container } = render(<SevensBoard {...defaultProps} />);
      const fadeHint = container.querySelector('.bg-gradient-to-l');
      expect(fadeHint).not.toBeInTheDocument();
    } finally {
      innerWidthSpy.mockRestore();
    }
  });
});
