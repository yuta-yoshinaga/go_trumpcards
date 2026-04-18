import { useEffect, useState } from 'react';

/** Mermaid API surface used by this component. */
interface MermaidApi {
  initialize: (config: Record<string, unknown>) => void;
  render: (id: string, code: string) => Promise<{ svg: string }>;
}

let mermaidPromise: Promise<MermaidApi> | null = null;
let mermaidIdCounter = 0;

/** Lazily loads mermaid via dynamic import (code-split, not bundled into main chunk). */
function loadMermaid(): Promise<MermaidApi> {
  if (!mermaidPromise) {
    mermaidPromise = import('mermaid').then((mod) => {
      const api = mod.default as MermaidApi;
      api.initialize({ startOnLoad: false, theme: 'dark' });
      return api;
    });
  }
  return mermaidPromise;
}

/** Renders a Mermaid diagram from the given code string. */
export function MermaidBlock({ code }: { code: string }) {
  const [svg, setSvg] = useState<string>('');
  const [error, setError] = useState<string>('');

  useEffect(() => {
    let cancelled = false;
    const id = `mermaid-${++mermaidIdCounter}`;
    loadMermaid()
      .then((api) => api.render(id, code))
      .then(({ svg: rendered }) => {
        if (!cancelled) setSvg(rendered);
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(String(err));
      });
    return () => {
      cancelled = true;
    };
  }, [code]);

  if (error) {
    return (
      <pre className="text-ds-error text-xs overflow-x-auto">
        <code>{code}</code>
      </pre>
    );
  }

  if (!svg) return null;

  return (
    <div
      className="my-4 flex justify-center overflow-x-auto"
      // biome-ignore lint/security/noDangerouslySetInnerHtml: mermaid.render produces trusted SVG from bundled manual markdown
      dangerouslySetInnerHTML={{ __html: svg }}
    />
  );
}
