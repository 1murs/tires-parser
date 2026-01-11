package parser

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"tires-parser/internal/models"

	"github.com/PuerkitoBio/goquery"
)

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

		// Ne code for studdet tires

		normalizedName := p.normalizeName(name)

		isStudded := p.StuddedTires[normalizedName]

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

		finalPrice := price*(1+p.PricePercentage/100) + 20
		displayName := name
		if isStudded {
			displayName = name + " -STUD"
		}

		p.Mu.Lock()
		p.Data = append(p.Data, models.TireData{
			Name:     displayName,
			Quantity: 8,
			Year:     year,
			Country:  "",
			Price:    roundFloat(finalPrice, 2),
		})
		p.Mu.Unlock()
	})

	nextPage, exists := doc.Find("a.tp-load-more-on-scroll").Attr("href")
	if exists {
		return nextPage, nil
	}

	return "", nil
}
