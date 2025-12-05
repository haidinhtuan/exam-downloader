package fetch

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"exam-downloader/internal/constants"
	"exam-downloader/internal/models"
	"exam-downloader/internal/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb/v3"
)

func getDataFromLink(link string) *models.QuestionData {
	doc, err := ParseHTML(link, *client)
	if err != nil {
		log.Printf("Failed parsing HTML data from link: %v", err)
		return nil
	}

	return getDataFromDoc(doc, link)
}

func getDataFromDoc(doc *goquery.Document, link string) *models.QuestionData {
	var allQuestions []string
	doc.Find("li.multi-choice-item").Each(func(i int, s *goquery.Selection) {
		allQuestions = append(allQuestions, utils.CleanText(s.Text()))
	})

	answerText := strings.TrimSpace(doc.Find(".correct-answer").Text())
	// FIX: Keep the full answer string, just clean it up
	answer := strings.ReplaceAll(strings.ReplaceAll(answerText, "\n", ""), "\t", "")

	// Extract content text
	contentText := utils.CleanText(doc.Find(".card-text").Text())

	// Remove "Actual Exam..." and "All ... Questions" lines if they appear in the text
	lines := strings.Split(contentText, "\n")
	var cleanLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Actual Exam") || (strings.HasPrefix(trimmed, "All") && strings.Contains(trimmed, "Questions")) {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	contentText = strings.Join(cleanLines, "\n")

	// Extract and append images
	doc.Find(".card-text img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			if !strings.HasPrefix(src, "http") {
				src = utils.AddToBaseUrl(src)
			}
			contentText += fmt.Sprintf("\n\n![Exhibit](%s)", src)
		}
	})

	return &models.QuestionData{
		Title:        utils.CleanText(doc.Find("h1").Text()),
		Header:       "", // User requested to remove "Actual Exam..." header info
		Content:      contentText,
		Questions:    allQuestions,
		Answer:       answer,
		Timestamp:    utils.CleanText(doc.Find(".discussion-meta-data > i").Text()),
		QuestionLink: link,
		Comments:     utils.CleanText(doc.Find(".discussion-container").Text()),
	}
}

var counter int = 0 //start counter at 1
func getJSONFromLink(link string) []*models.QuestionData {
	initialResp := FetchURL(link, *client)

	var githubResp map[string]any
	err := json.Unmarshal(initialResp, &githubResp)
	if err != nil {
		log.Printf("error unmarshalling GitHub API response: %v", err)
		return nil
	}

	downloadURL, ok := githubResp["download_url"].(string)
	if !ok {
		log.Printf("couldn't find download_url in GitHub API response")
		return nil
	}

	jsonResp := FetchURL(downloadURL, *client)

	var content models.JSONResponse
	err = json.Unmarshal(jsonResp, &content)
	if err != nil {
		log.Printf("error unmarshalling the questions data: %v", err)
		return nil
	}

	fmt.Println("Processing content from:", downloadURL)

	var questions []*models.QuestionData

	if content.PageProps.Questions == nil {
		log.Printf("no questions found in JSON content")
		return nil
	}

	for _, q := range content.PageProps.Questions {
		var comments string
		for _, discussion := range q.Discussion {
			comments += fmt.Sprintf("[%s] %s\n", discussion.Poster, discussion.Content)
		}

		var choicesHeader string
		var keys []string
		for key := range q.Choices {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			choicesHeader += fmt.Sprintf("**%s:** %s\n\n", key, q.Choices[key])
		}

		counter++

		var contentImages string
		for _, img := range q.QuestionImages {
			contentImages += fmt.Sprintf("\n![Exhibit](%s)", img)
		}

		// Combine QuestionText and Images into Content
		fullContent := q.QuestionText + "\n" + contentImages

		questions = append(questions, &models.QuestionData{
			Title:        "Exam Question #" + strconv.Itoa(counter),
			Header:       "", // Clear header to match live scraping
			Content:      fullContent,
			Questions:    []string{choicesHeader},
			Answer:       q.Answer,
			Timestamp:    q.Timestamp,
			QuestionLink: q.URL,
			Comments:     utils.CleanText(comments),
		})
	}

	return questions
}

func fetchAllPageLinksConcurrently(providerName, grepStr string, numPages, concurrency int) []string {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	results := make(chan []string, numPages)

	// Custom Colorful Progress Bar
	tmpl := `{{green "Scraping Pages:"}} {{bar . "[" "=" ">" "_" "]"}} {{percent .}} {{speed .}}`
	bar := pb.ProgressBarTemplate(tmpl).Start(numPages)

	rateLimiter := utils.CreateRateLimiter(constants.RequestsPerSecond)
	defer rateLimiter.Stop()

	for i := 1; i <= numPages; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			<-rateLimiter.C

			url := fmt.Sprintf("https://www.examtopics.com/discussions/%s/%d", providerName, i)
			results <- getLinksFromPage(url, grepStr)
			bar.Increment()
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var all []string
	for res := range results {
		all = append(all, res...)
	}

	bar.Finish()
	return all
}

// Main concurrent page scraping logic
func GetAllPages(providerName string, grepStr string) []models.QuestionData {
	baseURL := fmt.Sprintf("https://www.examtopics.com/discussions/%s/", providerName)
	numPages := getMaxNumPages(baseURL)

	allLinks := fetchAllPageLinksConcurrently(providerName, grepStr, numPages, constants.MaxConcurrentRequests)

	unique := utils.DeduplicateLinks(allLinks)
	sortedLinks := utils.SortLinksByQuestionNumber(unique)

	var wg sync.WaitGroup
	sem := make(chan struct{}, constants.MaxConcurrentRequests)
	results := make([]*models.QuestionData, len(sortedLinks))

	// Custom Colorful Progress Bar for Downloading
	tmpl := `{{green "Downloading Questions:"}} {{bar . "[" "=" ">" "_" "]"}} {{percent .}} {{speed .}}`
	bar := pb.ProgressBarTemplate(tmpl).Start(len(sortedLinks))

	rateLimiter := utils.CreateRateLimiter(constants.RequestsPerSecond)
	defer rateLimiter.Stop()

	for i, link := range sortedLinks {
		wg.Add(1)
		url := utils.AddToBaseUrl(link)

		go func(i int, url string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			<-rateLimiter.C

			data := getDataFromLink(url)
			if data != nil {
				results[i] = data
			}
			bar.Increment()
		}(i, url)
	}

	wg.Wait()
	bar.Finish()

	// Filter out nil entries
	var finalData []models.QuestionData
	for _, entry := range results {
		if entry != nil {
			finalData = append(finalData, *entry)
		}
	}

	return finalData
}
