package fetch

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGetDataFromDoc_WithImage(t *testing.T) {
	html := `
<!DOCTYPE html>
<html>
<body>
    <h1>Test Title</h1>
    <div class="question-discussion-header">
        Question 1
    </div>
    <div class="card-text">
        This is the question text.
        <img class="in-question-image" src="/assets/media/exam-media/04232/0000100001.jpg">
    </div>
    <ul>
        <li class="multi-choice-item">Option A</li>
        <li class="multi-choice-item">Option B</li>
    </ul>
    <div class="correct-answer">A</div>
    <div class="discussion-meta-data"><i>2023-10-27</i></div>
    <div class="discussion-container">Some comments</div>
</body>
</html>
`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse mock HTML: %v", err)
	}

	link := "https://www.examtopics.com/discussions/provider/1"
	data := getDataFromDoc(doc, link)

	if data == nil {
		t.Fatal("getDataFromDoc returned nil")
	}

	// Check if image src is present in Content
	expectedImageSrc := "https://www.examtopics.com/assets/media/exam-media/04232/0000100001.jpg"
	if !strings.Contains(data.Content, expectedImageSrc) {
		t.Errorf("Content should contain image src %q, but got:\n%s", expectedImageSrc, data.Content)
	}
}
