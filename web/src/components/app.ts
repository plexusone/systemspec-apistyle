import { LitElement, html, css } from 'lit';
import { customElement, state } from 'lit/decorators.js';
import type { LintResult } from '../types';
import { api } from '../api';

@customElement('api-style-app')
export class ApiStyleApp extends LitElement {
  static override styles = css`
    :host {
      display: block;
      min-height: 100vh;
    }

    .app {
      display: flex;
      flex-direction: column;
      min-height: 100vh;
    }

    header {
      background: var(--color-bg-secondary);
      border-bottom: 1px solid var(--color-border);
      padding: var(--spacing-md) var(--spacing-xl);
    }

    .header-content {
      max-width: 1400px;
      margin: 0 auto;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .logo {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);
      font-weight: 600;
      font-size: 1.25rem;
      color: var(--color-primary);
      text-decoration: none;
    }

    .logo-icon {
      width: 32px;
      height: 32px;
    }

    nav {
      display: flex;
      gap: var(--spacing-lg);
    }

    nav a {
      color: var(--color-text-secondary);
      text-decoration: none;
      font-size: 0.875rem;
      font-weight: 500;
      transition: color 0.15s;
    }

    nav a:hover {
      color: var(--color-primary);
    }

    main {
      flex: 1;
      max-width: 1400px;
      margin: 0 auto;
      padding: var(--spacing-xl);
      width: 100%;
    }

    .layout {
      display: grid;
      grid-template-columns: 1fr 400px;
      gap: var(--spacing-xl);
      align-items: start;
    }

    @media (max-width: 1024px) {
      .layout {
        grid-template-columns: 1fr;
      }
    }

    .editor-section {
      display: flex;
      flex-direction: column;
      gap: var(--spacing-md);
    }

    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .section-title {
      font-size: 1.125rem;
      font-weight: 600;
    }

    .results-section {
      position: sticky;
      top: var(--spacing-xl);
    }

    footer {
      background: var(--color-bg-secondary);
      border-top: 1px solid var(--color-border);
      padding: var(--spacing-md) var(--spacing-xl);
      text-align: center;
      color: var(--color-text-secondary);
      font-size: 0.875rem;
    }

    footer a {
      color: var(--color-primary);
      text-decoration: none;
    }

    footer a:hover {
      text-decoration: underline;
    }
  `;

  @state()
  private spec: string = '';

  @state()
  private profile: string = 'default';

  @state()
  private lintResult: LintResult | null = null;

  @state()
  private loading: boolean = false;

  @state()
  private error: string | null = null;

  override render() {
    return html`
      <div class="app">
        <header>
          <div class="header-content">
            <a href="/" class="logo">
              <svg class="logo-icon" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"></path>
              </svg>
              systemspec-apistyle
            </a>
            <nav>
              <a href="https://github.com/plexusone/systemspec-apistyle" target="_blank">GitHub</a>
              <a href="/docs">Docs</a>
            </nav>
          </div>
        </header>

        <main>
          <div class="layout">
            <div class="editor-section">
              <div class="section-header">
                <h2 class="section-title">OpenAPI Specification</h2>
                <profile-selector
                  .value=${this.profile}
                  @profile-change=${this.handleProfileChange}
                ></profile-selector>
              </div>
              <spec-editor
                .value=${this.spec}
                @spec-change=${this.handleSpecChange}
                @lint-request=${this.handleLintRequest}
              ></spec-editor>
            </div>

            <div class="results-section">
              <lint-results
                .result=${this.lintResult}
                .loading=${this.loading}
                .error=${this.error}
              ></lint-results>
            </div>
          </div>
        </main>

        <footer>
          Built with <a href="https://lit.dev" target="_blank">Lit</a> |
          <a href="https://github.com/plexusone/systemspec-apistyle" target="_blank">View on GitHub</a>
        </footer>
      </div>
    `;
  }

  private handleSpecChange(e: CustomEvent<string>) {
    this.spec = e.detail;
  }

  private handleProfileChange(e: CustomEvent<string>) {
    this.profile = e.detail;
  }

  private async handleLintRequest() {
    if (!this.spec.trim()) {
      this.error = 'Please enter an OpenAPI specification';
      return;
    }

    this.loading = true;
    this.error = null;

    try {
      // Call the backend API
      this.lintResult = await api.lint(this.spec, this.profile);
    } catch (err) {
      this.error = err instanceof Error ? err.message : 'An error occurred';
      this.lintResult = null;
    } finally {
      this.loading = false;
    }
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'api-style-app': ApiStyleApp;
  }
}
