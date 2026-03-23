import { createContext, type ReactNode, useCallback, useContext, useState } from 'react';
import { TutorialOverlay } from '../components/tutorial/TutorialOverlay';
import { TutorialSuggestDialog } from '../components/tutorial/TutorialSuggestDialog';
import { useFirstVisit } from '../hooks/useFirstVisit';
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

/** Provides tutorial state, first-visit suggestion, and renders the overlay when active. */
export function TutorialProvider({ config, translateMessage = identity, children }: TutorialProviderProps) {
  const tutorial = useTutorial(config);
  const reducedMotion = useReducedMotion();
  const { shouldShowDialog, dismiss, dismissPermanently } = useFirstVisit(config.gameName);
  const [dontShowAgain, setDontShowAgain] = useState(false);

  const handleStartTutorial = useCallback(() => {
    if (dontShowAgain) {
      dismissPermanently();
    } else {
      dismiss();
    }
    tutorial.start();
  }, [dontShowAgain, dismiss, dismissPermanently, tutorial]);

  const handleSkip = useCallback(() => {
    if (dontShowAgain) {
      dismissPermanently();
    } else {
      dismiss();
    }
  }, [dontShowAgain, dismiss, dismissPermanently]);

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
      {shouldShowDialog && !tutorial.isActive && (
        <TutorialSuggestDialog
          open={true}
          onStartTutorial={handleStartTutorial}
          onSkip={handleSkip}
          dontShowAgain={dontShowAgain}
          onDontShowAgainChange={setDontShowAgain}
        />
      )}
    </TutorialContext.Provider>
  );
}
