package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"examtopics-downloader/internal/fetch"
	"examtopics-downloader/internal/utils"
)

func main() {
	provider := flag.String("p", "google", "Name of the exam provider (default -> google)")
	grepStr := flag.String("s", "", "String to grep for in discussion links")
	outputPath := flag.String("o", "exam.md", "Optional path of the file where the data will be outputted")
	fileType := flag.String("type", "md", "Optionally include file type (default -> .md)")
	commentBool := flag.Bool("c", false, "Optionally include all the comment/discussion text")
	examsFlag := flag.Bool("exams", false, "Optionally show all the possible exams for your selected provider and exit")
	saveUrls := flag.Bool("save-links", false, "Optional argument to save unique links to questions")
	noCache := flag.Bool("no-cache", false, "Optional argument, set to disable looking through cached data on github")
	token := flag.String("t", "", "Optional argument to make cached requests faster to gh api")
	flag.Parse()

	if *examsFlag {
		exams := fetch.GetProviderExams(*provider)
		fmt.Printf("Exams for provider '%s'\n\n", *provider)
		for _, exam := range exams {
			fmt.Println(utils.AddToBaseUrl(exam))
		}
		os.Exit(0)
	}

	if *grepStr == "" {
		log.Printf("running without a valid string to search for with -s, (no_grep_str)!")
	}

	if !*noCache {
		fmt.Println("Attempting to fetch from cache...")
		links := fetch.GetCachedPages(*provider, *grepStr, *token)
		if len(links) > 0 {
			utils.WriteData(links, *outputPath, *commentBool, *fileType, *provider, *grepStr)
			fmt.Printf("Successfully saved cached output to %s (filetype: %s).\n", *outputPath, *fileType)
			os.Exit(0)
		}
		fmt.Println("Cache miss or empty. Switching to live scraping.")
	}

	fmt.Println("\n==================================================")
	fmt.Printf("  Exam Downloader - Live Scraping Mode\n")
	fmt.Printf("  Provider: %s\n", *provider)
	fmt.Printf("  Query:    %s\n", *grepStr)
	fmt.Println("==================================================")

	startTime := utils.StartTime()

	// Step 1: Discovery
	fmt.Println("Step 1/3: Discovering exam pages...")
	links := fetch.GetAllPages(*provider, *grepStr)

	// Step 2: Saving Links (Optional)
	if *saveUrls {
		fmt.Println("\nStep 2/3: Saving question links...")
		utils.SaveLinks("saved-links.txt", links)
		fmt.Printf("-> Saved %d links to 'saved-links.txt'\n", len(links))
	} else {
		fmt.Println("\nStep 2/3: Saving question links (Skipped)")
	}

	// Step 3: Writing Output
	fmt.Println("\nStep 3/3: Generating output file...")
	utils.WriteData(links, *outputPath, *commentBool, *fileType, *provider, *grepStr)

	fmt.Println("\n==================================================")
	fmt.Println("  Scraping Complete!")
	fmt.Println("==================================================")
	fmt.Printf("  Total Questions: %d\n", len(links))
	fmt.Printf("  Time Elapsed:    %s\n", utils.TimeSince(startTime))
	fmt.Printf("  Output File:     %s\n", *outputPath)
	fmt.Println("==================================================")
}
