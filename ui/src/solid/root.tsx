import { Outlet, useRouterState } from '@tanstack/solid-router'
import { createEffect } from 'solid-js'
import { AppShell } from './components/common/AppShell'
import { ScrollToTopButton } from './components/common/ScrollToTop'
import { AppNavbar } from './layout/AppNavbar'

export default function Root() {
  const pathname = useRouterState({
    select: (state) => state.location.pathname
  })

  createEffect(() => {
    pathname()
    window.scrollTo({
      top: 0,
      left: 0,
      behavior: 'auto'
    })
  })

  return (
    <AppShell>
      <AppNavbar currentPath={pathname()} />

      <main class="mt-4 pb-10">
        <Outlet />
      </main>

      <ScrollToTopButton />
    </AppShell>
  )
}
