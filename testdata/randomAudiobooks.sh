#!/bin/bash

# Check if test.mp3 exists, if not, generate a 5-second test tone
if [ ! -f test.mp3 ]; then
    echo "Error: test.mp3 does not exist."
    echo "Would you like to generate a 5-second test tone instead? (y/n)"
    read -r response
    if [ "$response" = "y" ]; then
        ffmpeg -f lavfi -i "sine=frequency=1000:duration=5" -c:a mp3 -b:a 32k test_tone.mp3
    else
        exit 1
    fi
fi

# Create directories for MP3 and M4B files
mkdir -p mp3 m4b

# Function to sanitize filenames
sanitize_filename() {
    echo "$1" | sed -e 's/[^A-Za-z0-9._-]/_/g'
}

# Audiobook data
audiobooks=(
    "Mystery Series|Mystery of the Lost City|Jane Doe|3|mp3"
    "Mystery Series|Mystery of the Lost City|Jane Doe|1|m4b"
    "Epic Saga™|Adventure: Quest & Glory!|John*Smith|3|mp3"
    "Epic Saga™|Adventure: Quest & Glory!|John*Smith|1|m4b"
    "Tales of Ångström 📚|Café Chronicles|María López|3|mp3"
    "Tales of Ångström 📚|Café Chronicles|María López|1|m4b"
    "Saga of Endless Horizons|The Epic Tale That Spans Generations|Alexander von Longname|3|mp3"
    "Saga of Endless Horizons|The Epic Tale That Spans Generations|Alexander von Longname|1|m4b"
    "Book: With Colons?|Book: With Colons?|Author*With|Invalid|3|mp3"
    "Book: With Colons?|Book: With Colons?|Author*With|Invalid|1|m4b"
    "Audiobook & Symbols % $ # @ !|Audiobook & Symbols % $ # @ !|Author+Plus-Minus±§|3|mp3"
    "Audiobook & Symbols % $ # @ !|Audiobook & Symbols % $ # @ !|Author+Plus-Minus±§|1|m4b"
    "Series With Tabs| Audiobook With Spaces |Author With Spaces |3|mp3"
    "Series With Tabs| Audiobook With Spaces |Author With Spaces |1|m4b"
    "Series With Emoji 📚📖|Audiobook With Emoji 📚🔍|Author With Emoji 👨‍💻|3|mp3"
    "Series With Emoji 📚📖|Audiobook With Emoji 📚🔍|Author With Emoji 👨‍💻|1|m4b"
)

# Counter for unique file naming
counter=1

for audiobook in "${audiobooks[@]}"; do
    IFS='|' read -r album title artist track_count extension <<< "$audiobook"

    # Sanitize components for filename
    safe_album=$(sanitize_filename "$album")
    safe_title=$(sanitize_filename "$title")
    safe_artist=$(sanitize_filename "$artist")

    # Determine format and directory
    if [ "$extension" = "m4b" ]; then
        format="mp4"
        dir="m4b"
        output_ext="m4b"
    else
        format="mp3"
        dir="mp3"
        output_ext="mp3"
    fi

    # Create tracks
    for ((track=1; track<=track_count; track++)); do
        output_file="./$dir/strange_audiobook_${counter}_${safe_album}_${safe_title}_${safe_artist}_Tr${track}.${output_ext}"

        echo "Creating $output_file with metadata: -metadata title='$title: Track $track' -metadata artist='$artist' -metadata album='$album' -metadata genre='Audiobook' -metadata track='$track'"

        ffmpeg -i test_tone.mp3 -c:a copy -f "$format" \
            -metadata title="$title: Track $track" \
            -metadata artist="$artist" \
            -metadata album="$album" \
            -metadata genre="Audiobook" \
            -metadata track="$track" \
            "$output_file"

        ((counter++))
    done

    # For M4B, create a single file with the same metadata (no track number)
    if [ "$extension" = "m4b" ]; then
        output_file="./$dir/strange_audiobook_${counter}_${safe_album}_${safe_title}_${safe_artist}.${output_ext}"

        echo "Creating $output_file with metadata: -metadata title='$title' -metadata artist='$artist' -metadata album='$album' -metadata genre='Audiobook'"

        ffmpeg -i test_tone.mp3 -c:a copy -f "$format" \
            -metadata title="$title" \
            -metadata artist="$artist" \
            -metadata album="$album" \
            -metadata genre="Audiobook" \
            "$output_file"

        ((counter++))
    fi
done

echo "Done! Created $((counter-1)) files (MP3s in ./mp3, M4Bs in ./m4b)."
