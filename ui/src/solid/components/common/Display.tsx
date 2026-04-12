import { For, Show, createMemo, createSignal, type JSX, type ParentProps } from 'solid-js'
import { classes } from '../../shared/utils'

type SurfaceTone = 'blue' | 'green' | 'yellow' | 'red' | 'dark'
type AvatarSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'
type StepStatus = 'complete' | 'current' | 'upcoming'

const toneClasses: Record<SurfaceTone, string> = {
  blue: 'border-blue-200 bg-blue-50 text-blue-900',
  green: 'border-green-200 bg-green-50 text-green-900',
  yellow: 'border-amber-200 bg-amber-50 text-amber-900',
  red: 'border-red-200 bg-red-50 text-red-900',
  dark: 'border-gray-200 bg-gray-100 text-gray-900'
}

const toneFillClasses: Record<SurfaceTone, string> = {
  blue: 'bg-blue-600',
  green: 'bg-green-600',
  yellow: 'bg-amber-500',
  red: 'bg-red-600',
  dark: 'bg-gray-700'
}

const indicatorClasses: Record<'blue' | 'green' | 'yellow' | 'red' | 'gray', string> = {
  blue: 'bg-blue-500',
  green: 'bg-green-500',
  yellow: 'bg-amber-400',
  red: 'bg-red-500',
  gray: 'bg-gray-400'
}

const avatarSizeClasses: Record<AvatarSize, string> = {
  xs: 'size-8 text-xs',
  sm: 'size-10 text-sm',
  md: 'size-12 text-base',
  lg: 'size-16 text-lg',
  xl: 'size-20 text-xl'
}

export type AlertTone = SurfaceTone

export function Alert(
  props: ParentProps<{
    tone?: AlertTone
    title?: string
    action?: JSX.Element
    class?: string
  }>
) {
  return (
    <div
      class={classes(
        'flex flex-col gap-3 rounded-2xl border px-4 py-3 shadow-sm md:flex-row md:items-start md:justify-between',
        toneClasses[props.tone ?? 'blue'],
        props.class
      )}
      role="alert"
    >
      <div class="space-y-1">
        <Show when={props.title}>
          <p class="text-sm font-semibold">{props.title}</p>
        </Show>
        <div class="text-sm">{props.children}</div>
      </div>
      <Show when={props.action}>
        <div class="shrink-0">{props.action}</div>
      </Show>
    </div>
  )
}

export type AccordionItem = {
  title: string
  content: JSX.Element
  description?: string
  defaultOpen?: boolean
  badge?: string
}

export function Accordion(props: { items: AccordionItem[]; class?: string }) {
  return (
    <div
      class={classes(
        'divide-y divide-gray-200 overflow-hidden rounded-2xl border border-gray-200 bg-white',
        props.class
      )}
    >
      <For each={props.items}>
        {(item) => (
          <details open={item.defaultOpen} class="group px-5 py-4">
            <summary class="flex cursor-pointer list-none items-start justify-between gap-4">
              <div class="space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <h3 class="text-sm font-semibold text-gray-900">{item.title}</h3>
                  <Show when={item.badge}>
                    <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600">
                      {item.badge}
                    </span>
                  </Show>
                </div>
                <Show when={item.description}>
                  <p class="text-sm text-gray-500">{item.description}</p>
                </Show>
              </div>
              <span class="pt-0.5 text-gray-400 transition group-open:rotate-180">⌄</span>
            </summary>
            <div class="mt-4 text-sm leading-6 text-gray-600">{item.content}</div>
          </details>
        )}
      </For>
    </div>
  )
}

function initials(name: string | undefined) {
  if (!name) return '?'
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

export function Indicator(props: {
  color?: 'blue' | 'green' | 'yellow' | 'red' | 'gray'
  ping?: boolean
  class?: string
}) {
  return (
    <span class={classes('relative inline-flex size-3', props.class)}>
      <Show when={props.ping}>
        <span
          class={classes(
            'absolute inline-flex h-full w-full animate-ping rounded-full opacity-75',
            indicatorClasses[props.color ?? 'green']
          )}
        />
      </Show>
      <span
        class={classes(
          'relative inline-flex size-3 rounded-full',
          indicatorClasses[props.color ?? 'green']
        )}
      />
    </span>
  )
}

export function Avatar(props: {
  src?: string
  alt?: string
  name?: string
  size?: AvatarSize
  rounded?: boolean
  status?: 'online' | 'offline' | 'busy' | 'away'
  class?: string
}) {
  const statusColor = () => {
    if (props.status === 'busy') return 'red'
    if (props.status === 'away') return 'yellow'
    if (props.status === 'online') return 'green'
    return 'gray'
  }

  return (
    <div class={classes('relative inline-flex', props.class)}>
      <div
        class={classes(
          'inline-flex items-center justify-center overflow-hidden bg-gray-200 font-semibold text-gray-700',
          avatarSizeClasses[props.size ?? 'md'],
          props.rounded === false ? 'rounded-2xl' : 'rounded-full'
        )}
      >
        <Show when={props.src} fallback={<span>{initials(props.name)}</span>}>
          <img
            src={props.src}
            alt={props.alt ?? props.name ?? 'Avatar'}
            class="h-full w-full object-cover"
          />
        </Show>
      </div>
      <Show when={props.status}>
        <Indicator color={statusColor()} class="absolute bottom-0 right-0 ring-2 ring-white" />
      </Show>
    </div>
  )
}

export function Banner(
  props: ParentProps<{
    tone?: AlertTone
    title?: string
    action?: JSX.Element
    dismissible?: boolean
    class?: string
  }>
) {
  const [dismissed, setDismissed] = createSignal(false)

  return (
    <Show when={!dismissed()}>
      <div
        class={classes(
          'flex flex-col gap-3 rounded-2xl border px-4 py-3 shadow-sm md:flex-row md:items-center md:justify-between',
          toneClasses[props.tone ?? 'dark'],
          props.class
        )}
      >
        <div class="space-y-1">
          <Show when={props.title}>
            <p class="text-sm font-semibold">{props.title}</p>
          </Show>
          <div class="text-sm">{props.children}</div>
        </div>
        <div class="flex items-center gap-2">
          <Show when={props.action}>
            <div>{props.action}</div>
          </Show>
          <Show when={props.dismissible}>
            <button
              type="button"
              class="rounded-full px-2 py-1 text-sm font-medium text-gray-500 transition hover:bg-white/70 hover:text-gray-900"
              onClick={() => setDismissed(true)}
              aria-label="Dismiss banner"
            >
              ✕
            </button>
          </Show>
        </div>
      </div>
    </Show>
  )
}

export function ButtonGroup(props: ParentProps<{ class?: string }>) {
  return (
    <div role="group" class={classes('inline-flex flex-wrap items-center gap-2', props.class)}>
      {props.children}
    </div>
  )
}

export type ChatBubbleProps = {
  author?: string
  copy: string
  meta?: string
  align?: 'start' | 'end'
  class?: string
}

export function ChatBubble(props: ChatBubbleProps) {
  const isEnd = () => props.align === 'end'

  return (
    <div class={classes('flex w-full', isEnd() ? 'justify-end' : 'justify-start', props.class)}>
      <div
        class={classes(
          'max-w-xl rounded-2xl px-4 py-3 shadow-sm',
          isEnd() ? 'bg-blue-700 text-white' : 'bg-white text-gray-900 ring-1 ring-gray-200'
        )}
      >
        <Show when={props.author || props.meta}>
          <div
            class={classes(
              'mb-1 flex flex-wrap items-center gap-2 text-xs',
              isEnd() ? 'text-blue-100' : 'text-gray-500'
            )}
          >
            <Show when={props.author}>
              <span class="font-semibold">{props.author}</span>
            </Show>
            <Show when={props.meta}>
              <span>{props.meta}</span>
            </Show>
          </div>
        </Show>
        <p class="text-sm leading-6">{props.copy}</p>
      </div>
    </div>
  )
}

export function Hero(props: {
  eyebrow?: string
  title: string
  copy?: string
  action?: JSX.Element
  secondaryAction?: JSX.Element
  class?: string
}) {
  return (
    <section
      class={classes(
        'rounded-3xl border border-gray-200 bg-gradient-to-br from-white to-gray-100 p-8 shadow-sm sm:p-10',
        props.class
      )}
    >
      <div class="max-w-3xl space-y-4">
        <Show when={props.eyebrow}>
          <p class="text-xs font-semibold uppercase tracking-[0.24em] text-blue-700">
            {props.eyebrow}
          </p>
        </Show>
        <h1 class="text-4xl font-semibold tracking-tight text-gray-900 sm:text-5xl">
          {props.title}
        </h1>
        <Show when={props.copy}>
          <p class="text-base leading-7 text-gray-600 sm:text-lg">{props.copy}</p>
        </Show>
        <Show when={props.action || props.secondaryAction}>
          <div class="flex flex-wrap gap-3 pt-2">
            <Show when={props.action}>
              <div>{props.action}</div>
            </Show>
            <Show when={props.secondaryAction}>
              <div>{props.secondaryAction}</div>
            </Show>
          </div>
        </Show>
      </div>
    </section>
  )
}

export const Jumbotron = Hero

export type ListGroupItem = {
  label: string
  description?: string
  href?: string
  prefix?: JSX.Element
  suffix?: JSX.Element
  active?: boolean
  onClick?: () => void
}

export function ListGroup(props: { items: ListGroupItem[]; class?: string }) {
  return (
    <div
      class={classes(
        'overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm',
        props.class
      )}
    >
      <For each={props.items}>
        {(item) => {
          const itemClass = classes(
            'flex w-full items-center justify-between gap-4 border-b border-gray-100 px-4 py-3 text-left transition last:border-b-0',
            item.active ? 'bg-blue-50 text-blue-900' : 'text-gray-700 hover:bg-gray-50'
          )

          const content = (
            <>
              <div class="flex min-w-0 items-start gap-3">
                <Show when={item.prefix}>
                  <div class="pt-0.5">{item.prefix}</div>
                </Show>
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium">{item.label}</p>
                  <Show when={item.description}>
                    <p class="mt-1 text-sm text-gray-500">{item.description}</p>
                  </Show>
                </div>
              </div>
              <Show when={item.suffix}>
                <div class="shrink-0">{item.suffix}</div>
              </Show>
            </>
          )

          return item.href ? (
            <a href={item.href} class={itemClass}>
              {content}
            </a>
          ) : item.onClick ? (
            <button type="button" class={itemClass} onClick={item.onClick}>
              {content}
            </button>
          ) : (
            <div class={itemClass}>{content}</div>
          )
        }}
      </For>
    </div>
  )
}

export function ProgressBar(props: {
  value: number
  max?: number
  label?: string
  tone?: AlertTone
  showValue?: boolean
  class?: string
}) {
  const percent = createMemo(() => {
    const max = props.max ?? 100
    if (max <= 0) return 0
    return Math.max(0, Math.min(100, Math.round((props.value / max) * 100)))
  })

  return (
    <div class={classes('space-y-2', props.class)}>
      <Show when={props.label || props.showValue}>
        <div class="flex items-center justify-between gap-4 text-sm">
          <Show when={props.label}>
            <span class="font-medium text-gray-700">{props.label}</span>
          </Show>
          <Show when={props.showValue}>
            <span class="text-gray-500">{percent()}%</span>
          </Show>
        </div>
      </Show>
      <div class="h-2.5 overflow-hidden rounded-full bg-gray-200">
        <div
          class={classes(
            'h-full rounded-full transition-[width]',
            toneFillClasses[props.tone ?? 'blue']
          )}
          style={{ width: `${percent()}%` }}
        />
      </div>
    </div>
  )
}

export function Rating(props: { value: number; max?: number; class?: string }) {
  const max = () => props.max ?? 5

  return (
    <div
      class={classes('inline-flex items-center gap-1', props.class)}
      aria-label={`Rated ${props.value} out of ${max()}`}
    >
      <For each={Array.from({ length: max() }, (_, index) => index + 1)}>
        {(item) => (
          <span
            class={classes('text-lg', item <= props.value ? 'text-amber-400' : 'text-gray-300')}
          >
            ★
          </span>
        )}
      </For>
    </div>
  )
}

export function Skeleton(props: { class?: string; circle?: boolean }) {
  return (
    <div
      class={classes(
        'animate-pulse bg-gray-200',
        props.circle ? 'rounded-full' : 'rounded-xl',
        props.class ?? 'h-4 w-full'
      )}
      aria-hidden="true"
    />
  )
}

export function SkeletonText(props: { lines?: number; class?: string }) {
  return (
    <div class={classes('space-y-3', props.class)} aria-hidden="true">
      <For each={Array.from({ length: props.lines ?? 3 }, (_, index) => index)}>
        {(index) => (
          <Skeleton
            class={classes('h-4 rounded-lg', index === (props.lines ?? 3) - 1 ? 'w-3/4' : 'w-full')}
          />
        )}
      </For>
    </div>
  )
}

export type StepperItem = {
  title: string
  description?: string
  status?: StepStatus
}

export function Stepper(props: { items: StepperItem[]; class?: string }) {
  return (
    <ol class={classes('space-y-4', props.class)}>
      <For each={props.items}>
        {(item, index) => {
          const status = () => item.status ?? 'upcoming'

          return (
            <li class="flex gap-4">
              <div class="flex flex-col items-center">
                <div
                  class={classes(
                    'flex size-8 items-center justify-center rounded-full text-xs font-semibold',
                    status() === 'complete'
                      ? 'bg-green-600 text-white'
                      : status() === 'current'
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-200 text-gray-600'
                  )}
                >
                  {status() === 'complete' ? '✓' : index() + 1}
                </div>
                <Show when={index() < props.items.length - 1}>
                  <div class="mt-2 h-full min-h-6 w-px bg-gray-200" />
                </Show>
              </div>
              <div class="space-y-1 pb-2">
                <p class="text-sm font-semibold text-gray-900">{item.title}</p>
                <Show when={item.description}>
                  <p class="text-sm text-gray-500">{item.description}</p>
                </Show>
              </div>
            </li>
          )
        }}
      </For>
    </ol>
  )
}

export type TimelineItem = {
  title: string
  meta?: string
  copy?: string
}

export function Timeline(props: { items: TimelineItem[]; class?: string }) {
  return (
    <ol class={classes('relative space-y-6 border-s border-gray-200 ps-6', props.class)}>
      <For each={props.items}>
        {(item) => (
          <li class="relative">
            <span class="absolute -start-[2.05rem] mt-1.5 size-3 rounded-full bg-blue-600 ring-4 ring-white" />
            <div class="space-y-1">
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="text-sm font-semibold text-gray-900">{item.title}</h3>
                <Show when={item.meta}>
                  <span class="text-xs uppercase tracking-wide text-gray-400">{item.meta}</span>
                </Show>
              </div>
              <Show when={item.copy}>
                <p class="text-sm leading-6 text-gray-500">{item.copy}</p>
              </Show>
            </div>
          </li>
        )}
      </For>
    </ol>
  )
}

export function Toast(
  props: ParentProps<{
    show?: boolean
    tone?: AlertTone
    title?: string
    fixed?: boolean
    action?: JSX.Element
    onClose?: () => void
    class?: string
  }>
) {
  return (
    <Show when={props.show ?? true}>
      <div
        class={classes(
          'z-50 flex max-w-sm items-start justify-between gap-4 rounded-2xl border bg-white px-4 py-3 shadow-xl',
          props.fixed !== false && 'fixed bottom-4 right-4',
          toneClasses[props.tone ?? 'dark'],
          props.class
        )}
        role="status"
      >
        <div class="space-y-1">
          <Show when={props.title}>
            <p class="text-sm font-semibold">{props.title}</p>
          </Show>
          <div class="text-sm">{props.children}</div>
        </div>
        <div class="flex items-start gap-2">
          <Show when={props.action}>
            <div class="shrink-0">{props.action}</div>
          </Show>
          <Show when={props.onClose}>
            <button
              type="button"
              class="rounded-full px-2 py-1 text-sm font-medium text-gray-500 hover:bg-white/70 hover:text-gray-900"
              onClick={() => props.onClose?.()}
              aria-label="Close toast"
            >
              ✕
            </button>
          </Show>
        </div>
      </div>
    </Show>
  )
}
