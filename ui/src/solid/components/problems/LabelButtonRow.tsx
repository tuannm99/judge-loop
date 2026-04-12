import { For, Show } from 'solid-js'
import { Button } from '../common/Primitives'
import { LoadingInline } from '../common/Feedback'

export function LabelButtonRow(props: {
  title: string
  helperText?: string
  values: string[]
  selected: string[]
  loading: boolean
  activeColor: 'blue' | 'indigo'
  onToggle: (value: string) => void
  onClear?: () => void
}) {
  return (
    <div class="space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="space-y-1">
          <div class="flex flex-wrap items-center gap-2">
            <div class="text-sm font-medium text-gray-700">{props.title}</div>
            <Show when={props.selected.length > 0}>
              <span class="rounded-full bg-gray-100 px-2.5 py-1 text-xs font-semibold text-gray-700">
                {props.selected.length} selected
              </span>
            </Show>
          </div>
          <p class="text-xs text-gray-500">
            {props.helperText ?? 'Multi-select. Click again to remove a selection.'}
          </p>
        </div>

        <Show when={props.onClear && props.selected.length > 0}>
          <Button pill size="xs" color="light" onClick={() => props.onClear?.()}>
            Clear all
          </Button>
        </Show>
      </div>
      <Show
        when={!props.loading}
        fallback={<LoadingInline label={`Loading ${props.title.toLowerCase()}...`} />}
      >
        <Show
          when={props.values.length > 0}
          fallback={<p class="text-sm text-gray-500">No {props.title.toLowerCase()} yet.</p>}
        >
          <div class="flex flex-wrap gap-2">
            <For each={props.values}>
              {(value) => (
                <Button
                  pill
                  size="xs"
                  aria-pressed={props.selected.includes(value)}
                  color={
                    props.selected.includes(value)
                      ? props.activeColor === 'indigo'
                        ? 'dark'
                        : 'blue'
                      : 'alternative'
                  }
                  onClick={() => props.onToggle(value)}
                >
                  {props.selected.includes(value) ? `✓ ${value}` : value}
                </Button>
              )}
            </For>
          </div>
        </Show>
      </Show>
    </div>
  )
}
