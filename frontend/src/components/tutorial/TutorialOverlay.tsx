import { useCallback, useEffect, useRef, useState } from 'react';
import type { TutorialStep } from '../../types/tutorial';
import { getFocusableElements } from '../ConfirmDialog';
import { TutorialTooltip } from './TutorialTooltip';

/** Rect representing position and size of an element. */
interface SpotlightRect {
  top: number;
  left: number;
  width: number;
  height: number;
}

/** Padding around the spotlight cutout. */
const SPOTLIGHT_PADDING = 8;

/** Border radius for the spotlight cutout. */
const SPOTLIGHT_RADIUS = 8;

/** Props for the TutorialOverlay component. */
export interface TutorialOverlayProps {
  /** The current tutorial step definition. */
  step: TutorialStep;
  /** Zero-based index of the current step. */
  stepIndex: number;
  /** Total number of steps. */
  totalSteps: number;
  /** Called to advance to the next step. */
  onNext: () => void;
  /** Called to skip/dismiss the tutorial. */
  onSkip: () => void;
  /** Whether reduced motion is preferred. */
  reducedMotion: boolean;
}

/** Computes spotlight rect from a target element. */
function getSpotlightRect(el: Element | null): SpotlightRect | null {
  if (!el) return null;
  const rect = el.getBoundingClientRect();
  return {
    top: rect.top - SPOTLIGHT_PADDING,
    left: rect.left - SPOTLIGHT_PADDING,
    width: rect.width + SPOTLIGHT_PADDING * 2,
    height: rect.height + SPOTLIGHT_PADDING * 2,
  };
}

/** Computes tooltip position based on spotlight rect and placement. */
function getTooltipStyle(rect: SpotlightRect | null, placement: TutorialStep['placement']): React.CSSProperties {
  if (!rect) return { top: '50%', left: '50%', transform: 'translate(-50%, -50%)' };
  const gap = 12;
  switch (placement) {
    case 'top':
      return { top: rect.top - gap, left: rect.left + rect.width / 2, transform: 'translate(-50%, -100%)' };
    case 'bottom':
      return { top: rect.top + rect.height + gap, left: rect.left + rect.width / 2, transform: 'translateX(-50%)' };
    case 'left':
      return { top: rect.top + rect.height / 2, left: rect.left - gap, transform: 'translate(-100%, -50%)' };
    case 'right':
      return { top: rect.top + rect.height / 2, left: rect.left + rect.width + gap, transform: 'translateY(-50%)' };
  }
}

/** Renders a full-screen overlay with a spotlight cutout and tooltip for the tutorial. */
export function TutorialOverlay({ step, stepIndex, totalSteps, onNext, onSkip, reducedMotion }: TutorialOverlayProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<Element | null>(null);
  const [spotlightRect, setSpotlightRect] = useState<SpotlightRect | null>(null);

  // Find and observe the target element
  useEffect(() => {
    triggerRef.current = document.activeElement;
    const targetEl = document.querySelector(step.target);
    setSpotlightRect(getSpotlightRect(targetEl));

    if (!targetEl) return;

    const ro = new ResizeObserver(() => {
      setSpotlightRect(getSpotlightRect(targetEl));
    });
    ro.observe(targetEl);

    return () => {
      ro.disconnect();
      if (triggerRef.current instanceof HTMLElement) {
        triggerRef.current.focus();
      }
    };
  }, [step.target]);

  // Listen for click on target when advanceOn is 'click'
  useEffect(() => {
    if (step.advanceOn !== 'click') return;
    const targetEl = document.querySelector(step.target);
    if (!targetEl) return;

    const handler = () => onNext();
    targetEl.addEventListener('click', handler);
    return () => targetEl.removeEventListener('click', handler);
  }, [step.target, step.advanceOn, onNext]);

  // Focus trap
  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    const focusable = getFocusableElements(dialog);
    if (focusable.length > 0) {
      focusable[0].focus();
    }

    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      const currentFocusable = getFocusableElements(dialog);
      if (currentFocusable.length === 0) return;
      const first = currentFocusable[0];
      const last = currentFocusable[currentFocusable.length - 1];
      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    dialog.addEventListener('keydown', handleTab);
    return () => dialog.removeEventListener('keydown', handleTab);
  }, []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') onNext();
      if (e.key === 'Escape') onSkip();
    },
    [onNext, onSkip],
  );

  const tooltipStyle = getTooltipStyle(spotlightRect, step.placement);
  const transitionStyle = reducedMotion ? {} : { transition: 'opacity 0.2s ease-in-out' };

  return (
    <div
      ref={dialogRef}
      role="dialog"
      aria-modal="true"
      aria-label="Tutorial"
      className="fixed inset-0 z-50"
      onKeyDown={handleKeyDown}
      style={transitionStyle}
    >
      {/* SVG overlay with spotlight cutout */}
      <svg className="absolute inset-0 w-full h-full pointer-events-none" aria-hidden="true">
        <defs>
          <mask id={`tutorial-mask-${stepIndex}`}>
            <rect width="100%" height="100%" fill="white" />
            {spotlightRect && (
              <rect
                x={spotlightRect.left}
                y={spotlightRect.top}
                width={spotlightRect.width}
                height={spotlightRect.height}
                rx={SPOTLIGHT_RADIUS}
                ry={SPOTLIGHT_RADIUS}
                fill="black"
              />
            )}
          </mask>
        </defs>
        <rect width="100%" height="100%" fill="rgba(0,0,0,0.6)" mask={`url(#tutorial-mask-${stepIndex})`} />
      </svg>

      {/* Tooltip positioned relative to spotlight */}
      <div className="absolute z-10" style={tooltipStyle}>
        <TutorialTooltip
          message={step.messageKey}
          placement={step.placement}
          stepIndex={stepIndex}
          totalSteps={totalSteps}
          onNext={onNext}
          onSkip={onSkip}
          advanceOn={step.advanceOn}
        />
      </div>
    </div>
  );
}
