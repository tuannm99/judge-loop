import { Link } from '@tanstack/solid-router'
import { For, Show, createSignal, onCleanup, onMount } from 'solid-js'
import { getProgressToday } from '@/api/client'
import type { ProgressToday } from '@/api/types'
import { Badge, Button } from '../components/common/Primitives'
import { classes, formatError, navKey } from '../shared/utils'

export function AppNavbar(props: { currentPath: string }) {
  const [progress, setProgress] = createSignal<ProgressToday | null>(null)
  const [error, setError] = createSignal('')
  const [menuOpen, setMenuOpen] = createSignal(false)

  onMount(() => {
    let active = true

    const load = async () => {
      try {
        const next = await getProgressToday()
        if (active) {
          setProgress(next)
          setError('')
        }
      } catch (loadError) {
        if (active) {
          setError(formatError(loadError))
        }
      }
    }

    void load()
    const timer = window.setInterval(() => {
      void load()
    }, 60_000)

    onCleanup(() => {
      active = false
      clearInterval(timer)
    })
  })

  const links = [
    {
      href: '/',
      label: 'Problems',
      active: () => navKey(props.currentPath) === 'problems'
    },
    {
      href: '/problems/contribute',
      label: 'New Problem',
      active: () => props.currentPath === '/problems/contribute'
    },
    {
      href: '/problem-labels',
      label: 'Tags',
      active: () => navKey(props.currentPath) === 'labels'
    },
    {
      href: '/dashboard',
      label: 'Dashboard',
      active: () => navKey(props.currentPath) === 'dashboard'
    }
  ]

  return (
    <header class="-mx-4 sticky top-0 z-40 border-b border-gray-200 bg-white/90 shadow-sm backdrop-blur sm:-mx-6 lg:-mx-8">
      <div class="flex min-h-14 items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
        <Link
          to="/"
          class="flex min-w-0 items-center gap-3"
          onClick={() => {
            setMenuOpen(false)
          }}
        >
          <span class="inline-flex size-8 shrink-0 items-center justify-center rounded-lg bg-blue-600 text-xs font-bold uppercase tracking-[0.22em] text-white">
            jl
          </span>
          <div class="min-w-0">
            <div class="truncate text-sm font-semibold text-gray-900">Judge Loop</div>
            <div class="hidden text-xs text-gray-500 sm:block">Train, solve, review</div>
          </div>
        </Link>

        <nav class="hidden items-center gap-1 md:flex">
          <For each={links}>
            {(link) => (
              <Link
                to={link.href}
                class={classes(
                  'rounded-full px-3 py-1.5 text-sm font-medium transition',
                  link.active()
                    ? 'bg-blue-50 text-blue-700'
                    : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                )}
                onClick={() => {
                  setMenuOpen(false)
                }}
              >
                {link.label}
              </Link>
            )}
          </For>
        </nav>

        <div class="flex shrink-0 items-center gap-2">
          <Show
            when={progress()}
            fallback={
              <Badge
                content={error() || 'syncing'}
                color="dark"
                class="hidden text-[11px] sm:inline-flex"
              />
            }
          >
            {(current) => (
              <>
                <Badge
                  content={`${current().solved} solved`}
                  color="green"
                  class="hidden text-[11px] sm:inline-flex"
                />
                <Badge
                  content={`${current().streak} streak`}
                  color="indigo"
                  class="hidden text-[11px] lg:inline-flex"
                />
              </>
            )}
          </Show>
          <Button
            pill
            color="alternative"
            size="sm"
            class="md:hidden"
            onClick={() => setMenuOpen((open) => !open)}
          >
            Menu
          </Button>
        </div>
      </div>

      <nav
        class={classes(
          'border-t border-gray-200 px-4 py-3 md:hidden',
          menuOpen() ? 'block' : 'hidden'
        )}
      >
        <div class="flex flex-col gap-2">
          <For each={links}>
            {(link) => (
              <Link
                to={link.href}
                class={classes(
                  'rounded-xl px-3 py-2 text-sm font-medium transition',
                  link.active()
                    ? 'bg-blue-50 text-blue-700'
                    : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                )}
                onClick={() => {
                  setMenuOpen(false)
                }}
              >
                {link.label}
              </Link>
            )}
          </For>
        </div>
      </nav>
    </header>
  )
}
