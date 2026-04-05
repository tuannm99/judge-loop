import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MantineProvider, createTheme } from '@mantine/core'
import { Navbar } from './components/Navbar'
import { ProblemList } from './pages/ProblemList'
import { Solve } from './pages/Solve'
import { Dashboard } from './pages/Dashboard'
import { ContributeProblem } from './pages/ContributeProblem'
import { ProblemLabelsPage } from './pages/ProblemLabels'

const theme = createTheme({
  primaryColor: 'teal',
  defaultRadius: 'sm'
})

const qc = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 }
  }
})

export default function App() {
  return (
    <MantineProvider theme={theme} defaultColorScheme="dark">
      <QueryClientProvider client={qc}>
        <BrowserRouter>
          <Navbar />
          <Routes>
            <Route path="/" element={<ProblemList />} />
            <Route path="/problems/:slug" element={<Solve />} />
            <Route
              path="/problems/contribute"
              element={<ContributeProblem />}
            />
            <Route path="/problem-labels" element={<ProblemLabelsPage />} />
            <Route path="/dashboard" element={<Dashboard />} />
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>
    </MantineProvider>
  )
}
