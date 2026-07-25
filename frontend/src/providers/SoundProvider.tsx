import { Howl, Howler } from 'howler';
import type { ReactNode } from 'react';
import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';

const SOUND_MUTED_KEY = 'trumpcards-sound-muted';

/** Sound name identifier for all game sounds. */
export type SoundName =
  | 'cardDeal'
  | 'cardFlip'
  | 'cardSelect'
  | 'cardPlace'
  | 'shuffle'
  | 'chipClick'
  | 'winFanfare'
  | 'lossThud'
  | 'errorBuzz'
  | 'turnTick';

/** Options for playSound calls. */
export interface PlayOptions {
  /** Pitch randomization range. E.g., 0.05 = playbackRate 0.95-1.05. */
  pitchVariation?: number;
  /** Volume override, 0.0-1.0. Overrides the default tier volume. */
  volume?: number;
}

/** Context value provided by SoundProvider. */
interface SoundContextValue {
  /** Fire-and-forget sound playback. Silently no-ops if muted or sound fails. */
  playSound: (name: SoundName, options?: PlayOptions) => void;
  /** Whether sound is currently muted. */
  muted: boolean;
  /** Toggle mute state. Persists to localStorage. */
  toggleMute: () => void;
  /**
   * Marks the next exec resolution's generic sound as already covered by a
   * more specific action sound (e.g., chipClick on a bet). The claim is
   * consumed by the next `consumeExecClaim` call or expires after
   * {@link CLAIM_EXPIRY_MS}, whichever comes first.
   */
  claimExecSound: () => void;
  /**
   * Consumes a pending exec-sound claim. Returns true when a live claim
   * existed (the caller should skip its generic exec sound), false otherwise.
   */
  consumeExecClaim: () => boolean;
}

const SoundContext = createContext<SoundContextValue | null>(null);

/** Base volume for all sounds. */
const BASE_VOLUME = 0.3;

/** Volume multiplier per tier: ambient (background), action (user feedback), event (state change). */
const VOLUME_TIERS: Record<SoundName, number> = {
  cardDeal: 0.6,
  cardPlace: 0.6,
  turnTick: 0.6,
  cardSelect: 0.6,
  cardFlip: 1.0,
  chipClick: 1.0,
  shuffle: 1.0,
  winFanfare: 1.4,
  lossThud: 1.4,
  errorBuzz: 1.4,
};

/**
 * Per-sound kill switch. Dev-side rollback lever: if a centrally-wired sound
 * proves annoying in real play (turnTick ships to ~150 pages with little
 * prior exposure), flip it off here instead of reverting the wiring PR.
 * Not a user setting — the user-facing control stays the single mute toggle.
 */
export const SOUND_ENABLED: Record<SoundName, boolean> = {
  cardDeal: true,
  cardFlip: true,
  cardSelect: true,
  cardPlace: true,
  shuffle: true,
  chipClick: true,
  winFanfare: true,
  lossThud: true,
  errorBuzz: true,
  turnTick: true,
};

/**
 * Minimum interval between plays of the same sound, in ms. Absorbs rapid
 * exec bursts (step/polling games) and per-card AnimatedCard deal bursts.
 * winFanfare/lossThud entries are a TEMPORARY dedupe guard while the 100+
 * page-level onCelebrate sound handlers still exist alongside the central
 * GamePageShell tap — remove those two entries in the sweep PR.
 */
const MIN_INTERVAL_MS: Partial<Record<SoundName, number>> = {
  cardPlace: 90,
  cardDeal: 70,
  cardFlip: 70,
  chipClick: 60,
  turnTick: 400,
  winFanfare: 3000,
  lossThud: 3000,
};

/** How long an unconsumed exec-sound claim stays valid, in ms. */
const CLAIM_EXPIRY_MS = 3000;

/** Idle window after which the cardPlace arpeggio resets, in ms. */
const ARPEGGIO_WINDOW_MS = 1500;

/** Playback-rate increase per consecutive cardPlace step. */
const ARPEGGIO_STEP = 0.035;

/** Upper bound for the arpeggio playback rate. */
const ARPEGGIO_MAX_RATE = 1.25;

/** Sound file paths mapped by name. */
const SOUND_FILES: Record<SoundName, string[]> = {
  cardDeal: ['/sounds/card-deal.ogg', '/sounds/card-deal.mp3'],
  cardFlip: ['/sounds/card-flip.ogg', '/sounds/card-flip.mp3'],
  cardSelect: ['/sounds/card-select.ogg', '/sounds/card-select.mp3'],
  cardPlace: ['/sounds/card-place.ogg', '/sounds/card-place.mp3'],
  shuffle: ['/sounds/shuffle.ogg', '/sounds/shuffle.mp3'],
  chipClick: ['/sounds/chip-click.ogg', '/sounds/chip-click.mp3'],
  winFanfare: ['/sounds/win-fanfare.ogg', '/sounds/win-fanfare.mp3'],
  lossThud: ['/sounds/loss-thud.ogg', '/sounds/loss-thud.mp3'],
  errorBuzz: ['/sounds/error-buzz.ogg', '/sounds/error-buzz.mp3'],
  turnTick: ['/sounds/turn-tick.ogg', '/sounds/turn-tick.mp3'],
};

/** Creates all Howl instances for preloading. Returns null on failure. */
function createSounds(): Record<SoundName, Howl> | null {
  try {
    const sounds = {} as Record<SoundName, Howl>;
    for (const [name, src] of Object.entries(SOUND_FILES)) {
      sounds[name as SoundName] = new Howl({
        src,
        preload: true,
        volume: BASE_VOLUME * VOLUME_TIERS[name as SoundName],
      });
    }
    return sounds;
  } catch {
    return null;
  }
}

/** Provides game sound context to the application. */
export function SoundProvider({ children }: { children: ReactNode }) {
  const [muted, setMuted] = useState(() => {
    try {
      return localStorage.getItem(SOUND_MUTED_KEY) === 'true';
    } catch {
      return false;
    }
  });

  const sounds = useMemo(() => createSounds(), []);

  useEffect(() => {
    try {
      localStorage.setItem(SOUND_MUTED_KEY, String(muted));
    } catch {
      // localStorage unavailable
    }
  }, [muted]);

  const toggleMute = useCallback(() => setMuted((prev) => !prev), []);

  // Policy state lives in refs: shared app-wide (the provider is a
  // singleton), and updating it must never re-render 200+ subscribers.
  const lastPlayedRef = useRef<Partial<Record<SoundName, number>>>({});
  const claimRef = useRef<number | null>(null);
  const arpeggioRef = useRef({ lastAt: 0, step: 0 });

  const claimExecSound = useCallback(() => {
    claimRef.current = Date.now();
  }, []);

  const consumeExecClaim = useCallback(() => {
    const claimedAt = claimRef.current;
    claimRef.current = null;
    return claimedAt !== null && Date.now() - claimedAt <= CLAIM_EXPIRY_MS;
  }, []);

  const playSound = useCallback(
    (name: SoundName, options?: PlayOptions) => {
      if (muted || !sounds) return;
      if (!SOUND_ENABLED[name]) return;

      // While the AudioContext is suspended (no user gesture yet), Howler's
      // autoUnlock QUEUES plays and replays them stale on the first click —
      // it does not drop them. Skip outright instead.
      if (Howler?.ctx?.state === 'suspended') return;

      const now = Date.now();
      const minInterval = MIN_INTERVAL_MS[name];
      const lastPlayed = lastPlayedRef.current[name];
      if (minInterval && lastPlayed !== undefined && now - lastPlayed < minInterval) return;

      const howl = sounds[name];
      if (!howl) return;

      lastPlayedRef.current[name] = now;
      const id = howl.play();

      // Apply volume (tier default or override)
      const vol = options?.volume ?? BASE_VOLUME * VOLUME_TIERS[name];
      howl.volume(vol, id);

      if (options?.pitchVariation) {
        // Explicit pitch randomization wins over the arpeggio.
        const variation = options.pitchVariation;
        const rate = 1 + (Math.random() * 2 - 1) * variation;
        howl.rate(rate, id);
      } else if (name === 'cardPlace') {
        // Burst arpeggio: consecutive card plays inside the window step up
        // in pitch (a solitaire click-run or step-game CPU burst becomes a
        // little rising phrase); idle resets to the base rate.
        const arp = arpeggioRef.current;
        arp.step = now - arp.lastAt <= ARPEGGIO_WINDOW_MS ? arp.step + 1 : 0;
        arp.lastAt = now;
        const rate = Math.min(1 + arp.step * ARPEGGIO_STEP, ARPEGGIO_MAX_RATE);
        howl.rate(rate, id);
      }
    },
    [muted, sounds],
  );

  const value = useMemo(
    () => ({ playSound, muted, toggleMute, claimExecSound, consumeExecClaim }),
    [playSound, muted, toggleMute, claimExecSound, consumeExecClaim],
  );

  return <SoundContext.Provider value={value}>{children}</SoundContext.Provider>;
}

/** Hook to access sound playback from any component within SoundProvider. */
export function useSound(): SoundContextValue {
  const context = useContext(SoundContext);
  if (!context) {
    throw new Error('useSound must be used within a SoundProvider');
  }
  return context;
}

/**
 * Variant of `useSound` that returns `null` when called outside a
 * `SoundProvider` instead of throwing. Used by leaf presentational
 * components (e.g., `AnimatedCard`) that should play their default SFX
 * when wired into the real app but degrade gracefully in unit tests
 * that render them in isolation without a provider.
 */
export function useOptionalSound(): SoundContextValue | null {
  return useContext(SoundContext);
}
