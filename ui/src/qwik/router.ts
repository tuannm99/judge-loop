export type RouteMatch =
  | { name: 'problems'; path: '/' }
  | { name: 'contribute-problem'; path: '/problems/contribute' }
  | { name: 'edit-problem'; path: string; slug: string }
  | { name: 'solve-problem'; path: string; slug: string }
  | { name: 'problem-labels'; path: '/problem-labels' }
  | { name: 'dashboard'; path: '/dashboard' }
  | { name: 'not-found'; path: string }

function normalizePath(pathname: string) {
  if (!pathname) return '/'
  const stripped = pathname.split('?')[0].split('#')[0] || '/'
  if (stripped === '/') return '/'
  return stripped.endsWith('/') ? stripped.slice(0, -1) : stripped
}

export function matchRoute(pathname: string): RouteMatch {
  const path = normalizePath(pathname)

  if (path === '/') return { name: 'problems', path: '/' }
  if (path === '/problems/contribute') {
    return { name: 'contribute-problem', path }
  }
  if (path === '/problem-labels') return { name: 'problem-labels', path }
  if (path === '/dashboard') return { name: 'dashboard', path }

  const editMatch = path.match(/^\/problems\/([^/]+)\/edit$/)
  if (editMatch) {
    return {
      name: 'edit-problem',
      path,
      slug: decodeURIComponent(editMatch[1])
    }
  }

  const solveMatch = path.match(/^\/problems\/([^/]+)$/)
  if (solveMatch) {
    return {
      name: 'solve-problem',
      path,
      slug: decodeURIComponent(solveMatch[1])
    }
  }

  return { name: 'not-found', path }
}
