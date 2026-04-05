import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { TutorialProvider } from '../../providers/TutorialProvider';
import type { TutorialStep } from '../../types/tutorial';

/** Props for the TutorialWrapper component. */
export interface TutorialWrapperProps {
  /** Game identifier used for i18n namespace and tutorial localStorage key. */
  gameName: string;
  /** Ordered list of tutorial steps for this game. */
  steps: TutorialStep[];
  /** Child elements rendered inside the tutorial context. */
  children: ReactNode;
}

/** Wraps a game page with TutorialProvider, providing i18n translation automatically. */
export function TutorialWrapper({ gameName, steps, children }: TutorialWrapperProps) {
  const { t } = useTranslation(gameName);
  return (
    <TutorialProvider config={{ gameName, steps }} translateMessage={t}>
      {children}
    </TutorialProvider>
  );
}
