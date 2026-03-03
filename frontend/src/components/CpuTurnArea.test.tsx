import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CpuTurnArea } from './CpuTurnArea';

const baseClass = 'bg-black/35 rounded-[10px] border-2 border-transparent p-[10px]';

describe('CpuTurnArea', () => {
  it('renders player name for CPU', () => {
    render(<CpuTurnArea playerId={2} isHuman={false} isCurrentTurn={false} isFinished={false} className={baseClass} />);
    expect(screen.getByText('CPU 2')).toBeInTheDocument();
  });

  it('renders player name for human', () => {
    render(<CpuTurnArea playerId={0} isHuman={true} isCurrentTurn={false} isFinished={false} className={baseClass} />);
    expect(screen.getByText('あなた')).toBeInTheDocument();
  });

  it('shows thinking badge when current turn and not finished', () => {
    render(<CpuTurnArea playerId={1} isHuman={false} isCurrentTurn={true} isFinished={false} className={baseClass} />);
    expect(screen.getByText('考え中...')).toBeInTheDocument();
  });

  it('does not show thinking badge when finished', () => {
    render(
      <CpuTurnArea
        playerId={1}
        isHuman={false}
        isCurrentTurn={true}
        isFinished={true}
        finishedLabel="上がり"
        className={baseClass}
      />,
    );
    expect(screen.queryByText('考え中...')).not.toBeInTheDocument();
    expect(screen.getByText('上がり')).toBeInTheDocument();
  });

  it('shows finished label when finished', () => {
    render(
      <CpuTurnArea
        playerId={1}
        isHuman={false}
        isCurrentTurn={false}
        isFinished={true}
        finishedLabel="上がり (大富豪)"
        className={baseClass}
      />,
    );
    expect(screen.getByText('上がり (大富豪)')).toBeInTheDocument();
  });

  it('does not show finished label when finishedLabel is not provided', () => {
    render(<CpuTurnArea playerId={1} isHuman={false} isCurrentTurn={false} isFinished={true} className={baseClass} />);
    // No badge at all
    expect(screen.queryByText('上がり')).not.toBeInTheDocument();
    expect(screen.queryByText('考え中...')).not.toBeInTheDocument();
  });

  it('applies dimmed style when finished and dimFinished is true (default)', () => {
    const { container } = render(
      <CpuTurnArea
        playerId={1}
        isHuman={false}
        isCurrentTurn={false}
        isFinished={true}
        finishedLabel="上がり"
        className={baseClass}
      />,
    );
    expect(container.firstChild).toHaveStyle({ opacity: 0.5 });
  });

  it('does not apply dimmed style when dimFinished is false', () => {
    const { container } = render(
      <CpuTurnArea
        playerId={1}
        isHuman={false}
        isCurrentTurn={false}
        isFinished={true}
        dimFinished={false}
        finishedLabel="上がり"
        className={baseClass}
      />,
    );
    expect(container.firstChild).not.toHaveStyle({ opacity: 0.5 });
  });

  it('applies active turn style when current turn', () => {
    const { container } = render(
      <CpuTurnArea playerId={1} isHuman={false} isCurrentTurn={true} isFinished={false} className={baseClass} />,
    );
    const el = container.firstChild as HTMLElement;
    expect(el.style.border).toContain('2px solid');
    expect(el.style.boxShadow).toContain('0 0 12px');
  });

  it('applies no conditional style when idle', () => {
    const { container } = render(
      <CpuTurnArea playerId={1} isHuman={false} isCurrentTurn={false} isFinished={false} className={baseClass} />,
    );
    expect(container.firstChild).not.toHaveStyle({ opacity: 0.5 });
    expect(container.firstChild).not.toHaveStyle({ border: '2px solid #f0ad4e' });
  });

  it('passes id prop to the outer div', () => {
    const { container } = render(
      <CpuTurnArea
        id="player-area-1"
        playerId={1}
        isHuman={false}
        isCurrentTurn={false}
        isFinished={false}
        className={baseClass}
      />,
    );
    expect(container.querySelector('#player-area-1')).toBeInTheDocument();
  });

  it('renders children', () => {
    render(
      <CpuTurnArea playerId={1} isHuman={false} isCurrentTurn={false} isFinished={false} className={baseClass}>
        <span>5枚</span>
      </CpuTurnArea>,
    );
    expect(screen.getByText('5枚')).toBeInTheDocument();
  });

  it('applies nameClassName to name div', () => {
    const { container } = render(
      <CpuTurnArea
        playerId={1}
        isHuman={false}
        isCurrentTurn={false}
        isFinished={false}
        className={baseClass}
        nameClassName="text-sm"
      />,
    );
    const nameDiv = container.querySelector('.text-sm');
    expect(nameDiv).toBeInTheDocument();
    expect(nameDiv?.textContent).toContain('CPU 1');
  });

  it('does not add extra class when nameClassName is not provided', () => {
    const { container } = render(
      <CpuTurnArea playerId={1} isHuman={false} isCurrentTurn={false} isFinished={false} className={baseClass} />,
    );
    const nameDiv = container.querySelector('.font-bold');
    expect(nameDiv?.className).toBe('text-white font-bold mb-1');
  });
});
