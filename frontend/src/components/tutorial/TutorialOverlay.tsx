import { useCallback, useEffect, useId, useLayoutEffect, useRef, useState } from 'react';
import { useBodyScrollLock } from '../../hooks/useBodyScrollLock';
import { useFocusTrap } from '../../hooks/useFocusTrap';
import type { TutorialStep } from '../../types/tutorial';
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

/** Minimum distance between tooltip and viewport edge in pixels. */
const TOOLTIP_VIEWPORT_MARGIN = 8;

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
  const maskId = useId();
  const dialogRef = useRef<HTMLDivElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);
  const [spotlightRect, setSpotlightRect] = useState<SpotlightRect | null>(null);

  useBodyScrollLock(true);

  // Track the spotlight target. This re-runs on every step, so it must not
  // touch focus: restoring focus here fired on each step change and threw it
  // out of the overlay for the rest of the tutorial (issue #5184). Focus is
  // owned by useFocusTrap below, whose deps are stable so it neither re-steals
  // focus per step nor hands it back before the tutorial ends.
  useEffect(() => {
    const targetEl = document.querySelector(step.target);
    setSpotlightRect(getSpotlightRect(targetEl));

    if (!targetEl) return;

    const ro = new ResizeObserver(() => {
      setSpotlightRect(getSpotlightRect(targetEl));
    });
    ro.observe(targetEl);

    return () => ro.disconnect();
  }, [step.target]);

  // Listen for click on target when advanceOn is 'click'.
  // Note: the target must be within the spotlight cutout for clicks to reach it,
  // as the overlay dialog intercepts clicks outside the cutout area.
  useEffect(() => {
    if (step.advanceOn !== 'click') return;
    const targetEl = document.querySelector(step.target);
    if (!targetEl) return;

    const handler = () => onNext();
    targetEl.addEventListener('click', handler);
    return () => targetEl.removeEventListener('click', handler);
  }, [step.target, step.advanceOn, onNext]);

  // Focus trap, Escape, and focus restore, from the shared hook: it listens on
  // `document` (so it keeps working if focus ever does leave) and re-queries
  // focusable children on each Tab, unlike the copy this replaces.
  useFocusTrap(dialogRef, true, onSkip);

  // Escape is handled by useFocusTrap. Handling it here as well would call
  // onSkip twice for a single keypress, since this synthetic handler and the
  // hook's document listener both see the same event.
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') onNext();
    },
    [onNext],
  );

  const tooltipStyle = getTooltipStyle(spotlightRect, step.placement);
  const transitionStyle = reducedMotion ? {} : { transition: 'opacity 0.2s ease-in-out' };

  // Clamp tooltip within viewport after render — depends on tooltipStyle which
  // is derived from spotlightRect and step.placement, so it re-runs when they change.
  // biome-ignore lint/correctness/useExhaustiveDependencies: tooltipStyle captures spotlightRect+placement changes
  useLayoutEffect(() => {
    const el = tooltipRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    let clamped = false;
    let newLeft = Number.parseFloat(el.style.left) || 0;
    let newTop = Number.parseFloat(el.style.top) || 0;

    if (rect.left < TOOLTIP_VIEWPORT_MARGIN) {
      newLeft = TOOLTIP_VIEWPORT_MARGIN;
      clamped = true;
    } else if (rect.right > window.innerWidth - TOOLTIP_VIEWPORT_MARGIN) {
      newLeft = window.innerWidth - TOOLTIP_VIEWPORT_MARGIN - rect.width;
      clamped = true;
    }

    if (rect.top < TOOLTIP_VIEWPORT_MARGIN) {
      newTop = TOOLTIP_VIEWPORT_MARGIN;
      clamped = true;
    } else if (rect.bottom > window.innerHeight - TOOLTIP_VIEWPORT_MARGIN) {
      newTop = window.innerHeight - TOOLTIP_VIEWPORT_MARGIN - rect.height;
      clamped = true;
    }

    if (clamped) {
      el.style.left = `${newLeft}px`;
      el.style.top = `${newTop}px`;
      el.style.transform = 'none';
    }
  }, [tooltipStyle]);

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
          <mask id={`tutorial-mask-${maskId}`}>
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
        <rect width="100%" height="100%" fill="rgba(0,0,0,0.75)" mask={`url(#tutorial-mask-${maskId})`} />
      </svg>

      {/* Tooltip positioned relative to spotlight */}
      <div ref={tooltipRef} className="absolute z-10" style={tooltipStyle}>
        <TutorialTooltip
          message={step.messageKey}
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
