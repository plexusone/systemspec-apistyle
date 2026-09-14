// API client for the systemspec-apistyle backend

import type { LintResult, Profile } from './types';

// Default API base URL - can be overridden for development
const API_BASE = import.meta.env.VITE_API_BASE || '';

/**
 * API client for systemspec-apistyle backend.
 */
export class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl;
  }

  /**
   * Lint an OpenAPI specification.
   */
  async lint(spec: string, profile: string = 'default'): Promise<LintResult> {
    const response = await fetch(`${this.baseUrl}/api/lint`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ spec, profile }),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  /**
   * Get available style profiles.
   */
  async getProfiles(): Promise<Profile[]> {
    const response = await fetch(`${this.baseUrl}/api/profiles`);

    if (!response.ok) {
      throw new Error(`Failed to fetch profiles: HTTP ${response.status}`);
    }

    return response.json();
  }

  /**
   * Health check.
   */
  async health(): Promise<{ status: string; version: string }> {
    const response = await fetch(`${this.baseUrl}/api/health`);

    if (!response.ok) {
      throw new Error(`Health check failed: HTTP ${response.status}`);
    }

    return response.json();
  }
}

// Singleton instance
export const api = new ApiClient();
