import (
    // ... existing imports
    "encoding/json"
    "io"
    "net/http"
    "os/exec"
)

// 1. DEFINE THE STRUCTURE OF THE PYTHON OUTPUT
type OCRResult struct {
    Vendor     string  `json:"vendor"`
    Total      float64 `json:"total"`
    Confidence string  `json:"confidence"`
    Error      string  `json:"error,omitempty"`
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == s.State.User.ID {
        return
    }

    // ... existing !spend and !balance commands ...

    // 3. SCAN COMMAND (The Bridge to Python)
    // Trigger: User uploads an image and comments "!scan"
    if strings.HasPrefix(m.Content, "!scan") && len(m.Attachments) > 0 {
        
        s.ChannelMessageSend(m.ChannelID, "👀 Analyzing receipt...")

        // A. Download the image
        attachment := m.Attachments[0]
        localFileName := "temp_receipt_" + attachment.ID + ".jpg"
        
        err := downloadFile(attachment.URL, localFileName)
        if err != nil {
            s.ChannelMessageSend(m.ChannelID, "❌ Failed to download image.")
            return
        }
        
        // Clean up file after we are done
        defer os.Remove(localFileName) 

        // B. Call Python (The "Rhino" Script)
        // Ensure "ocr_brain.py" is in the same folder or provide full path
        cmd := exec.Command("python3", "ocr_brain.py", localFileName)
        output, err := cmd.CombinedOutput()
        if err != nil {
            log.Printf("Python execution error: %v", err)
            s.ChannelMessageSend(m.ChannelID, "❌ Python script failed.")
            return
        }

        // C. Parse the JSON result from Python
        var result OCRResult
        err = json.Unmarshal(output, &result)
        if err != nil {
            // If JSON fails, print raw output for debugging
            log.Printf("Raw Python Output: %s", string(output))
            s.ChannelMessageSend(m.ChannelID, "❌ Could not understand Python's response.")
            return
        }

        if result.Error != "" {
            s.ChannelMessageSend(m.ChannelID, "❌ OCR Error: "+result.Error)
            return
        }

        // D. Reply with the specific data found
        response := fmt.Sprintf("🧾 **Receipt Scanned!**\n"+
            "Vendor: **%s**\n"+
            "Detected Amount: **$%.2f**\n"+
            "Confidence: *%s*\n\n"+
            "Type `!spend %.2f %s #Project` to save this.", 
            result.Vendor, result.Total, result.Confidence, result.Total, "Category")

        s.ChannelMessageSend(m.ChannelID, response)
    }
}

// Helper: Download file from URL
func downloadFile(url string, filepath string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}
