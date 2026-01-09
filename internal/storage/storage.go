package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"tires-parser/internal/models"
)

func LoadCategories(filename string) []models.Category {
	file, err := os.Open(filename)
	if err != nil {
		return []models.Category{}
	}
	defer file.Close()

	var categories []models.Category
	json.NewDecoder(file).Decode(&categories)
	return categories
}

func SaveCategories(filename string, categories []models.Category) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(categories)
}

func LoadWordsFromFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}
	return words
}
