/** Spring transition for card deal animations. Settles in ~300ms (DESIGN.md medium token). */
export const dealSpring = { type: 'spring' as const, stiffness: 300, damping: 25 };

/** Spring transition for card flip animations. Settles in ~200ms (DESIGN.md short token). */
export const flipSpring = { type: 'spring' as const, stiffness: 400, damping: 30 };

/** Animate target for selected card state: lift -8px + scale 1.02. */
export const selectLift = { y: -8, scale: 1.02 };

/** Animate target for card hover state: lift -4px + scale 1.05. */
export const hoverLift = { y: -4, scale: 1.05 };

/** Parent variant for staggered children animations (120ms between cards). */
export const staggerTiming = { staggerChildren: 0.12 };

/** Spring transition for neighbor expansion animation on mobile card tap. */
export const expansionSpring = { type: 'spring' as const, stiffness: 350, damping: 28 };

/** Extra gap (px) added to each neighbor of a selected card on mobile. */
export const EXPANSION_GAP_PX = 12;
