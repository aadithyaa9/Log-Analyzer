package parser

import "strings"

// DetectParser automatically detects the appropriate parser for a log line
func DetectParser(line string) LogParser {
	line = strings.TrimSpace(line)

	// Try JSON parser first
	jsonParser := &JSONLogParser{}
	if jsonParser.CanParse(line) {
		return jsonParser
	}

	// Fallback to plain text parser
	return &PlainTextLogParser{}
}

// GetParser returns a parser based on the specified type
func GetParser(parserType ParserType) LogParser {
	switch parserType {
	case JSONParser:
		return &JSONLogParser{}
	case PlainTextParser:
		return &PlainTextLogParser{}
	default:
		return &PlainTextLogParser{}
	}
}
