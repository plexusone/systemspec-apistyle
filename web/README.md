# systemspec-apistyle Web UI

Web-based interface for systemspec-apistyle linting and evaluation.

## Tech Stack

- **Framework**: [Lit](https://lit.dev/) Web Components
- **Build**: [Vite](https://vitejs.dev/)
- **Language**: TypeScript
- **Styling**: CSS Custom Properties (design tokens)

## Development

### Prerequisites

- Node.js 20+
- pnpm (recommended) or npm

### Setup

```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev
```

The app will be available at http://localhost:3000

### Build

```bash
# Production build
pnpm build

# Preview production build
pnpm preview
```

## Project Structure

```
web/
├── public/              # Static assets
│   └── favicon.svg
├── src/
│   ├── components/      # Lit web components
│   │   ├── app.ts       # Main application
│   │   ├── spec-editor.ts
│   │   ├── lint-results.ts
│   │   └── profile-selector.ts
│   ├── types.ts         # TypeScript type definitions
│   └── main.ts          # Entry point
├── index.html           # HTML template
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## Components

### `<api-style-app>`

Main application component. Orchestrates the spec editor, profile selection, and lint results display.

### `<spec-editor>`

OpenAPI specification editor with:
- Syntax-highlighted textarea
- Sample spec loading
- Clear functionality

### `<lint-results>`

Displays linting results including:
- Summary counts (errors, warnings, info, hints)
- Individual violation details
- Pass/fail status

### `<profile-selector>`

Dropdown for selecting style profiles:
- default
- azure
- google
- zalando

## Design Tokens

The UI uses CSS custom properties for theming:

```css
/* Colors */
--color-primary: #4f46e5;
--color-success: #10b981;
--color-warning: #f59e0b;
--color-error: #ef4444;

/* Typography */
--font-sans: 'Inter', system-ui, sans-serif;
--font-mono: 'JetBrains Mono', monospace;

/* Spacing */
--spacing-sm: 0.5rem;
--spacing-md: 1rem;
--spacing-lg: 1.5rem;
```

Dark mode is automatically supported via `prefers-color-scheme`.

## Backend Integration

The UI connects to the Go backend REST API for linting.

### Development Setup

Run both the backend and frontend in separate terminals:

```bash
# Terminal 1: Start the Go backend (port 8080)
go run ./cmd/api-style serve --port 8080

# Terminal 2: Start the Vite dev server (port 3000)
cd web && pnpm dev
```

The Vite dev server proxies `/api/*` requests to the Go backend.

### Production Setup

Build the frontend and serve everything from the Go server:

```bash
# Build the frontend
cd web && pnpm build

# Serve from Go (serves both API and static files)
go run ./cmd/api-style serve --port 3000 --web-dir ./web/dist
```

### API Endpoints

- `POST /api/lint` - Lint an OpenAPI specification
- `GET /api/profiles` - List available style profiles
- `GET /api/health` - Health check

## Future Enhancements

- [ ] Real-time linting as you type
- [ ] Monaco editor integration
- [ ] Side-by-side spec/results view
- [ ] Rule explanation tooltips
- [ ] Export results (JSON, SARIF)
- [ ] Profile comparison view
- [ ] Custom profile editor
