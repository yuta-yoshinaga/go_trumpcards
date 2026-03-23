import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TutorialConfig } from '../types/tutorial';
import { TutorialProvider, useTutorialContext } from './TutorialProvider';

// Mock ResizeObserver
class MockResizeObserver {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
  callback: ResizeObserverCallback;
  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
  }
}
vi.stubGlobal('ResizeObserver', MockResizeObserver);

const config: TutorialConfig = {
  gameName: 'testgame',
  steps: [
    { target: '[data-tutorial="step1"]', messageKey: 'ステップ1の説明', placement: 'bottom', advanceOn: 'next' },
    { target: '[data-tutorial="step2"]', messageKey: 'ステップ2の説明', placement: 'top', advanceOn: 'next' },
  ],
};

function TestConsumer() {
  const ctx = useTutorialContext();
  return (
    <div>
      <span data-testid="active">{String(ctx.isActive)}</span>
      <span data-testid="step">{ctx.currentStepIndex}</span>
      <span data-testid="completed">{String(ctx.isCompleted)}</span>
      <button type="button" onClick={ctx.start}>
        Start
      </button>
      <button type="button" onClick={ctx.next}>
        Next
      </button>
      <button type="button" onClick={ctx.skip}>
        Skip
      </button>
    </div>
  );
}

describe('TutorialProvider', () => {
  let targetEl: HTMLDivElement;

  beforeEach(() => {
    targetEl = document.createElement('div');
    targetEl.setAttribute('data-tutorial', 'step1');
    targetEl.getBoundingClientRect = vi.fn().mockReturnValue({
      top: 100,
      left: 50,
      width: 200,
      height: 40,
      right: 250,
      bottom: 140,
    });
    document.body.appendChild(targetEl);
  });

  afterEach(() => {
    if (document.body.contains(targetEl)) {
      document.body.removeChild(targetEl);
    }
    localStorage.clear();
  });

  it('provides tutorial context to children', () => {
    render(
      <TutorialProvider config={config}>
        <TestConsumer />
      </TutorialProvider>,
    );
    expect(screen.getByTestId('active')).toHaveTextContent('false');
    expect(screen.getByTestId('step')).toHaveTextContent('0');
  });

  it('starts tutorial and shows overlay', async () => {
    render(
      <TutorialProvider config={config}>
        <TestConsumer />
      </TutorialProvider>,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Start' }));
    await waitFor(() => {
      expect(screen.getByTestId('active')).toHaveTextContent('true');
    });
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText('ステップ1の説明')).toBeInTheDocument();
  });

  it('does not show overlay when tutorial is inactive', () => {
    render(
      <TutorialProvider config={config}>
        <TestConsumer />
      </TutorialProvider>,
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('advances steps via context', async () => {
    render(
      <TutorialProvider config={config}>
        <TestConsumer />
      </TutorialProvider>,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Start' }));
    await waitFor(() => {
      expect(screen.getByTestId('active')).toHaveTextContent('true');
    });
    // Advance using context next
    act(() => {
      fireEvent.click(screen.getByRole('button', { name: 'Next' }));
    });
    await waitFor(() => {
      expect(screen.getByTestId('step')).toHaveTextContent('1');
    });
  });

  it('completes tutorial after last step', async () => {
    render(
      <TutorialProvider config={config}>
        <TestConsumer />
      </TutorialProvider>,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Start' }));
    await waitFor(() => expect(screen.getByTestId('active')).toHaveTextContent('true'));
    // Advance through all steps
    act(() => fireEvent.click(screen.getByRole('button', { name: 'Next' })));
    act(() => fireEvent.click(screen.getByRole('button', { name: 'Next' })));
    await waitFor(() => {
      expect(screen.getByTestId('active')).toHaveTextContent('false');
      expect(screen.getByTestId('completed')).toHaveTextContent('true');
    });
  });

  it('skips tutorial via context', async () => {
    render(
      <TutorialProvider config={config}>
        <TestConsumer />
      </TutorialProvider>,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Start' }));
    await waitFor(() => expect(screen.getByTestId('active')).toHaveTextContent('true'));
    fireEvent.click(screen.getByRole('button', { name: 'Skip' }));
    await waitFor(() => {
      expect(screen.getByTestId('active')).toHaveTextContent('false');
      expect(screen.getByTestId('completed')).toHaveTextContent('false');
    });
  });

  it('throws when useTutorialContext is used outside provider', () => {
    // Suppress console.error for expected error
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() => render(<TestConsumer />)).toThrow('useTutorialContext must be used within a TutorialProvider');
    spy.mockRestore();
  });
});
