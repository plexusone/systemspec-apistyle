// Type definitions for systemspec-apistyle Web UI

export type Severity = 'error' | 'warn' | 'info' | 'hint';
export type Status = 'pass' | 'fail';

export interface Violation {
  ruleId: string;
  severity: Severity;
  message: string;
  path: string;
  line?: number;
  column?: number;
}

export interface Summary {
  errors: number;
  warnings: number;
  infos: number;
  hints: number;
  total: number;
}

export interface LintResult {
  status: Status;
  violations: Violation[];
  summary: Summary;
  profile: string;
}

export interface Profile {
  name: string;
  description: string;
  version: string;
  ruleCount: number;
}

export interface Rule {
  id: string;
  title: string;
  category: string;
  severity: Severity;
  rationale?: string;
}

export interface AnalysisResult {
  decision: 'GO' | 'NO-GO';
  summary: string;
  lint?: LintResult;
  evaluation?: EvaluationResult;
}

export interface EvaluationResult {
  status: Status;
  findings: Finding[];
  categories: CategoryScore[];
}

export interface Finding {
  ruleId: string;
  passed: boolean;
  score: number;
  reasoning: string;
  suggestions?: string[];
}

export interface CategoryScore {
  name: string;
  score: number;
}
