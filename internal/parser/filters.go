package parser

import (
	"regexp"
	"strings"
)

func (p *TiresParser) checkItemName(itemName []string) ([]string, bool) {
	filtered := make([]string, 0)

	for _, word := range itemName {
		isBad := false
		for _, badWord := range p.BadWords {
			if word == badWord {
				isBad = true
				break
			}
		}
		if !isBad {
			filtered = append(filtered, word)
		}
	}

	nameStr := strings.Join(filtered, " ")
	for _, delWord := range p.DelItemWords {
		if strings.Contains(nameStr, delWord) {
			return nil, true
		}
	}

	return filtered, false
}

func roundFloat(val float64, precision int) float64 {
	ratio := float64(1)
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}

func (p *TiresParser) normalizeName(name string) string {
	dotRegex := regexp.MustCompile(`DOT\d{4}`)
	normalized := dotRegex.ReplaceAllString(name, "")

	normalized = strings.Join(strings.Fields(normalized), " ")
	normalized = strings.TrimSpace(normalized)

	return strings.ToLower(normalized)
}
