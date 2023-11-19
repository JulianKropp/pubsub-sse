import requests
import re
import os

# Regular expression to find URLs
url_pattern = r'url\((https://fonts.gstatic.com/s/[^)]+)\)'
# Original font list file
fontsOrginal = 'fonts-orginal.css'
# New font list file
fonts = 'fonts.css'

# Function to download a font from a URL and return the local path
def download_font(url, folder='fonts'):
    if not os.path.exists(folder):
        os.makedirs(folder)

    file_name = url.split('/')[-1]
    path = os.path.join(folder, file_name)

    response = requests.get(url, stream=True)

    if response.status_code == 200:
        with open(path, 'wb') as f:
            f.write(response.content)
        print(f"Downloaded: {file_name}")
        return os.path.join(folder, file_name)
    else:
        print(f"Failed to download {file_name}")
        return None

# Read CSS content from file
with open(fontsOrginal, 'r') as file:
    css_text = file.read()

# Find all URLs and replace them with local paths
urls = re.findall(url_pattern, css_text)
for url in urls:
    local_path = download_font(url)
    # Replace \ with /
    local_path = local_path.replace('\\', '/')
    if local_path:
        css_text = css_text.replace(url, local_path)

# Save the modified CSS to a new file
with open(fonts, 'w') as file:
    file.write(css_text)
