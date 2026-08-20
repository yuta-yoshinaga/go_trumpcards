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
  /**
   * Interpolation values for `messageKey`.
   *
   * **数値を文言に焼き込まない**ための口。焼き込むと、ドメイン定数を変えたとき
   * 盤面の表示だけが追随し、チュートリアルは古い数字を教え続ける (#5936)。
   * 応答が値を運んでいる場合は、静的に書かずに
   * `useTutorialMessageParams` で毎回の state から渡す。
   */
  messageParams?: Record<string, string | number>;
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
