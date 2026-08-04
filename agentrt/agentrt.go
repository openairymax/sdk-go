// SPDX-FileCopyrightText: 2025-2026 SPHARX Ltd.
// SPDX-License-Identifier: AGPL-3.0-or-later OR Apache-2.0

package agentrt

import (
	"log"
	"os"
)

const (
	Version = "0.1.1"
	Author  = "SPHARX Ltd."
	License = "AGPL-3.0-or-later OR Apache-2.0"
)

var defaultLogger = log.New(os.Stderr, "[AgentOS] ", log.LstdFlags|log.Lshortfile)

func GetLogger() *log.Logger {
	return defaultLogger
}

func SetLogger(l *log.Logger) {
	if l != nil {
		defaultLogger = l
	}
}
