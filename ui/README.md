# judge-loop UI

This frontend is now a client-side SolidJS app under `src/solid`.

Commands:

- `npm run dev` starts Vite on `http://localhost:3000`
- `npm run build` runs a type-check and production build
- `npm run preview` serves the built bundle locally

The app still talks to the existing backend through the `/api` proxy in
`vite.config.ts`.

Structure:

- `src/solid/app-router.tsx` defines the TanStack Router tree
- `src/solid/routes/*` are lazy route wrappers for code splitting
- `src/solid/pages/*` are page-level views and state
- `src/solid/components/common/*` are reusable layout/form/display primitives
- `src/solid/components/index.ts` is the main barrel for future reuse
- `src/solid/layout/*` holds app chrome
- `src/solid/shared/*` holds framework-agnostic constants, helpers, and types

Reusable UI surface:

- Layout: `AppShell`, `Container`, `PageShell`, `SectionLead`, `SectionTitle`, `PanelHeader`, `Breadcrumbs`
- Forms: `InputField`, `SearchInputField`, `NumberInputField`, `PhoneInputField`, `DateInputField`, `TimeInputField`, `DatepickerField`, `Calendar`, `FileInputField`, `SelectField`, `TextareaField`, `CheckboxField`, `RadioGroupField`, `ToggleField`, `RangeField`, `FloatingLabelField`
- Primitives: `Button`, `Badge`, `Card`, `Spinner`, `ButtonGroup`
- Navigation: `Navbar`, `Sidebar`, `BottomNavigation`, `Footer`, `SpeedDial`, `MegaMenu`, `Breadcrumbs`, `Pagination`, `Tabs`
- Surfaces: `Accordion`, `Alert`, `Banner`, `Hero`, `Jumbotron`, `Carousel`, `ListGroup`, `ProgressBar`, `Stepper`, `Timeline`, `Toast`, `ChatBubble`
- Media and utilities: `Avatar`, `Indicator`, `GalleryGrid`, `VideoEmbed`, `QrCode`, `QrCodeCard`, `ClipboardButton`, `DeviceMockup`, `CodeBlock`
- Typography: `Heading`, `Paragraph`, `Blockquote`, `TextLink`, `InlineCode`, `Kbd`, `Divider`, `ProseList`
- Data and feedback: `DataTable`, `FormSection`, `EmptyBlock`, `ErrorAlert`, `WarningAlert`, `InfoAlert`, `LoadingInline`, `LoadingBlock`, `Skeleton`, `SkeletonText`, `Rating`
- Overlays: `DropdownMenu`, `Modal`, `Drawer`, `Popover`, `Tooltip`

Production image:

- `deploy/docker/Dockerfile.ui` builds the Solid app and serves it through nginx
- `deploy/docker/nginx.ui.conf.template` serves the SPA and proxies `/api`
- `deploy/docker/ui-entrypoint.sh` injects `API_UPSTREAM` at container start

Example:

```bash
docker build -f deploy/docker/Dockerfile.ui -t judge-loop-ui .
docker run --rm -p 3000:3000 -e API_UPSTREAM=http://host.docker.internal:8080 judge-loop-ui
```
