import { render, screen } from '@testing-library/react';
import type { ComponentType } from 'react';
import { describe, expect, it } from 'vitest';
import type { TutorialStep } from '../../types/tutorial';
import { withTutorial } from './withTutorial';

const steps: TutorialStep[] = [
  { target: '[data-tutorial="x"]', messageKey: 'tutorial.step1', placement: 'top', advanceOn: 'next' },
];

function Inner({ label }: { label: string }) {
  return <div data-testid="inner">inner-{label}</div>;
}

describe('withTutorial', () => {
  it('renders the wrapped component inside TutorialWrapper', () => {
    const Wrapped = withTutorial(Inner, 'common', steps);
    render(<Wrapped label="hello" />);
    expect(screen.getByTestId('inner')).toHaveTextContent('inner-hello');
  });

  it('sets a debugging-friendly displayName', () => {
    function Named() {
      return null;
    }
    const Wrapped = withTutorial(Named, 'common', steps);
    expect(Wrapped.displayName).toBe('withTutorial(Named)');
  });

  it('falls back to "Component" in displayName for anonymous components', () => {
    const Anonymous = (() => null) as ComponentType<unknown>;
    Object.defineProperty(Anonymous, 'name', { value: '' });
    const Wrapped = withTutorial(Anonymous, 'common', steps);
    expect(Wrapped.displayName).toBe('withTutorial(Component)');
  });
});
