// Package agent provides self-maintenance capabilities for LUCID.
//
// Deprecated: ContextMaintainer is superseded by the Signal→Coordinator→Executor
// pipeline implemented via react.Engine agents. All context maintenance now flows
// through AgentService.ProcessSignal(). This file is retained only for types that
// may be referenced by external packages during migration.
package agent
