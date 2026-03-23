import { createContext, type ReactNode, useContext } from 'react';
import { TutorialOverlay } from '../components/tutorial/TutorialOverlay';
import { useReducedMotion } from '../hooks/useReducedMotion';
import { type UseTutorialReturn, useTutorial } from '../hooks/useTutorial';
import type { TutorialConfig } from '../types/tutorial';

const TutorialContext = createContext<UseTutorialReturn | null>(null);

/** Accesses the tutorial context. Must be used within a TutorialProvider. */
export function useTutorialContext(): UseTutorialReturn {
  const ctx = useContext(TutorialContext);
  if (!ctx) throw new Error('useTutorialContext must be used within a TutorialProvider');
  return ctx;
}

/** Props for the TutorialProvider component. */
export interface TutorialProviderProps {
  /** Tutorial configuration with game name and steps. */
  config: TutorialConfig;
  /** Optional function to translate step messageKey values. Defaults to identity. */
  translateMessage?: (key: string) => string;
  /** Child elements that can access the tutorial context. */
  children: ReactNode;
}

/** Identity function used as default translateMessage. */
const identity = (key: string) => key;

/** Provides tutorial state and renders the overlay when the tutorial is active. */
export function TutorialProvider({ config, translateMessage = identity, children }: TutorialProviderProps) {
  const tutorial = useTutorial(config);
  const reducedMotion = useReducedMotion();

  const currentStep = tutorial.currentStep;
  const translatedStep = currentStep ? { ...currentStep, messageKey: translateMessage(currentStep.messageKey) } : null;

  return (
    <TutorialContext.Provider value={tutorial}>
      {children}
      {tutorial.isActive && translatedStep && (
        <TutorialOverlay
          step={translatedStep}
          stepIndex={tutorial.currentStepIndex}
          totalSteps={tutorial.totalSteps}
          onNext={tutorial.next}
          onSkip={tutorial.skip}
          reducedMotion={reducedMotion}
        />
      )}
    </TutorialContext.Provider>
  );
}
