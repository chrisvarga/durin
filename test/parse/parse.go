package token

import (
    "strings"
    "unicode"
)

func Parse1(s string) string {
	var command string
	var key string
	var value string

	request := strings.Split(s, " ")
	for i, c := range request {
		if i == 0 {
			command = c
		} else if i == 1 {
			key = c
		} else if i == 2 {
			value = s[strings.Index(s, c):]
		} else {
			break
		}
	}
	command = strings.TrimSuffix(command, "\n")
	key = strings.TrimSuffix(key, "\n")
	value = strings.TrimSuffix(value, "\n")

	switch command {
	case "set", "get", "del":
		if key == "" {
			return "(error) no key specified"
		}
	case "keys":
	default:
		return "(error) invalid command"
	}
	return value
}

func Parse2(message string) string {
	var command string
	var key string
	var value string

	// Parse command
	if len(message) < 4 {
		return "(error) invalid syntax"
	}
	switch message[0:4] {
	case "set ", "get ", "del ":
		command = message[0:3]
	case "keys":
		command = message[0:4]
	default:
		return "(error) invalid command"
	}

	// Parse key
	if command == "keys" {
		if idx := strings.IndexByte(message[4:], '\n'); idx >= 0 {
			if idx >= 1 && message[0:5] != "keys " {
				return "(error) invalid command"
			}
			key = strings.TrimSuffix(message[5:], "\n")
			key = strings.ReplaceAll(key, " ", "")
			if len(key) != 0 && len(key)+1 != idx {
				return "(error) invalid keys argument"
			}
		}
	} else {
		if idx := strings.IndexByte(message[4:], ' '); idx >= 0 {
			// set command
			key = message[4 : idx+4]
		} else {
			// get or del command, need to trim newline
			key = strings.TrimSuffix(message[4:], "\n")
		}
	}
	if len(key) == 0 && command != "keys" {
		return "(error) invalid key"
	}

	// If we got this far, the rest of the message is the value.
	if command == "set" {
		value = strings.TrimSpace(message[len(command)+len(key)+2:])
		if len(value) == 0 {
			return "(error) invalid value"
		}
	}
	return value
}

func Spaces(s string) []int {
	runes := []rune(s)
	var spaces []int
	numSpaces := 0
	for i := 0; i < len(runes); i++ {
		if unicode.IsSpace(runes[i]) {
			spaces = append(spaces, i)
			numSpaces += 1
			if numSpaces == 2 {
				break
			}
		}
	}
	return spaces
}

func Parse3(message string) (string, string, string, string) {
	var command strings.Builder
	var key strings.Builder
	var value strings.Builder

	spaces := Spaces(message)
	switch len(spaces) {
	case 0:
		// keys command
		command.WriteString(message)
		return command.String(), "", "", ""
	case 1:
		// get, del, or keys [prefix] command
		command.WriteString(message[:spaces[0]])
		key.WriteString(message[spaces[0]+1:])
		return command.String(), key.String(), "", ""
	case 2:
		// set command
		command.WriteString(message[:spaces[0]])
		key.WriteString(message[spaces[0]+1 : spaces[1]])
		value.WriteString(message[spaces[1]+1:])
		if len(strings.TrimSpace(value.String())) == 0 && command.String() == "set" {
			return "", "", "", "(error) empty value"
		}
		return command.String(), key.String(), strings.TrimSuffix(value.String(), "\n"), ""
	default:
		return "", "", "", "(error) invalid syntax"
	}
}
