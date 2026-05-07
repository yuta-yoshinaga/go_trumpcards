/** Props for ChipBetInput. */
export interface ChipBetInputProps {
  /** DOM id used to associate the label with the input. */
  id: string;
  /** Visible label text. */
  label: string;
  /** Current bet value. */
  value: number;
  /** Called with the new numeric value when the user edits the field. */
  onChange: (value: number) => void;
  /**
   * Maximum allowed bet (typically the player's chip balance). When omitted, no
   * max attribute is rendered on the input and onChange clamping is unbounded
   * above min — useful for "no-limit" poker bet amounts.
   */
  max?: number;
  /** Minimum allowed bet. Defaults to 10. */
  min?: number;
  /** Step size. Defaults to 10. */
  step?: number;
  /** Disable the input. */
  disabled?: boolean;
  /** Tailwind width class for the input. Defaults to "w-24". */
  widthClass?: string;
  /**
   * Whether to clamp emitted values to [min, max] inside onChange. Defaults to true.
   * Set to false when the parent wants to display and validate out-of-range values
   * (e.g. poker BettingControls showing a range hint).
   */
  autoClamp?: boolean;
  /** When true, mark the input as invalid (error styling + aria-invalid). */
  invalid?: boolean;
  /** Optional id of an element describing the input (typically an error message). */
  describedBy?: string;
}

/**
 * Reusable label + numeric chip-bet input used across casino games
 * (Baccarat, Caribbean Stud, Three Card Poker, Pai Gow, Let It Ride, BlackJack, Poker).
 */
export function ChipBetInput({
  id,
  label,
  value,
  onChange,
  max,
  min = 10,
  step = 10,
  disabled,
  widthClass = 'w-24',
  autoClamp = true,
  invalid,
  describedBy,
}: ChipBetInputProps) {
  // Foreground stays text-ds-text-primary (10.1:1 AAA on surface) — pairing
  // text-ds-error with bg-ds-surface only hits ~2.7:1, well below WCAG AA.
  // The error semantic is carried entirely by the coloured border.
  const errorClasses = invalid ? 'bg-ds-surface border-ds-error text-ds-text-primary' : '';
  return (
    <div className="flex items-center gap-2">
      <label htmlFor={id} className="text-ds-text-primary text-sm">
        {label}
      </label>
      <input
        id={id}
        type="text"
        inputMode="numeric"
        pattern="[0-9]*"
        autoComplete="off"
        enterKeyHint="done"
        min={min}
        max={max}
        step={step}
        value={value}
        aria-invalid={invalid || undefined}
        aria-describedby={describedBy}
        onChange={(e) => {
          const raw = e.target.value;
          const cleaned = raw.replace(/[^0-9]/g, '');
          // User typed only non-digit chars: reject (mirrors prior type=number rejection of NaN input).
          if (raw !== '' && cleaned === '') return;
          const parsed = cleaned === '' ? 0 : Number(cleaned);
          if (!autoClamp) {
            onChange(parsed);
            return;
          }
          const upper = max ?? Number.POSITIVE_INFINITY;
          onChange(Math.max(min, Math.min(parsed, upper)));
        }}
        onWheel={(e) => e.currentTarget.blur()}
        disabled={disabled}
        className={`${widthClass} px-3 py-2 rounded text-base min-h-[44px] ${errorClasses}`}
      />
    </div>
  );
}
