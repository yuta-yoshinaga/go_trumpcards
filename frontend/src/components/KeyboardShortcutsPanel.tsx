import type { ComponentPropsWithoutRef } from 'react';

/** A single keyboard-shortcut row: one or more key chips plus a description. */
export interface KeyboardShortcut {
  /** Key labels shown as `<kbd>` chips, e.g. `["Enter"]` or `["1", "0"]` for a range. */
  keys: string[];
  /** Plain-language description of what the shortcut does. */
  description: string;
}

/** Props for the {@link KeyboardShortcutsPanel} component. */
export interface KeyboardShortcutsPanelProps extends ComponentPropsWithoutRef<'details'> {
  /** Heading rendered in the always-visible `<summary>`. */
  title: string;
  /** Shortcut rows to list when the panel is expanded. */
  shortcuts: KeyboardShortcut[];
}

/**
 * Collapsible, keyboard-discoverable list of the shortcuts available on a game
 * page. Rendered as a native `<details>`/`<summary>` so it is closed by
 * default, focusable, and operable without JavaScript. Unlike {@link components/KbdBadge.KbdBadge | KbdBadge}
 * (a silent affordance on a button whose name already conveys the action), the
 * key chips here are read by assistive tech because advertising the keys *is*
 * the panel's purpose.
 */
export function KeyboardShortcutsPanel({ title, shortcuts, className, ...rest }: KeyboardShortcutsPanelProps) {
  return (
    <details className={`mt-2 w-full max-w-md mx-auto text-ds-text-muted ${className ?? ''}`} {...rest}>
      <summary className="text-xs cursor-pointer min-h-[44px] flex items-center justify-center">{title}</summary>
      <ul className="mt-1 flex flex-col gap-1 px-2 pb-2">
        {shortcuts.map((s, si) => (
          <li key={`${s.description}-${si}`} className="flex items-center justify-between gap-3 text-xs">
            <span className="flex flex-wrap items-center gap-1">
              {s.keys.map((k, i) => (
                <span key={`${k}-${i}`} className="flex items-center gap-1">
                  {i > 0 && (
                    <span aria-hidden="true" className="opacity-60">
                      –
                    </span>
                  )}
                  <kbd className="inline-flex items-center px-1.5 py-0.5 rounded border border-white/40 bg-white/15 font-mono leading-none">
                    {k}
                  </kbd>
                </span>
              ))}
            </span>
            <span className="flex-1 text-right">{s.description}</span>
          </li>
        ))}
      </ul>
    </details>
  );
}
