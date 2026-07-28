import { motion } from 'framer-motion';
import { useRef } from 'react';
import { useReducedMotion } from '../../hooks/useReducedMotion';
import { useOptionalSound } from '../../providers/SoundProvider';
import { dealSpring, hoverLift, selectLift } from '../../styles/motionPresets';
import { CardImage } from '../CardImage';

/** Props for {@link AnimatedCard}. */
export interface AnimatedCardProps extends React.ComponentProps<typeof CardImage> {
  /** Stagger delay in seconds for deal animation. */
  dealDelay?: number;
  /** Whether this card is selected (lift + glow). */
  isSelected?: boolean;
  /** Shared layout animation ID. */
  layoutId?: string;
  /** Additional class name for the motion wrapper div. */
  wrapperClassName?: string;
  /**
   * Suppress the default 'cardDeal' SFX. Useful for silent re-renders
   * or when a parent already plays a louder placement sound.
   */
  silent?: boolean;
  /**
   * Optional callback fired after the deal-in animation completes,
   * in addition to the default SFX. Most call sites can omit this;
   * pass one only when extra side-effects (e.g., onPlace plumbing in
   * AnimatedPile) need to fire alongside the default sound.
   */
  onDealComplete?: () => void;
}

/** Renders an animated face-up playing card with deal and select animations. */
export function AnimatedCard({
  dealDelay = 0,
  isSelected = false,
  layoutId,
  wrapperClassName,
  silent = false,
  onDealComplete,
  ...rest
}: AnimatedCardProps) {
  const reduced = useReducedMotion();
  const sound = useOptionalSound();
  // Guard: onAnimationComplete fires for any animation (hover, selection), not just the initial deal.
  // Use a ref so the deal callback fires exactly once per component instance.
  const dealCalledRef = useRef(false);

  if (reduced) {
    // Reduced motion means no movement, not no feedback. This branch used to return a
    // bare CardImage, so anyone with `prefers-reduced-motion: reduce` got no hover
    // affordance on cards at all while everyone else got a lift — the gap #930 asked
    // about. A brightness/shadow change is a state indication rather than an
    // animation, so it is allowed here (and the universal reduced-motion block in
    // index.css strips any transition on it anyway).
    return (
      <span
        className={wrapperClassName ?? 'inline-block'}
        data-testid="animated-card"
        style={wrapperClassName ? undefined : { display: 'inline-block' }}
      >
        <span className="block hover:brightness-110 hover:shadow-lg hover:shadow-black/40">
          <CardImage {...rest} />
        </span>
      </span>
    );
  }

  return (
    <motion.div
      layoutId={layoutId}
      className={wrapperClassName}
      initial={{ opacity: 0, y: 30 }}
      animate={{
        opacity: 1,
        y: isSelected ? selectLift.y : 0,
        scale: isSelected ? selectLift.scale : 1,
      }}
      whileHover={hoverLift}
      transition={{
        ...dealSpring,
        delay: dealDelay,
      }}
      onAnimationComplete={() => {
        if (!dealCalledRef.current) {
          dealCalledRef.current = true;
          if (!silent) sound?.playSound('cardDeal', { pitchVariation: 0.03 });
          onDealComplete?.();
        }
      }}
      style={wrapperClassName ? undefined : { display: 'inline-block' }}
      data-testid="animated-card"
    >
      <CardImage {...rest} />
    </motion.div>
  );
}
