// Conceptual Go Code
cmd := exec.Command("python3", "ocr_brain.py", "downloaded_image.jpg")
output, _ := cmd.Output()

// Now 'output' contains the JSON {"total": 50.00, "vendor": "Walmart"}
// You parse that JSON and automatically fill the database.
