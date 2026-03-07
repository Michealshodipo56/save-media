package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

// ================== STRUCTS ==================

type VideoMetadata struct {
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	Platform  string `json:"platform"`
	Duration  string `json:"duration"`
	Error     string `json:"error,omitempty"`
}

type FormatEntry struct {
	FormatID string  `json:"format_id"`
	Ext      string  `json:"ext"`
	Quality  string  `json:"quality"`
	Filesize float64 `json:"filesize"`
}

type FormatsResponse struct {
	Formats []FormatEntry `json:"formats"`
	Error   string        `json:"error,omitempty"`
}

type YtDlpFormat struct {
	FormatID       string  `json:"format_id"`
	Ext            string  `json:"ext"`
	Height         int     `json:"height"`
	Filesize       float64 `json:"filesize"`
	FilesizeApprox float64 `json:"filesize_approx"`
	VCodec         string  `json:"vcodec"`
	ACodec         string  `json:"acodec"`
	FormatNote     string  `json:"format_note"`
	AudioExt       string  `json:"audio_ext"`
	VideoExt       string  `json:"video_ext"`
	Abr            float64 `json:"abr"`
	Tbr            float64 `json:"tbr"`
}

type YtDlpOutput struct {
	Title     string        `json:"title"`
	Thumbnail string        `json:"thumbnail"`
	Extractor string        `json:"extractor_key"`
	Duration  float64       `json:"duration"`
	Formats   []YtDlpFormat `json:"formats"`
}

// ================== HELPERS ==================

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func detectPlatform(rawURL string) string {
	lower := strings.ToLower(rawURL)
	switch {
	case strings.Contains(lower, "youtube.com") || strings.Contains(lower, "youtu.be"):
		return "YouTube"
	case strings.Contains(lower, "instagram.com"):
		return "Instagram"
	case strings.Contains(lower, "facebook.com") || strings.Contains(lower, "fb.watch"):
		return "Facebook"
	case strings.Contains(lower, "tiktok.com"):
		return "TikTok"
	case strings.Contains(lower, "twitter.com") || strings.Contains(lower, "x.com"):
		return "X (Twitter)"
	default:
		return "Other"
	}
}

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "N/A"
	}
	total := int(seconds)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func getFilesize(f YtDlpFormat) float64 {
	if f.Filesize > 0 {
		return f.Filesize
	}
	return f.FilesizeApprox
}

func getQualityLabel(f YtDlpFormat) string {
	if f.VCodec == "none" || f.VCodec == "" {
		// Audio-only format
		if f.Abr > 0 {
			return fmt.Sprintf("Audio %.0f kbps", f.Abr)
		}
		return "Audio"
	}
	if f.Height > 0 {
		return strconv.Itoa(f.Height) + "p"
	}
	if f.FormatNote != "" {
		return f.FormatNote
	}
	return "Unknown"
}

func getExtLabel(f YtDlpFormat) string {
	ext := strings.ToLower(f.Ext)
	if ext == "" || ext == "none" {
		if f.VCodec == "none" || f.VCodec == "" {
			return "m4a"
		}
		return "mp4"
	}
	return ext
}

// Run yt-dlp and parse output for a URL
func fetchYtDlpData(videoURL string) (*YtDlpOutput, error) {
	cmd := exec.Command(
		"yt-dlp",
		"--dump-json",
		"--no-playlist",
		"--skip-download",
		videoURL,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp error: %v", err)
	}
	var ytOut YtDlpOutput
	if err := json.Unmarshal(out, &ytOut); err != nil {
		return nil, fmt.Errorf("parse error: %v", err)
	}
	return &ytOut, nil
}

// ================== METADATA ENDPOINT ==================

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	videoURL := r.URL.Query().Get("url")
	if videoURL == "" {
		json.NewEncoder(w).Encode(VideoMetadata{Error: "No URL provided"})
		return
	}

	ytOut, err := fetchYtDlpData(videoURL)
	if err != nil {
		json.NewEncoder(w).Encode(VideoMetadata{Error: "Failed to fetch video info: " + err.Error()})
		return
	}

	// Use the direct thumbnail URL from yt-dlp
	thumbnail := ytOut.Thumbnail

	meta := VideoMetadata{
		Title:     ytOut.Title,
		Thumbnail: thumbnail,
		Platform:  detectPlatform(videoURL),
		Duration:  formatDuration(ytOut.Duration),
	}

	json.NewEncoder(w).Encode(meta)
}

// ================== FORMATS ENDPOINT ==================

func formatsHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	videoURL := r.URL.Query().Get("url")
	if videoURL == "" {
		json.NewEncoder(w).Encode(FormatsResponse{Error: "No URL provided"})
		return
	}

	ytOut, err := fetchYtDlpData(videoURL)
	if err != nil {
		json.NewEncoder(w).Encode(FormatsResponse{Error: "Failed to fetch formats: " + err.Error()})
		return
	}

	var entries []FormatEntry
	seen := make(map[string]bool)

	// First pass: collect video+audio combined formats (mp4, webm)
	for _, f := range ytOut.Formats {
		isVideo := f.VCodec != "none" && f.VCodec != ""
		isAudio := f.ACodec != "none" && f.ACodec != "" && f.AudioExt != "none"
		ext := getExtLabel(f)

		// Skip manifest formats
		if ext == "m3u8" || ext == "mpd" || strings.Contains(ext, "m3u") {
			continue
		}

		quality := getQualityLabel(f)
		key := ext + "|" + quality

		if seen[key] {
			continue
		}

		if isVideo && isAudio {
			seen[key] = true
			entries = append(entries, FormatEntry{
				FormatID: f.FormatID,
				Ext:      ext,
				Quality:  quality,
				Filesize: getFilesize(f),
			})
		}
	}

	// Second pass: if no combined video+audio, try video-only formats (yt-dlp can merge)
	if len(entries) == 0 {
		for _, f := range ytOut.Formats {
			isVideo := f.VCodec != "none" && f.VCodec != ""
			ext := getExtLabel(f)

			if ext == "m3u8" || ext == "mpd" || strings.Contains(ext, "m3u") {
				continue
			}

			quality := getQualityLabel(f)
			key := ext + "|" + quality

			if seen[key] || !isVideo {
				continue
			}
			seen[key] = true
			entries = append(entries, FormatEntry{
				FormatID: f.FormatID,
				Ext:      ext,
				Quality:  quality,
				Filesize: getFilesize(f),
			})
		}
	}

	// Third pass: audio-only formats (mp3, m4a, etc.)
	for _, f := range ytOut.Formats {
		isVideoOnly := f.VCodec != "none" && f.VCodec != ""
		isAudio := f.ACodec != "none" && f.ACodec != ""
		ext := getExtLabel(f)

		if ext == "m3u8" || ext == "mpd" || strings.Contains(ext, "m3u") || isVideoOnly {
			continue
		}

		quality := getQualityLabel(f)
		key := ext + "|" + quality

		if seen[key] || !isAudio {
			continue
		}
		seen[key] = true
		entries = append(entries, FormatEntry{
			FormatID: f.FormatID,
			Ext:      ext,
			Quality:  quality,
			Filesize: getFilesize(f),
		})
	}

	// Always append MP3 at the end using the best audio format for extraction
	// Find the best audio-only format (highest bitrate)
	var bestAudio *YtDlpFormat
	for i := range ytOut.Formats {
		f := &ytOut.Formats[i]
		isAudio := f.ACodec != "none" && f.ACodec != ""
		isVideoOnly := f.VCodec != "none" && f.VCodec != ""
		ext := getExtLabel(*f)
		if ext == "m3u8" || ext == "mpd" || strings.Contains(ext, "m3u") {
			continue
		}
		if isAudio && !isVideoOnly {
			if bestAudio == nil || f.Abr > bestAudio.Abr {
				bestAudio = f
			}
		}
	}
	if bestAudio != nil {
		entries = append(entries, FormatEntry{
			FormatID: bestAudio.FormatID,
			Ext:      "mp3",
			Quality:  "Best Audio",
			Filesize: getFilesize(*bestAudio),
		})
	} else {
		// No audio format found – add a generic MP3 placeholder
		entries = append(entries, FormatEntry{
			FormatID: "bestaudio",
			Ext:      "mp3",
			Quality:  "Best Audio",
			Filesize: 0,
		})
	}

	json.NewEncoder(w).Encode(FormatsResponse{Formats: entries})
}

// ================== THUMBNAIL PROXY ==================

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	thumbURL := r.URL.Query().Get("url")
	if thumbURL == "" {
		http.Error(w, "No URL", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(thumbURL)
	if err != nil {
		http.Error(w, "Failed to fetch thumbnail", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	io.Copy(w, resp.Body)
}

// ================== DOWNLOAD ENDPOINT ==================

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	videoURL := r.URL.Query().Get("url")
	formatID := r.URL.Query().Get("format_id")
	ext := r.URL.Query().Get("ext")
	title := r.URL.Query().Get("title")

	if videoURL == "" {
		http.Error(w, "No URL provided", http.StatusBadRequest)
		return
	}
	if ext == "" {
		ext = "mp4"
	}
	if title == "" {
		title = "video"
	}

	// Basic sanitization for filename
	safeTitle := strings.Map(func(r rune) rune {
		if strings.ContainsRune("\\/:*?\"<>|", r) {
			return '_'
		}
		return r
	}, title)

	// Set headers for file download
	filename := safeTitle + "." + ext
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Build yt-dlp command: stream to stdout
	args := []string{
		"-o", "-", // output to stdout
		"--no-playlist",
		"--quiet",
	}

	if ext == "mp3" {
		// Audio extraction: convert to mp3
		args = append(args, "-x", "--audio-format", "mp3", "--audio-quality", "0")
		if formatID != "" && formatID != "bestaudio" {
			args = append(args, "-f", formatID)
		} else {
			args = append(args, "-f", "bestaudio")
		}
	} else if formatID != "" {
		args = append(args, "-f", formatID)
	} else {
		// Default: best single mp4 file (merged audio+video)
		args = append(args, "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best")
	}

	args = append(args, videoURL)

	cmd := exec.Command("yt-dlp", args...)
	cmd.Stdout = w

	if err := cmd.Run(); err != nil {
		log.Printf("Download error: %v", err)
		// Can't write error to response at this point since headers are sent
	}
}

// ================== MAIN ==================

func main() {
	http.HandleFunc("/api/metadata", metadataHandler)
	http.HandleFunc("/api/formats", formatsHandler)
	http.HandleFunc("/api/thumbnail", thumbnailHandler)
	http.HandleFunc("/api/download", downloadHandler)

	port := "8080"
	log.Println("Server running on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
