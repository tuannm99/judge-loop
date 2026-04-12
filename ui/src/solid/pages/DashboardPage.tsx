import { For, Show, onCleanup, onMount } from 'solid-js'
import { createStore } from 'solid-js/store'
import { getProgressToday, getReviewsToday, getStreak } from '@/api/client'
import type { ProgressToday, ReviewItem, Streak } from '@/api/types'
import { EmptyBlock, ErrorAlert, LoadingBlock } from '../components/common/Feedback'
import { PageShell } from '../components/common/PageShell'
import { Badge, Card } from '../components/common/Primitives'
import { SectionLead } from '../components/common/SectionLead'
import { SectionTitle } from '../components/common/SectionTitle'
import { StatCard } from '../components/dashboard/StatCard'
import type { NavigateFn } from '../shared/types'
import { formatDate, formatError } from '../shared/utils'

export function DashboardPage(props: { navigate: NavigateFn }) {
  const [state, setState] = createStore({
    loading: true,
    error: '',
    progress: null as ProgressToday | null,
    streak: null as Streak | null,
    reviews: [] as ReviewItem[]
  })

  onMount(() => {
    let active = true

    const load = async () => {
      setState('loading', true)
      try {
        const [progress, streak, reviews] = await Promise.all([
          getProgressToday(),
          getStreak(),
          getReviewsToday()
        ])

        if (active) {
          setState({
            progress,
            streak,
            reviews: reviews.reviews,
            error: ''
          })
        }
      } catch (error) {
        if (active) {
          setState('error', formatError(error))
        }
      } finally {
        if (active) {
          setState('loading', false)
        }
      }
    }

    void load()
    const timer = window.setInterval(() => {
      void load()
    }, 30_000)

    onCleanup(() => {
      active = false
      clearInterval(timer)
    })
  })

  return (
    <PageShell>
      <Card class="space-y-4">
        <SectionLead
          eyebrow="Daily signal"
          title="Track progress with Solid."
          copy="The dashboard is now rendered with Solid components and the same backend endpoints."
        />
      </Card>

      <Show when={state.error}>
        <ErrorAlert>{state.error}</ErrorAlert>
      </Show>

      <Show
        when={!state.loading}
        fallback={
          <Card>
            <LoadingBlock label="Loading dashboard..." />
          </Card>
        }
      >
        <div class="space-y-6">
          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <StatCard label="Solved today" value={String(state.progress?.solved ?? 0)} />
            <StatCard label="Attempted" value={String(state.progress?.attempted ?? 0)} />
            <StatCard label="Time today" value={`${state.progress?.time_spent_minutes ?? 0}m`} />
            <StatCard label="Current streak" value={String(state.streak?.current ?? 0)} />
          </div>

          <Card class="space-y-4">
            <SectionTitle
              title="Streak line"
              subtitle={state.progress?.date ? `As of ${state.progress.date}` : undefined}
            />
            <div class="grid gap-3 text-sm text-gray-600 md:grid-cols-3">
              <div class="rounded-xl bg-gray-50 p-4">current: {state.streak?.current ?? 0}</div>
              <div class="rounded-xl bg-gray-50 p-4">longest: {state.streak?.longest ?? 0}</div>
              <div class="rounded-xl bg-gray-50 p-4">
                last practiced: {formatDate(state.streak?.last_practiced)}
              </div>
            </div>
          </Card>

          <Card class="space-y-4">
            <SectionTitle
              title="Reviews due today"
              subtitle="Jump back into stale problems before the streak cools off."
            />
            <Show
              when={state.reviews.length > 0}
              fallback={
                <EmptyBlock title="No reviews due." copy="You have a clean board right now." />
              }
            >
              <div class="space-y-3">
                <For each={state.reviews}>
                  {(review) => (
                    <button
                      class="flex w-full items-center justify-between rounded-2xl border border-gray-200 bg-white px-4 py-4 text-left transition hover:border-gray-300"
                      onClick={() => props.navigate(`/problems/${review.slug}`)}
                    >
                      <div class="space-y-1">
                        <h3 class="text-base font-semibold text-gray-900">{review.title}</h3>
                        <p class="text-sm text-gray-500">slug: {review.slug}</p>
                      </div>
                      <Badge
                        content={
                          review.days_overdue > 0 ? `${review.days_overdue}d overdue` : 'due today'
                        }
                        color={review.days_overdue > 0 ? 'pink' : 'blue'}
                      />
                    </button>
                  )}
                </For>
              </div>
            </Show>
          </Card>
        </div>
      </Show>
    </PageShell>
  )
}
