import { fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
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

  it('calls hideActionLog when panel close button is clicked', () => {
    const hideActionLog = vi.fn();
    render(
      <ActionLogSection isEndPhase={false} actionLog={[entry]} showActionLog={vi.fn()} hideActionLog={hideActionLog} />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'close' }));
    expect(hideActionLog).toHaveBeenCalledTimes(1);
  });

  // Opening the panel unmounts the trigger (it is gated on `!actionLog`), and
  // closing remounts it as a different element. ActionLogPanel's own restore
  // therefore cannot reach it, so this component owns the restore. Driven
  // through real state rather than fixed props, because the bug only exists
  // across the open -> close transition. See issue #5183.
  it('returns focus to the trigger after the panel closes', () => {
    function Harness() {
      const [log, setLog] = useState<ActionLogEntry[] | null>(null);
      return (
        <ActionLogSection
          isEndPhase={true}
          actionLog={log}
          showActionLog={() => setLog([entry])}
          hideActionLog={() => setLog(null)}
        />
      );
    }
    render(<Harness />);

    const trigger = screen.getByRole('button', { name: '棋譜を見る' });
    trigger.focus();
    fireEvent.click(trigger);

    // The trigger really is gone while the panel is open — this is what makes
    // the panel-side restore insufficient.
    expect(trigger.isConnected).toBe(false);

    fireEvent.click(screen.getByRole('button', { name: 'close' }));

    const remounted = screen.getByRole('button', { name: '棋譜を見る' });
    expect(remounted).not.toBe(trigger);
    expect(document.activeElement).toBe(remounted);
  });

  it('does not steal focus on the initial render', () => {
    const outside = document.createElement('button');
    document.body.appendChild(outside);
    outside.focus();

    render(<ActionLogSection isEndPhase={true} actionLog={null} showActionLog={vi.fn()} hideActionLog={vi.fn()} />);

    expect(document.activeElement).toBe(outside);
    document.body.removeChild(outside);
  });
});
