package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

const (
	CategoriesFile = "categories.json"
	BaseURL        = "https://rengasketola.fi/"
	maxWorkers     = 5
)

type Category struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type TireData struct {
	Name     string
	Quantity int
	Year     int
	Country  string
	Price    float64
}

type TiresParser struct {
	data            []TireData
	badWords        []string
	delItemWords    []string
	categories      []Category
	pricePercentage float64
	mu              sync.Mutex
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		printMenu()
		choice := readInput(reader, "Ваш вибір: ")

		switch choice {
		case "1":
			addCategory(reader)
		case "2":
			listCategories()
		case "3":
			removeCategory(reader)
		case "4":
			startParsing(reader)
		case "5":
			fmt.Println("\n👋 До побачення!")
			return
		default:
			fmt.Println("\n❌ Невірний вибір. Спробуйте ще раз.\n")
		}

	}

}

func printMenu() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║     ПАРСЕР ШИН - ГОЛОВНЕ МЕНЮ         ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  1 ➜ Додати категорію")
	fmt.Println("  2 ➜ Показати всі категорії")
	fmt.Println("  3 ➜ Видалити категорію")
	fmt.Println("  4 ➜ ЗАПУСТИТИ ПАРСИНГ")
	fmt.Println("  5 ➜ Вихід")
	fmt.Println()
}

func readInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func loadCategories() []Category {
	file, err := os.Open(CategoriesFile)
	if err != nil {
		return []Category{}
	}
	defer file.Close()

	var categories []Category
	json.NewDecoder(file).Decode(&categories)
	return categories
}

func saveCategories(categories []Category) error {
	file, err := os.Create(CategoriesFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(categories)

}

func addCategory(reader *bufio.Reader) {
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("         ДОДАТИ НОВУ КАТЕГОРІЮ")
	fmt.Println("─────────────────────────────────────────")
	url := readInput(reader, "📎 Введіть URL категорії: ")
	if url == "" {
		fmt.Println("❌ URL не може бути пустим")
		return
	}
	name := readInput(reader, "📝 Введіть назву категорії: ")
	if name == "" {
		fmt.Println("❌ Назва не може бути пустою")
		return
	}
	categories := loadCategories()
	categories = append(categories, Category{URL: url, Name: name})

	if err := saveCategories(categories); err != nil {
		fmt.Printf("❌ Помилка: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Категорію '%s' успішно додано!\n\n", name)
}

func listCategories() {
	categories := loadCategories()

	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("         СПИСОК КАТЕГОРІЙ")
	fmt.Println("─────────────────────────────────────────")

	if len(categories) == 0 {
		fmt.Println("📭 Категорій поки немає")
		fmt.Println("💡 Додайте першу категорію (пункт 1)\n")
		return
	}

	for i, cat := range categories {
		fmt.Printf("\n%d. 📦 %s\n", i+1, cat.Name)
		fmt.Printf("   🔗 %s\n", cat.URL)
	}
	fmt.Println()
}

func removeCategory(reader *bufio.Reader) {
	categories := loadCategories()

	if len(categories) == 0 {
		fmt.Println("\n📭 Немає категорій для видалення\n")
		return
	}

	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("         ВИДАЛИТИ КАТЕГОРІЮ")
	fmt.Println("─────────────────────────────────────────")

	for i, cat := range categories {
		fmt.Printf("%d. %s\n", i+1, cat.Name)
	}

	input := readInput(reader, "\n🗑️  Введіть номер для видалення (0 - відміна): ")

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 0 || choice > len(categories) {
		fmt.Println("❌ Невірний номер\n")
		return
	}

	if choice == 0 {
		fmt.Println("↩️  Відмінено\n")
		return
	}

	choice--
	removed := categories[choice]
	categories = append(categories[:choice], categories[choice+1:]...)

	if err := saveCategories(categories); err != nil {
		fmt.Printf("❌ Помилка: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Категорію '%s' видалено!\n\n", removed.Name)

}

func startParsing(reader *bufio.Reader) {
	categories := loadCategories()

	if len(categories) == 0 {
		fmt.Println("\n❌ Спочатку додайте категорії (пункт 1)\n")
		return
	}

	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("         НАЛАШТУВАННЯ ПАРСИНГУ")
	fmt.Println("─────────────────────────────────────────")

	fmt.Printf("\n📋 Буде оброблено категорій: %d\n", len(categories))

	defaultPercentage := 9.0

	input := readInput(reader, fmt.Sprintf("\n💰 Відсоток додавання до ціни (Enter = %.0f%%): ", defaultPercentage))

	percentage := defaultPercentage

	if input != "" {
		if val, err := strconv.ParseFloat(input, 64); err == nil {
			percentage = val
		}
	}

	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🚀 ПОЧАТОК ПАРСИНГУ...")
	fmt.Println("─────────────────────────────────────────\n")

	parser := NewTiresParser(categories, percentage)

	parser.Run()

}

func NewTiresParser(categories []Category, percentage float64) *TiresParser {
	return &TiresParser{
		data:            make([]TireData, 0),
		badWords:        loadWordsFromFile("bad_words.txt"),
		delItemWords:    loadWordsFromFile("del_item_words.txt"),
		categories:      categories,
		pricePercentage: percentage,
	}

}

func loadWordsFromFile(fileName string) []string {
	file, err := os.Open(fileName)
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

func (p *TiresParser) Run() error {
	var wg sync.WaitGroup

	for _, cat := range p.categories {
		wg.Add(1)
		go func(url, name string) {
			defer wg.Done()
			p.scrapePages(url, name)
		}(cat.URL, cat.Name)
	}
	wg.Wait()
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
			fmt.Printf("   ❌ Помилка обробки: %v\n", err)
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
		currentURL = BaseURL + nextPage

	}
	return p.saveToExcel(tableName)
}

func (p *TiresParser) request(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en,uk;q=0.9")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil

}

func (p *TiresParser) processHTML(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	blockContent := doc.Find(".mt-0")
	if blockContent.Length() == 0 {
		return "", nil
	}

	dotRegex := regexp.MustCompile(`DOT(\d{4})`)
	priceRegex := regexp.MustCompile(`[^\d.]`)

	blockContent.Find(".tp-product-item-grid-1").Each(func(i int, item *goquery.Selection) {
		nameText := item.Find(".tp-product-title").Text()
		nameWords := strings.Fields(nameText)

		filteredName, shouldDelete := p.checkItemName(nameWords)
		if shouldDelete {
			return
		}

		name := strings.Join(filteredName, " ")

		var year int
		if match := dotRegex.FindStringSubmatch(name); len(match) > 1 {
			year, _ = strconv.Atoi(match[1])
		}

		priceText := item.Find("span .oe_currency_value").Text()
		priceText = strings.ReplaceAll(priceText, ",", ".")
		priceText = strings.ReplaceAll(priceText, "\u00a0", "")
		priceText = priceRegex.ReplaceAllString(priceText, "")

		price, err := strconv.ParseFloat(priceText, 64)
		if err != nil {
			return
		}

		finalPrice := price*(1+p.pricePercentage/100) + 20

		p.mu.Lock()
		p.data = append(p.data, TireData{
			Name:     name,
			Quantity: 8,
			Year:     year,
			Country:  "",
			Price:    roundFloat(finalPrice, 2),
		})
		p.mu.Unlock()
	})

	nextPage, exists := doc.Find("a.tp-load-more-on-scroll").Attr("href")
	if exists {
		return nextPage, nil
	}

	return "", nil
}

func roundFloat(val float64, precision int) float64 {
	ratio := float64(1)
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
}

func (p *TiresParser) saveToExcel(tableName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.data) == 0 {
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

	for i, item := range p.data {
		row := i + 2
		f.SetCellValue(tableName, fmt.Sprintf("A%d", row), item.Name)
		f.SetCellValue(tableName, fmt.Sprintf("B%d", row), item.Quantity)
		f.SetCellValue(tableName, fmt.Sprintf("C%d", row), item.Year)
		f.SetCellValue(tableName, fmt.Sprintf("D%d", row), item.Country)
		f.SetCellValue(tableName, fmt.Sprintf("E%d", row), item.Price)
	}
	filename := fmt.Sprintf("%s.xlsx", tableName)
	if err := f.SaveAs(filename); err != nil {
		return err
	}
	fmt.Printf("   ✅ %s.xlsx - збережено %d товарів\n", tableName, len(p.data))
	p.data = make([]TireData, 0)
	return nil
}

func (p *TiresParser) checkItemName(itemName []string) ([]string, bool) {
	filtered := make([]string, 0)

	for _, word := range itemName {
		isBad := false
		for _, badWord := range p.badWords {
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
	for _, delWord := range p.delItemWords {
		if strings.Contains(nameStr, delWord) {
			return nil, true
		}
	}

	return filtered, false

}
