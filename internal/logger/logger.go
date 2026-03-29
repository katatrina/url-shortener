package logger

import (
	"log/slog"
	"os"
)

// Setup configures the global slog logger.
//
// JSON format is chosen for production readiness:
//   - Machine-parseable (Loki, ELK, CloudWatch can ingest directly)
//   - Structured fields enable precise querying
//   - Consistent format across all log entries
//
// For development, you could use slog.NewTextHandler instead for
// human-readable output. We'll use JSON from the start so you
// get used to reading it — and it's what production will use anyway.
func Setup() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		// AddSource adds file:line to every log entry.
		// Extremely helpful for debugging ("which log.Printf was this?")
		// but adds ~10% overhead. Enable in dev, consider disabling in prod.
		AddSource: true,

		// Level controls what gets logged.
		// DEBUG: very verbose, includes batch flush counts, cache operations
		// INFO:  normal operations, startup/shutdown
		// WARN:  recoverable issues (cache miss fallback, rate limit hit)
		// ERROR: things that need human attention
		//
		// In dev, use DEBUG. In production, INFO or WARN.
		Level: slog.LevelDebug,
	})

	slog.SetDefault(slog.New(handler))
}
