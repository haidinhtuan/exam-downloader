# Exam Downloader

**Exam Downloader** is a powerful CLI tool designed to scrape and download exam questions from online exam discussion sites. It bypasses paywall barriers and pagination by aggregating community discussions, questions, and exhibits into a single, clean, and easy-to-study file.

## 🚀 Features

- **Complete Scrape**: Downloads questions, options, **suggested answers**, and even exhibits (images).
- **Clean Output**: Automatically formats questions, removes clutter ("Actual Exam..." headers), and organizes them sequentially.
- **Multiple Formats**: Export to **Markdown** (`.md`), **PDF** (`.pdf`), **HTML** (`.html`), or **Text** (`.txt`).
- **Smart Caching**: Uses a community GitHub cache for instant results, with a fallback to live scraping for the freshest data.
- **Docker Ready**: Run immediately without installing Go or dependencies.

---

## 🛠️ Installation & Setup

### Option 1: Docker (Recommended)

The easiest way to run the tool without installing anything else.

1.  **Pull the image:**
    ```bash
    docker pull haidinhtuan/exam-downloader:latest
    ```

2.  **Run to download questions:**
    ```bash
    docker run -it --rm \
      -v $(pwd):/app/output \
      haidinhtuan/exam-downloader:latest \
      -p google -s devops -o /app/output/google-devops.md
    ```
    *(This command mounts your current folder so the file is saved directly to your computer.)*

### Option 2: Build from Source

1.  **Install Go**: Ensure you have [Go 1.24+](https://go.dev/doc/install).
2.  **Clone & Run**:
    ```bash
    git clone https://github.com/tdinh/exam-downloader
    cd exam-downloader
    go run ./cmd/main.go -p google -s devops
    ```

---

## 📖 How to Use

The general syntax is:
```bash
# If using binary/source:
./exam-downloader -p <provider> -s <search-term> [flags]

# If using Docker:
docker run ... [image] -p <provider> -s <search-term> [flags]
```

### Examples

**1. The Standard Run (Markdown)**
Get the "Google DevOps" exam questions. Best for most users.
```bash
go run ./cmd/main.go -p google -s devops -o google-devops.md
```

**2. Get a PDF for Printing**
Download Cisco CCNA (200-301) questions directly as a PDF.
```bash
go run ./cmd/main.go -p cisco -s 200-301 -o cisco-ccna.md -type pdf
```

**3. Force Live Scraping**
If the cached data is old or missing images, force the tool to visit the site directly.
```bash
go run ./cmd/main.go -p microsoft -s az-900 -no-cache
```

**4. See All Exams for a Provider**
Not sure of the exam code? List them all.
```bash
go run ./cmd/main.go -p amazon -exams
```

### Flags & Options

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-p` | `google` | **Required.** The provider name (e.g., `amazon`, `cisco`, `juniper`, `microsoft`). |
| `-s` | `""` | **Required.** The search term (e.g., `devops`, `200-301`, `az-900`). |
| `-o` | `output.md` | The output filename. |
| `-type` | `md` | Output format: `md` (Markdown), `pdf`, `html`, `txt`. |
| `-no-cache` | `false` | Ignore the cache and scrape the live website (slower but fresher). |
| `-c` | `false` | Include the community discussion comments (can be very long). |
| `-save-links`| `false` | Save a list of all question URLs to `saved-links.txt`. |
| `-t` | `""` | GitHub Token (optional) to increase API limits for cached downloads. |

---

## 💡 Tips & Troubleshooting

- **Images/Exhibits**: For the best experience with diagrams and charts, output to **Markdown (`.md`)** or **PDF**. The tool automatically embeds them.
- **"No questions found"**: 
  - Try using `-no-cache`. The cache might be missing that specific exam.
  - Check the spelling of your search term `-s`.
- **Rate Limiting**: If you are scraping live (`-no-cache`) for a huge exam (500+ questions), it might take a while. The tool has built-in delays to be polite.

---

## ⚖️ Disclaimer

This tool is for **educational purposes only**. It is intended to help students consolidate their study materials. Please respect the Terms of Service of the websites you visit.
