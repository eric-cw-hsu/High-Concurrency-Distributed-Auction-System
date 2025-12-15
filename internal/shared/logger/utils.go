package logger

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

// addCallerInfo automatically adds caller info and call stack for error/fatal
func addCallerInfo(fields map[string]interface{}, level LogLevel) map[string]interface{} {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	// Only capture caller/func for WARN or above to reduce noise
	if level == WARN || level == ERROR || level == FATAL {
		pc, file, line, ok := runtime.Caller(5) // 3: skip logger wrapper
		if ok {
			// Use canonical keys similar to common structured loggers (e.g. zap/logrus)
			fields["caller"] = fmt.Sprintf("%s:%d", file, line)
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				fields["func"] = fn.Name()
			}
		}
	}
	// Add call stack for error/fatal
	if level == ERROR || level == FATAL {
		fields["stacktrace"] = filterStack(debug.Stack())
	}
	return fields
}

// LogToConsole outputs log to console in a readable format
func LogToConsole(p LogPayload) {
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	// Add color to log level
	var colorLevel string
	switch p.Level {
	case DEBUG:
		colorLevel = "\033[36m[DEBUG]\033[0m" // Cyan
	case INFO:
		colorLevel = "\033[32m[INFO]\033[0m" // Green
	case WARN:
		colorLevel = "\033[33m[WARN]\033[0m" // Yellow
	case ERROR:
		colorLevel = "\033[31m[ERROR]\033[0m" // Red
	case FATAL:
		colorLevel = "\033[35m[FATAL]\033[0m" // Magenta
	default:
		colorLevel = fmt.Sprintf("[%s]", p.Level)
	}

	// Add color to service name (blue)
	colorService := fmt.Sprintf("\033[34m%s\033[0m", p.Service)

	// Main log line
	fmt.Printf("%s %s %s: %s\n", timestamp, colorLevel, colorService, p.Message)

	// Fields on separate lines if present
	if p.Fields != nil && len(p.Fields) > 0 {
		fieldsOutput := formatFieldsMultiline(p.Fields)
		if fieldsOutput != "" {
			fmt.Print(fieldsOutput)
		}
	}
}

// formatFieldsMultiline formats fields in a readable multi-line format
func formatFieldsMultiline(fields map[string]interface{}) string {
	if fields == nil || len(fields) == 0 {
		return ""
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var builder strings.Builder

	for _, k := range keys {
		v := fields[k]
		lowerKey := strings.ToLower(k)

		// Mask sensitive values
		switch lowerKey {
		case "password", "pwd", "secret", "token", "access_token", "auth_token", "authorization":
			builder.WriteString(fmt.Sprintf("  %s: <redacted>\n", k))
			continue
		}

		// For stacktrace, use multiple lines with proper indentation
		if lowerKey == "stack" || lowerKey == "stacktrace" {
			if str, ok := v.(string); ok {
				builder.WriteString(fmt.Sprintf("  %s:\n", k))
				// Indent each line of the stacktrace
				lines := strings.Split(str, "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						builder.WriteString(fmt.Sprintf("    %s\n", line))
					}
				}
			} else {
				builder.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
			}
			continue
		}

		// Handle other field types
		var sval string
		switch t := v.(type) {
		case string:
			sval = t // No quotes for cleaner output
		case fmt.Stringer:
			sval = t.String()
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			sval = fmt.Sprintf("%v", t)
		case float32, float64:
			sval = fmt.Sprintf("%.2f", t)
		case bool:
			sval = fmt.Sprintf("%t", t)
		default:
			if b, err := json.Marshal(v); err == nil {
				sval = string(b)
			} else {
				sval = fmt.Sprintf("%v", v)
			}
		}

		builder.WriteString(fmt.Sprintf("  %s: %s\n", k, sval))
	}

	return builder.String()
}

func filterStack(stack []byte) string {
	lines := strings.Split(string(stack), "\n")
	filtered := []string{}
	skipNext := false

	loggerFunctions := []string{
		"addCallerInfo",
		"dispatchLog",
		"(*Logger).log",
		"(*Logger).Debug",
		"(*Logger).Info",
		"(*Logger).Warn",
		"(*Logger).Error",
		"(*Logger).Fatal",
	}

	for _, line := range lines {
		shouldSkip := false

		for _, fn := range loggerFunctions {
			if strings.Contains(line, fn) {
				shouldSkip = true
				skipNext = true
				break
			}
		}

		if shouldSkip {
			continue
		}

		if skipNext {
			skipNext = false
			continue
		}

		filtered = append(filtered, line)
	}

	return strings.Join(filtered, "\n")
}
