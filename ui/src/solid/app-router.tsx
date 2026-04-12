import {
  createRootRoute,
  createRoute,
  createRouter,
  lazyRouteComponent
} from '@tanstack/solid-router'
import Root from './root'

const rootRoute = createRootRoute({
  component: Root,
  notFoundComponent: lazyRouteComponent(() => import('./routes/NotFoundRoute'))
})

const problemsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: lazyRouteComponent(() => import('./routes/ProblemsRoute'))
})

const contributeProblemRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/problems/contribute',
  component: lazyRouteComponent(() => import('./routes/ContributeProblemRoute'))
})

const solveProblemRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/problems/$slug',
  component: lazyRouteComponent(() => import('./routes/SolveProblemRoute'))
})

const editProblemRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/problems/$slug/edit',
  component: lazyRouteComponent(() => import('./routes/EditProblemRoute'))
})

const problemLabelsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/problem-labels',
  component: lazyRouteComponent(() => import('./routes/ProblemLabelsRoute'))
})

const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/dashboard',
  component: lazyRouteComponent(() => import('./routes/DashboardRoute'))
})

const routeTree = rootRoute.addChildren([
  problemsRoute,
  contributeProblemRoute,
  solveProblemRoute,
  editProblemRoute,
  problemLabelsRoute,
  dashboardRoute
])

export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
  scrollRestoration: true
})

declare module '@tanstack/solid-router' {
  interface Register {
    router: typeof router
  }
}
