import sys
import json
import re
try:
    from PIL import Image
    import pytesseract
except ImportError:
    print(json.dumps({"error": "Missing libraries. pip install pytesseract pillow"}))
    sys.exit(1)

# 1. SETUP
# If you are on Windows, you might need to point to the exe:
# pytesseract.pytesseract.tesseract_cmd = r'C:\Program Files\Tesseract-OCR\tesseract.exe'

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
                  
