import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ActionLogEntry } from '../types/card';
import { ActionLogSection } from './ActionLogSection';

vi.mock('./ActionLogPanel', () => ({
  ActionLogPanel: ({ entries, onClose }: { entries: ActionLogEntry[]; onClose: () => void }) => (
    <div data-testid="action-log-panel">
      <span>{entries.length} entries</span>
      <button type="button" onClick={onClose}>
        close
      </button>
    </div>
  ),
}));

const entry: ActionLogEntry = { turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'test' };

describe('ActionLogSection', () => {
  it('renders nothing when isEndPhase=false and actionLog=null', () => {
    const { container } = render(
      <ActionLogSection isEndPhase={false} actionLog={null} showActionLog={vi.fn()} hideActionLog={vi.fn()} />,
    );
    expect(container.innerHTML).toBe('');
  });

  it('renders button when isEndPhase=true and actionLog=null', () => {
    const showActionLog = vi.fn();
    render(
      <ActionLogSection isEndPhase={true} actionLog={null} showActionLog={showActionLog} hideActionLog={vi.fn()} />,
    );
    expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument();
    expect(screen.queryByTestId('action-log-panel')).not.toBeInTheDocument();
  });

  it('calls showActionLog on button click', () => {
    const showActionLog = vi.fn();
    render(
      <ActionLogSection isEndPhase={true} actionLog={null} showActionLog={showActionLog} hideActionLog={vi.fn()} />,
    );
    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    expect(showActionLog).toHaveBeenCalledTimes(1);
  });

  it('renders ActionLogPanel when isEndPhase=true and actionLog is provided', () => {
    render(<ActionLogSection isEndPhase={true} actionLog={[entry]} showActionLog={vi.fn()} hideActionLog={vi.fn()} />);
    expect(screen.getByTestId('action-log-panel')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '棋譜を見る' })).not.toBeInTheDocument();
  });

  it('renders ActionLogPanel when isEndPhase=false and actionLog is provided', () => {
    const hideActionLog = vi.fn();
    render(
      <ActionLogSection isEndPhase={false} actionLog={[entry]} showActionLog={vi.fn()} hideActionLog={hideActionLog} />,
    );
    expect(screen.getByTestId('action-log-panel')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '棋譜を見る' })).not.toBeInTheDocument();
  });
});
