import { For, Show, createSignal, onCleanup, onMount } from 'solid-js'
import { Link } from '@tanstack/solid-router'
import { getProgressToday } from '@/api/client'
import type { ProgressToday } from '@/api/types'
import { formatError, navKey, classes } from '../shared/utils'
import { Badge, Button } from '../components/common/Primitives'

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
      label: 'Labels',
      active: () => navKey(props.currentPath) === 'labels'
    },
    {
      href: '/dashboard',
      label: 'Dashboard',
      active: () => navKey(props.currentPath) === 'dashboard'
    }
  ]

  return (
    <header class="sticky top-4 z-20 rounded-2xl border border-gray-200 bg-white/95 shadow-sm backdrop-blur">
      <div class="flex flex-wrap items-center justify-between gap-4 px-4 py-3">
        <Link
          to="/"
          class="flex items-center gap-3"
          onClick={() => {
            setMenuOpen(false)
          }}
        >
          <span class="inline-flex size-10 items-center justify-center rounded-xl bg-blue-600 text-sm font-bold uppercase tracking-wide text-white">
            jl
          </span>
          <div>
            <div class="text-sm font-semibold uppercase tracking-[0.22em] text-gray-900">
              judge-loop
            </div>
            <div class="text-xs text-gray-500">Solid + Flowbite</div>
          </div>
        </Link>

        <div class="flex items-center gap-2">
          <Show when={progress()} fallback={<Badge content={error() || 'syncing'} color="dark" />}>
            {(current) => (
              <>
                <Badge content={`${current().solved} solved`} color="green" />
                <Badge content={`streak ${current().streak}`} color="indigo" />
              </>
            )}
          </Show>
          <Button
            pill
            color="alternative"
            class="md:hidden"
            onClick={() => setMenuOpen((open) => !open)}
          >
            Menu
          </Button>
        </div>
      </div>

      <nav
        class={classes(
          'border-t border-gray-100 px-4 py-3 md:border-t-0 md:px-4 md:pb-4',
          menuOpen() ? 'block' : 'hidden md:block'
        )}
      >
        <div class="flex flex-col gap-2 md:flex-row md:flex-wrap">
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
