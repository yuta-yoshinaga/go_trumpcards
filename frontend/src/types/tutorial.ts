/** Placement direction for a tutorial tooltip relative to the target element. */
export type TutorialPlacement = 'top' | 'bottom' | 'left' | 'right';

/** How the tutorial advances to the next step. */
export type TutorialAdvanceOn = 'click' | 'next';

/** Defines a single step in a game tutorial sequence. */
export interface TutorialStep {
  /** CSS selector or data-tutorial attribute value to highlight. */
  target: string;
  /** i18n key for the tooltip message text. */
  messageKey: string;
  /** Tooltip placement relative to the target element. */
  placement: TutorialPlacement;
  /** How to advance: 'click' watches the target, 'next' shows a next button. */
  advanceOn: TutorialAdvanceOn;
  /** Optional callback executed when this step becomes active. */
  onEnter?: () => void;
}

/** Configuration for a game's tutorial sequence. */
export interface TutorialConfig {
  /** Game identifier used for localStorage key and i18n namespace. */
  gameName: string;
  /** Ordered list of tutorial steps. */
  steps: TutorialStep[];
}
