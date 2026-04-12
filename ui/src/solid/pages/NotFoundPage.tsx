import { PageShell } from '../components/common/PageShell'
import { Button, Card } from '../components/common/Primitives'
import { SectionLead } from '../components/common/SectionLead'
import type { NavigateFn } from '../shared/types'

export function NotFoundPage(props: { navigate: NavigateFn; path: string }) {
  return (
    <PageShell>
      <Card class="space-y-6">
        <SectionLead
          eyebrow="Route not found"
          title="That page does not exist."
          copy={`The current path "${props.path}" is not wired into the Solid app.`}
        />
        <div class="flex justify-start">
          <Button pill onClick={() => props.navigate('/')}>
            Go back to problems
          </Button>
        </div>
      </Card>
    </PageShell>
  )
}
