import { render, screen, waitFor } from '@testing-library/react';
import type { ComponentType } from 'react';
import { Suspense } from 'react';
import { describe, expect, it } from 'vitest';
import { pickPageExport, resolvePageComponent } from './resolvePageComponent';

describe('resolvePageComponent', () => {
  it('returns a lazy component that resolves to the named export', async () => {
    const modules = {
      './pages/FooPage.tsx': async () => ({
        FooPage: (() => <div data-testid="foo">foo-content</div>) as ComponentType<unknown>,
      }),
    };
    const Lazy = resolvePageComponent(modules, '/foo', 'Foo');
    render(
      <Suspense fallback={<div>loading</div>}>
        <Lazy />
      </Suspense>,
    );
    await waitFor(() => expect(screen.getByTestId('foo')).toBeInTheDocument());
  });

  it('throws synchronously when the module is missing', () => {
    expect(() => resolvePageComponent({}, '/foo', 'Foo')).toThrow(
      /no module at \.\/pages\/FooPage\.tsx for path "\/foo"/,
    );
  });
});

describe('pickPageExport', () => {
  it('returns the named export when present', () => {
    const Foo: ComponentType<unknown> = () => null;
    const got = pickPageExport({ FooPage: Foo }, './pages/FooPage.tsx', 'Foo');
    expect(got).toBe(Foo);
  });

  it('throws when the named export is missing', () => {
    expect(() =>
      pickPageExport({ OtherPage: (() => null) as ComponentType<unknown> }, './pages/FooPage.tsx', 'Foo'),
    ).toThrow(/no export named FooPage/);
  });
});
