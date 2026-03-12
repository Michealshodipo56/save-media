# Save Media - High Quality Video Downloader

Save Media is a professional, high-performance web application designed to download videos and audio from virtually any social media platform, including YouTube, Facebook, Instagram, TikTok, and many more.

![Save Media Logo](images/cloud.svg)

## 🚀 Features

- **Multi-Platform Support**: Works with YouTube, Instagram, Facebook, TikTok, X (Twitter), and hundreds more.
- **High Quality**: Choose your preferred resolution, from 144p up to 4K/8K (depending on source).
- **Audio Extraction**: Instantly convert any video to a high-quality MP3.
- **Dynamic Naming**: Downloads are automatically named after the video title for easy organization.
- **CORS-Aware Thumbnail Proxy**: Handles image loading across different domains smoothly.
- **Responsive Design**: Optimized for both mobile devices and desktop computers.
- **Clean Architecture**: Refactored for maintainability with separated CSS, JS, and Backend code.

## 🛠 Tech Stack

- **Frontend**: HTML5, Vanilla CSS3 (Custom Design System), JavaScript (ES6+), [GSAP](https://gsap.com/) for animations.
- **Backend**: [Go (Golang)](https://golang.org/) - Lightweight, concurrent HTTP server.
- **Engine**: [yt-dlp](https://github.com/yt-dlp/yt-dlp) - The industry-standard command-line media downloader.

## 📦 Installation & Setup

### Prerequisites

- [Go](https://go.dev/doc/install) installed on your system.
- [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) installed and available in your PATH.
- [ffmpeg](https://ffmpeg.org/) (required by yt-dlp for merging video and audio).

### Running the Application

1. **Clone the repository**:

   ```bash
   git clone https://github.com/Michealshodipo56/save-media.git
   cd save-media
   ```

2. **Start the Go Backend**:

   ```bash
   go run main.go
   ```

   The server will start on `http://localhost:8080`.

3. **Open the Frontend**:
   Simply open `index.html` in any modern web browser.

## 📖 Usage

1. Copy the URL of the video you want to download.
2. Paste it into the input field on the Save Media homepage.
3. Click "Fetch Video" to analyze the link.
4. Select your desired format and quality from the dropdowns.
5. Click "Download Now" to save the file.

## 👨‍💻 Developed By

**Shodipo Micheal**

- [Facebook](https://www.facebook.com/michealshodipo56)
- [GitHub](https://github.com/Michealshodipo56)

---

© 2026 Shodipo Micheal. All rights reserved.
