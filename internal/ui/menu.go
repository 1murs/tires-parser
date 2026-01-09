package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"tires-parser/internal/config"
	"tires-parser/internal/models"
	"tires-parser/internal/parser"
	"tires-parser/internal/storage"
)

func Run() {
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

	categories := storage.LoadCategories(config.CategoriesFile)
	categories = append(categories, models.Category{URL: url, Name: name})

	if err := storage.SaveCategories(config.CategoriesFile, categories); err != nil {
		fmt.Printf("❌ Помилка: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Категорію '%s' успішно додано!\n\n", name)
}

func listCategories() {
	categories := storage.LoadCategories(config.CategoriesFile)

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
	categories := storage.LoadCategories(config.CategoriesFile)

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

	if err := storage.SaveCategories(config.CategoriesFile, categories); err != nil {
		fmt.Printf("❌ Помилка: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Категорію '%s' видалено!\n\n", removed.Name)
}

func startParsing(reader *bufio.Reader) {
	categories := storage.LoadCategories(config.CategoriesFile)

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

	p := parser.New(categories, percentage)
	p.Run()
}
