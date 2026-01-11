# Tires Parser

A Go-based web scraper for extracting tire product information from rengasketola.fi and exporting data to Excel files.

## Features

- Multi-threaded scraping for improved performance
- Automatic studded tire detection and marking
- Customizable price markup calculation
- Excel export with structured data
- Category management system
- Word filtering for data cleaning

## Installation

### Prerequisites

- Go 1.25.1 or higher

### Setup

1. Clone the repository:

```bash
git clone <repository-url>
cd tires-parser
```

1. Install dependencies:

```bash
go mod download
```

1. Create required configuration files in the project root:
   - `bad_words.txt` - Words to filter from product names
   - `del_item_words.txt` - Keywords that trigger product exclusion

## Usage

### Running the Application

```bash
go run cmd/tires-parser/main.go
```

### Menu Options

1. **Add Category** - Add new product categories to scrape
2. **List Categories** - Display all saved categories
3. **Remove Category** - Delete categories from the list
4. **Start Parsing** - Begin scraping process
5. **Exit** - Close the application

### Parsing Process

1. The parser first collects all studded tire names from the studded tires category
2. Then scrapes configured categories in parallel
3. Compares each tire against the studded tires database
4. Adds "-STUD" suffix to matching studded tires
5. Exports results to separate Excel files per category

### Price Calculation

The default price markup is 9%. During parsing, you can specify a custom percentage.

Final price formula:

```
final_price = (original_price * (1 + percentage/100)) + 20
```

## Output

Excel files are generated in the project root directory with the following columns:

- Product Name (with -STUD suffix for studded tires)
- Quantity
- Year (DOT code year)
- Country
- Price (EUR)

## Configuration

### Categories File

Categories are stored in `categories.json`:

```json
[
  {
    "url": "https://rengasketola.fi/category/example",
    "name": "Category Name"
  }
]
```

### Word Filters

- `bad_words.txt` - Single words to remove from product names (one per line)
- `del_item_words.txt` - Phrases that exclude entire products (one per line)

## Project Structure

```
tires-parser/
├── cmd/tires-parser/
│   └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── excel/
│   │   └── excel.go
│   ├── models/
│   │   └── models.go
│   ├── parser/
│   │   ├── parser.go
│   │   ├── scraper.go
│   │   └── filters.go
│   ├── storage/
│   │   └── storage.go
│   └── ui/
│       └── menu.go
├── go.mod
├── go.sum
└── README.md
```

## Dependencies

- `github.com/PuerkitoBio/goquery` - HTML parsing
- `github.com/xuri/excelize/v2` - Excel file generation

## License

This project is provided as-is for educational purposes.
