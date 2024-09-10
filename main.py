
import json
import yt_dlp

ydl_opts = {
    'format': 'bestaudio/best',
    'noplaylist': 'True',
    'postprocessors': [
        {
            'key': 'FFmpegExtractAudio',
            'preferredcodec': 'm4a',
            'preferredquality': '192'
        }
    ]
}

URL = 'https://www.youtube.com/watch?v=oNx35-nsmZM'

# ℹ️ See help(yt_dlp.YoutubeDL) for a list of available options and public functions
ydl_opts = {}
with yt_dlp.YoutubeDL(ydl_opts) as ydl:
    info = ydl.extract_info(URL, download=False)

    # ℹ️ ydl.sanitize_info makes the info json-serializable
    info = ydl.sanitize_info(info)

    print(info['formats'][2]['url'])
