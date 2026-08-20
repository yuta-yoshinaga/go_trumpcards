import { createContext, type ReactNode, useCallback, useContext, useEffect, useState } from 'react';
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
  translateMessage?: (key: string, params?: Record<string, string | number>) => string;
  /** Child elements that can access the tutorial context. */
  children: ReactNode;
}

/** Identity function used as default translateMessage. */
const identity = (key: string) => key;

/** Interpolation values a page feeds into the current step's message. */
export type TutorialMessageParams = Record<string, string | number>;

const TutorialMessageParamsContext = createContext<((params: TutorialMessageParams) => void) | null>(null);

/**
 * Feeds interpolation values into the running tutorial's messages.
 *
 * Steps are module-level constants, so a number the **response** carries
 * (a bonus, an ante) cannot be written into them. Call this from the page with
 * the value it already renders, and the tutorial says the same number the board
 * does (#5936).
 *
 * @param params - Values for the current step's `{{placeholders}}`.
 */
export function useTutorialMessageParams(params: TutorialMessageParams): void {
  const setParams = useContext(TutorialMessageParamsContext);
  // **値で比較する。**呼び出し側はほぼ必ずリテラルを渡すので、参照で比較すると
  // 毎レンダリング更新して無限ループになる。
  const serialized = JSON.stringify(params);
  useEffect(() => {
    setParams?.(JSON.parse(serialized) as TutorialMessageParams);
  }, [serialized, setParams]);
}

/** Provides tutorial state, first-visit suggestion, and renders the overlay when active. */
export function TutorialProvider({ config, translateMessage = identity, children }: TutorialProviderProps) {
  const tutorial = useTutorial(config);
  const reducedMotion = useReducedMotion();
  const { shouldShowDialog, dismiss, dismissPermanently } = useFirstVisit(config.gameName);
  const [dontShowAgain, setDontShowAgain] = useState(false);

  const dismissDialog = useCallback(() => {
    if (dontShowAgain) {
      dismissPermanently();
    } else {
      dismiss();
    }
  }, [dontShowAgain, dismiss, dismissPermanently]);

  const handleStartTutorial = useCallback(() => {
    dismissDialog();
    tutorial.start();
  }, [dismissDialog, tutorial]);

  const handleSkip = useCallback(() => {
    dismissDialog();
  }, [dismissDialog]);

  const [dynamicParams, setDynamicParams] = useState<TutorialMessageParams>({});

  const currentStep = tutorial.currentStep;
  const translatedStep = currentStep
    ? {
        ...currentStep,
        // ページが渡した値が後勝ち。ステップ側の値は既定にすぎない。
        messageKey: translateMessage(currentStep.messageKey, { ...currentStep.messageParams, ...dynamicParams }),
      }
    : null;

  return (
    <TutorialContext.Provider value={tutorial}>
      <TutorialMessageParamsContext.Provider value={setDynamicParams}>{children}</TutorialMessageParamsContext.Provider>
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
