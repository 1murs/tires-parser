package parser

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"sync"
	"tires-parser/internal/config"
	"tires-parser/internal/models"
	"tires-parser/internal/storage"
)

type TiresParser struct {
	Data            []models.TireData
	BadWords        []string
	DelItemWords    []string
	Categories      []models.Category
	PricePercentage float64
	Mu              sync.Mutex
}

func New(categories []models.Category, percentage float64) *TiresParser {
	return &TiresParser{
		Data:            make([]models.TireData, 0),
		BadWords:        storage.LoadWordsFromFile(config.BadWordsFile),
		DelItemWords:    storage.LoadWordsFromFile(config.DelWordsFile),
		Categories:      categories,
		PricePercentage: percentage,
	}
}

func (p *TiresParser) Run() error {
	var wg sync.WaitGroup

	for _, cat := range p.Categories {
		wg.Add(1)
		go func(url, name string) {
			defer wg.Done()
			p.scrapePages(url, name)
		}(cat.URL, cat.Name)
	}
	wg.Wait()

	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("✅ ПАРСИНГ ЗАВЕРШЕНО!")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("📁 Excel файли збережено в поточній папці\n")

	return nil
}

func (p *TiresParser) scrapePages(startURL, tableName string) error {
	currentURL := startURL
	pageCount := 0

	fmt.Printf("   📦 %s - обробка...\n", tableName)

	for currentURL != "" {
		pageCount++

		html, err := p.request(currentURL)
		if err != nil {
			fmt.Printf("   ❌ Помилка запиту: %v\n", err)
			break
		}

		nextPage, err := p.processHTML(html)
		if err != nil {
			fmt.Printf("   ❌ Помилка обробки: %v\n", err)
			break
		}

		if nextPage == "" {
			break
		}

		currentURL = config.BaseURL + nextPage
	}

	return p.saveToExcel(tableName)
}

func (p *TiresParser) saveToExcel(tableName string) error {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if len(p.Data) == 0 {
		fmt.Printf("   ⚠️  Немає даних для %s\n", tableName)
		return nil
	}

	f := excelize.NewFile()
	sheetName := "Sheet1"
	f.SetSheetName(sheetName, tableName)

	headers := []string{"Товар", "Кількість", "Рік", "Країна", "Ціна (евро)"}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(tableName, cell, header)
	}

	for i, item := range p.Data {
		row := i + 2
		f.SetCellValue(tableName, fmt.Sprintf("A%d", row), item.Name)
		f.SetCellValue(tableName, fmt.Sprintf("B%d", row), item.Quantity)
		if item.Year == 0 {
			f.SetCellValue(tableName, fmt.Sprintf("C%d", row), "")
		} else {
			f.SetCellValue(tableName, fmt.Sprintf("C%d", row), item.Year)
		}
		// f.SetCellValue(tableName, fmt.Sprintf("C%d", row), item.Year)
		f.SetCellValue(tableName, fmt.Sprintf("D%d", row), item.Country)
		f.SetCellValue(tableName, fmt.Sprintf("E%d", row), item.Price)
	}

	filename := fmt.Sprintf("%s.xlsx", tableName)
	if err := f.SaveAs(filename); err != nil {
		return err
	}

	fmt.Printf("   ✅ %s.xlsx - збережено %d товарів\n", tableName, len(p.Data))
	p.Data = p.Data[:0]
	return nil
}
