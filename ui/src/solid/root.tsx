import { Outlet, useRouterState } from '@tanstack/solid-router'
import { AppShell } from './components/common/AppShell'
import { AppNavbar } from './layout/AppNavbar'

export default function Root() {
  const pathname = useRouterState({
    select: (state) => state.location.pathname
  })

  return (
    <AppShell>
      <AppNavbar currentPath={pathname()} />

      <main class="mt-6">
        <Outlet />
      </main>
    </AppShell>
  )
}
