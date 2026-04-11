import { component$, Slot, type PropsOf } from '@builder.io/qwik'

export type LinkProps = PropsOf<'a'>

export const Link = component$<LinkProps>(({ ...props }) => {
  return (
    <a {...props}>
      <Slot />
    </a>
  )
})
