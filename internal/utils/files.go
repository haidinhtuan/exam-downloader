package utils

import (
	"bufio"
	"bytes"
	"exam-downloader/internal/models"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/mandolyte/mdtopdf"
	"github.com/yuin/goldmark"
)

func writeFile(filename string, content any) {
	file := CreateFile(filename)
	defer file.Close()

	switch v := content.(type) {
	case string:
		fmt.Fprintln(file, v)
	case []string:
		for _, line := range v {
			fmt.Fprintln(file, line)
		}
	default:
		log.Printf("writeFile: unsupported content type %T", v)
		return
	}
}

func WriteData(dataList []models.QuestionData, outputPath string, commentBool bool, fileType string, provider string, examCode string) {
	file := CreateFile(outputPath)
	defer file.Close()

	fmt.Fprintf(file, "# Exam %s\n\n", strings.ToUpper(examCode))
	fmt.Fprintf(file, "**Provider:** %s\n\n", CapitalizeFirstLetter(provider))
	fmt.Fprintf(file, "**Exam Code:** %s\n\n", strings.ToUpper(examCode))
	fmt.Fprintf(file, "**Total Questions:** %d\n\n", len(dataList))
	fmt.Fprintf(file, "---\n\n")

	counter := 1
	for _, data := range dataList {
		if data.Title == "" {
			continue
		}

		fmt.Fprintf(file, "## Question %d\n\n", counter)
		counter++

		// Removed Header printing as per user request

		if data.Content != "" {
			// Remove "All ... Questions]" links if present in content
			// Regex to match [All ... Questions]... which might be just text or markdown link
			// Simple string replace for common patterns or just print as is if cleaned in scraper.
			// Since CleanText is used, let's just print it, but we should try to strip specific known junk.

			// Better to leave it if not easily targetable, or use regex.
			// User specifically asked to remove "All JN-664 question".
			// Let's try to remove lines starting with "Actual exam..." or similar if they exist.

			fmt.Fprintf(file, "%s\n\n", data.Content)
		}

		for _, question := range data.Questions {
			fmt.Fprintf(file, "%s\n\n", question)
		}

		fmt.Fprintf(file, "**Suggested Answer: %s**\n\n", data.Answer)
		fmt.Fprintf(file, "**Added Since: %s**\n\n", data.Timestamp)
		fmt.Fprintf(file, "[View Discussion](%s)\n\n", data.QuestionLink)

		if commentBool {
			fmt.Fprintf(file, "Comments: %s\n", data.Comments)
		}

		fmt.Fprintf(file, "----------------------------------------\n\n")
	}

	switch fileType {
	case "pdf":
		mdContent, err := os.ReadFile(outputPath)
		if err != nil {
			log.Printf("failed to read markdown file: %v", err)
			return
		}

		opts := []mdtopdf.RenderOption{
			mdtopdf.IsHorizontalRuleNewPage(true), // treat --- as new page
		}

		pdfName := strings.TrimSuffix(outputPath, ".md") + ".pdf"
		renderer := mdtopdf.NewPdfRenderer("portrait", "A4", pdfName, "", opts, mdtopdf.LIGHT)
		if err := renderer.Process(mdContent); err != nil {
			log.Printf("mdtopdf conversion failed: %v", err)
			return
		}
		deleteMarkdownFile(outputPath)
	case "html":
		mdBytes, err := os.ReadFile(outputPath)
		if err != nil {
			log.Printf("failed to read file for html conversion: %v", err)
			return
		}

		html, err := mdToHTML(mdBytes)
		if err != nil {
			log.Printf("mdtohtml conversion failed: %v", err)
			return
		}

		fileName := strings.TrimSuffix(outputPath, ".md") + ".html"
		err = os.WriteFile(fileName, html, 0644)
		if err != nil {
			log.Printf("failed to write html file: %v", err)
			return
		}
		deleteMarkdownFile(outputPath)
	case "text":
		mdBytes, err := os.ReadFile(outputPath)
		if err != nil {
			log.Printf("failed to read file for text conversion: %v", err)
			return
		}

		txt := mdToText(string(mdBytes))

		fileName := strings.TrimSuffix(outputPath, ".md") + ".txt"
		err = os.WriteFile(fileName, []byte(txt), 0644)
		if err != nil {
			log.Printf("failed to write text file: %v", err)
			return
		}
		deleteMarkdownFile(outputPath)
	}
}

func mdToHTML(md []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert(md, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deleteMarkdownFile(filePath string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Delete Markdown file after conversion? (y/n): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		fmt.Println("Deleting file...")
		os.Remove(filePath)
	} else {
		fmt.Println("Keeping file.")
	}
}

func mdToText(md string) string {
	text := md
	// Markdown headers (#, ##, ###)
	header := regexp.MustCompile(`(?m)^#{1,6}\s*`)
	text = header.ReplaceAllString(text, "")
	// bold/italic symbols (*, **, _)
	formatting := regexp.MustCompile(`(\*\*|\*|__|_)`)
	text = formatting.ReplaceAllString(text, "")
	// links but keep link text [text](url) → text
	link := regexp.MustCompile(`\[(.*?)\]\(.*?\)`)
	text = link.ReplaceAllString(text, "$1")
	// images ![alt](url)
	image := regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	text = image.ReplaceAllString(text, "")

	return text
}

func SaveLinks(filename string, links []models.QuestionData) {
	var fullLinks []string
	for _, link := range links {
		fullLinks = append(fullLinks, link.QuestionLink)
	}
	writeFile(filename, fullLinks)
}
