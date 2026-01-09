package excel

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type Parser interface {
	GetData() []TireData
	ClearData()
	Lock()
	Unlock()
}

type TireData struct {
	Name     string
	Quantity int
	Year     int
	Country  string
	Price    float64
}

func SaveToExcel(data []TireData, tableName string) error {
	if len(data) == 0 {
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

	for i, item := range data {
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

	fmt.Printf("   ✅ %s.xlsx - збережено %d товарів\n", tableName, len(data))
	return nil
}
