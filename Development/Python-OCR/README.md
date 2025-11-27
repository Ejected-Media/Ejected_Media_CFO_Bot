This is a great architectural question. When you control both the mobile app and the server, you have two choices for where the "brain" lives.
Here is the thinking behind putting the OCR in Python (Server-side) rather than Kotlin (On-Device) for this specific project:
The "Central Brain" Strategy (Why Python?)
 * The Discord Factor: You are building a system where data enters from two places: the Android App and the Discord Bot (desktop/web).
   * If you put the OCR logic in Kotlin, your Android app is smart, but your Discord Bot remains "dumb." If you drag a PDF invoice into Discord on your laptop, the Android app can't help you.
   * By putting OCR in Python on the server, both the Discord Bot and the Android App can send an image to the "Brain" and get the same result.
 * The "Rhino" Factor (Iterability): Receipt parsing is messy. Walmart changes their receipt font; Amazon changes their invoice layout.
   * Kotlin (On-Device): To fix a parsing bug, you have to recompile the app, build an APK, and reinstall/update it on your phone.
   * Python (Server): You just tweak the regex in your script, save it, and boom—it's fixed for everyone instantly. This matches your desire for a "scripting layer" like Rhino3D.
The Fun Part: The Python OCR Script
Let's build the "Rhino Script" that the Go bot will eventually call.
The Goal: This script takes an image file, reads the text, and tries to find the Total Amount and the Vendor.
Prerequisites:
 * Python 3.x
 * tesseract-ocr (The actual engine installed on your OS)
 * Python libraries: pip install pytesseract pillow
ocr_brain.py

```
import sys
import json
import re
try:
    from PIL import Image
    import pytesseract
except ImportError:
    print(json.dumps({"error": "Missing libraries. pip install pytesseract pillow"}))
    sys.exit(1)

```

#### 1. SETUP
If you are on Windows, you might need to point to the exe:
_pytesseract.pytesseract.tesseract_cmd = r'C:\Program Files\Tesseract-OCR\tesseract.exe'_

```
def extract_financials(image_path):
    """
    The 'Rhino' logic: Analyzes raw text to find money and names.
    """
    try:
        # Load the image
        img = Image.open(image_path)
        
        # EXTRACT RAW TEXT
        # --psm 6 assumes a single uniform block of text
        raw_text = pytesseract.image_to_string(img, config='--psm 6')
        
        lines = raw_text.split('\n')
        
        # DATA CONTAINERS
        data = {
            "vendor": "Unknown",
            "date": None,
            "total": 0.00,
            "confidence": "low",
            "raw_preview": raw_text[:100].replace('\n', ' ') + "..."
        }

        # 2. LOGIC: FIND THE VENDOR (Heuristic: First non-empty line)
        for line in lines:
            clean_line = line.strip()
            if len(clean_line) > 2:
                data["vendor"] = clean_line
                break

        # 3. LOGIC: FIND THE TOTAL
        # Regex to look for money patterns: $10.00, 10.99, etc.
        money_pattern = r'[\$]?(\d+[\.,]\d{2})'
        
        amounts = []
        for line in lines:
            # specifically look for lines with "Total" or "Amount" to be safer
            matches = re.findall(money_pattern, line)
            for match in matches:
                # Clean up the string (replace comma with dot if needed)
                clean_match = match.replace(',', '.')
                try:
                    amounts.append(float(clean_match))
                except ValueError:
                    continue
        
        # Heuristic: The largest number on the receipt is usually the Total
        if amounts:
            data["total"] = max(amounts)
            data["confidence"] = "medium"

        return data

    except Exception as e:
        return {"error": str(e)}

if __name__ == "__main__":
    # How to run: python ocr_brain.py receipt.jpg
    if len(sys.argv) < 2:
        print(json.dumps({"error": "No image path provided"}))
        sys.exit(1)

    image_path = sys.argv[1]
    result = extract_financials(image_path)
    
    # Output JSON for the Go Bot to read
    print(json.dumps(result, indent=2))

```

How to test this manually
 * Take a picture of a receipt (or find one on Google Images).
 * Save it as test_receipt.jpg.
 * Run the script:
   python ocr_brain.py test_receipt.jpg

How this connects to your Go Bot (The "Rhino" Link)
In the future, your Go backend will receive an image attachment from Discord. It will save that image locally, and then run a command like this:
// Conceptual Go Code
cmd := exec.Command("python3", "ocr_brain.py", "downloaded_image.jpg")
output, _ := cmd.Output()

// Now 'output' contains the JSON {"total": 50.00, "vendor": "Walmart"}
// You parse that JSON and automatically fill the database.

This keeps your Go code clean and stable, while your Python script can get as messy and complex as it needs to be to handle different receipt types!
