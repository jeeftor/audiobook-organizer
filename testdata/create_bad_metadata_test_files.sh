#!/bin/bash

# Create test files with various metadata issues
# These will be used to test error handling and edge cases in the organizer

# Create directories if they don't exist
mkdir -p mp3/bad_metadata m4b/bad_metadata

# Function to sanitize filenames
sanitize_filename() {
    echo "$1" | sed -e 's/[^A-Za-z0-9._-]/_/g'
}

# Create a minimal MP3 file with ID3 tags
create_bad_mp3() {
    local filename=$1
    local title=$2
    local artist=$3
    local album=$4
    local track=$5
    local total_tracks=$6

    # Create a 1-second silent MP3
    ffmpeg -f lavfi -i anullsrc=channel_layout=stereo:sample_rate=44100 -t 1 -c:a libmp3lame -b:a 128k "$filename" -y

    # Add ID3 tags with eyeD3
    if [ -n "$title" ]; then
        eyeD3 --title "$title" "$filename" >/dev/null 2>&1
    fi
    if [ -n "$artist" ]; then
        eyeD3 --artist "$artist" "$filename" >/dev/null 2>&1
    fi
    if [ -n "$album" ]; then
        eyeD3 --album "$album" "$filename" >/dev/null 2>&1
    fi
    if [ -n "$track" ]; then
        eyeD3 --track "$track" "$filename" >/dev/null 2>&1
    fi
    if [ -n "$total_tracks" ]; then
        eyeD3 --track-total "$total_tracks" "$filename" >/dev/null 2>&1
    fi
}

# Function to create a minimal M4A file with metadata
create_bad_m4a() {
    local filename=$1
    local title=$2
    local artist=$3
    local album=$4
    local track=$5
    local total_tracks=$6

    # Create a 1-second silent M4A
    ffmpeg -f lavfi -i anullsrc=channel_layout=stereo:sample_rate=44100 -t 1 -c:a aac -b:a 128k "$filename" -y

    # Add metadata with ffmpeg
    local metadata_args=()
    if [ -n "$title" ]; then
        metadata_args+=(-metadata "title=$title")
    fi
    if [ -n "$artist" ]; then
        metadata_args+=(-metadata "artist=$artist")
    fi
    if [ -n "$album" ]; then
        metadata_args+=(-metadata "album=$album")
    fi
    if [ -n "$track" ] && [ -n "$total_tracks" ]; then
        metadata_args+=(-metadata "track=$track/$total_tracks")
    fi

    # Only re-encode if we have metadata to add
    if [ ${#metadata_args[@]} -gt 0 ]; then
        local tempfile="${filename}.temp.m4a"
        ffmpeg -i "$filename" -c copy ${metadata_args[@]} "$tempfile" -y
        mv "$tempfile" "$filename"
    fi
}

# 1. Missing title
create_bad_mp3 "mp3/bad_metadata/missing_title.mp3" "" "Test Artist" "Test Album" 1 3
create_bad_m4a "m4b/bad_metadata/missing_title.m4a" "" "Test Artist" "Test Album" 1 3

# 2. Missing artist
create_bad_mp3 "mp3/bad_metadata/missing_artist.mp3" "Test Title" "" "Test Album" 2 3
create_bad_m4a "m4b/bad_metadata/missing_artist.m4a" "Test Title" "" "Test Album" 2 3

# 3. Missing album
create_bad_mp3 "mp3/bad_metadata/missing_album.mp3" "Test Title" "Test Artist" "" 3 3
create_bad_m4a "m4b/bad_metadata/missing_album.m4a" "Test Title" "Test Artist" "" 3 3

# 4. Missing track number
create_bad_mp3 "mp3/bad_metadata/missing_track.mp3" "Test Title" "Test Artist" "Test Album" "" 3
create_bad_m4a "m4b/bad_metadata/missing_track.m4a" "Test Title" "Test Artist" "Test Album" "" 3

# 5. Invalid track number (text)
create_bad_mp3 "mp3/bad_metadata/invalid_track.mp3" "Test Title" "Test Artist" "Test Album" "one" 3
create_bad_m4a "m4b/bad_metadata/invalid_track.m4a" "Test Title" "Test Artist" "Test Album" "one" 3

# 6. Very long fields
long_string=$(printf 'A%.0s' {1..1000})
create_bad_mp3 "mp3/bad_metadata/long_fields.mp3" "$long_string" "$long_string" "$long_string" 1 3
create_bad_m4a "m4b/bad_metadata/long_fields.m4a" "$long_string" "$long_string" "$long_string" 1 3

# 7. Special characters in all fields
special_chars="!@#$%^&*()_+{}|:<>?[];',.\""
create_bad_mp3 "mp3/bad_metadata/special_chars.mp3" "Title $special_chars" "Artist $special_chars" "Album $special_chars" 1 3
create_bad_m4a "m4b/bad_metadata/special_chars.m4a" "Title $special_chars" "Artist $special_chars" "Album $special_chars" 1 3

# 8. Empty file (corrupted)
touch "mp3/bad_metadata/empty.mp3"
touch "m4b/bad_metadata/empty.m4a"

echo "Created test files with bad metadata in mp3/bad_metadata/ and m4b/bad_metadata/"
echo "You may need to install eyeD3 for MP3 tag editing: pip install eyeD3"
