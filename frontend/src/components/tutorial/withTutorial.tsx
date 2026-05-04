import type { ComponentType } from 'react';
import type { TutorialStep } from '../../types/tutorial';
import { TutorialWrapper } from './TutorialWrapper';

/**
 * Wraps a page-level component with `TutorialWrapper`, eliminating the
 * `function FooPage() { return <TutorialWrapper>...</TutorialWrapper> }`
 * boilerplate that every game page repeats.
 *
 * Usage in a game page:
 *
 *   function BlackJackPageContent() { ... }
 *   export const BlackJackPage = withTutorial(BlackJackPageContent, 'blackjack', BJ_TUTORIAL_STEPS);
 */
export function withTutorial<P extends object>(
  Component: ComponentType<P>,
  gameName: string,
  steps: TutorialStep[],
): ComponentType<P> {
  const Wrapped = (props: P) => (
    <TutorialWrapper gameName={gameName} steps={steps}>
      <Component {...props} />
    </TutorialWrapper>
  );
  // Treat empty strings as missing too — anonymous arrow components have
  // `name === ""`, which `??` would otherwise prefer over the literal fallback.
  const label = Component.displayName || Component.name || 'Component';
  Wrapped.displayName = `withTutorial(${label})`;
  return Wrapped;
}
