#!/bin/bash

# Create test files with various missing/invalid metadata using only FFmpeg

# Create directories
mkdir -p mp3/bad_metadata m4b/bad_metadata

# Function to create test files with specific metadata
create_test_file() {
    local format=$1
    local output=$2
    local title=$3
    local artist=$4
    local album=$5
    local track=$6
    local total_tracks=$7

    local metadata=()
    [ -n "$title" ] && metadata+=(-metadata "title=$title")
    [ -n "$artist" ] && metadata+=(-metadata "artist=$artist")
    [ -n "$album" ] && metadata+=(-metadata "album=$album")
    [ -n "$track" ] && [ -n "$total_tracks" ] && metadata+=(-metadata "track=$track/$total_tracks")

    # Create a 1-second silent audio file with specified format
    ffmpeg -f lavfi -i anullsrc=channel_layout=stereo:sample_rate=44100 -t 1 \
        -c:a ${format} -b:a 64k "${metadata[@]}" "$output" -y 2>/dev/null
}

# 1. Completely empty metadata
create_test_file "libmp3lame" "mp3/bad_metadata/empty_metadata.mp3" "" "" "" "" ""
create_test_file "aac" "m4b/bad_metadata/empty_metadata.m4a" "" "" "" "" ""

# 2. Missing title
create_test_file "libmp3lame" "mp3/bad_metadata/no_title.mp3" "" "Test Artist" "Test Album" "1" "3"
create_test_file "aac" "m4b/bad_metadata/no_title.m4a" "" "Test Artist" "Test Album" "1" "3"

# 3. Missing artist
create_test_file "libmp3lame" "mp3/bad_metadata/no_artist.mp3" "Test Title" "" "Test Album" "2" "3"
create_test_file "aac" "m4b/bad_metadata/no_artist.m4a" "Test Title" "" "Test Album" "2" "3"

# 4. Missing album
create_test_file "libmp3lame" "mp3/bad_metadata/no_album.mp3" "Test Title" "Test Artist" "" "3" "3"
create_test_file "aac" "m4b/bad_metadata/no_album.m4a" "Test Title" "Test Artist" "" "3" "3"

# 5. Missing track numbers
create_test_file "libmp3lame" "mp3/bad_metadata/no_track.mp3" "Test Title" "Test Artist" "Test Album" "" ""
create_test_file "aac" "m4b/bad_metadata/no_track.m4a" "Test Title" "Test Artist" "Test Album" "" ""

# 6. Invalid track format
create_test_file "libmp3lame" "mp3/bad_metadata/invalid_track.mp3" "Test Title" "Test Artist" "Test Album" "one" "three"
create_test_file "aac" "m4b/bad_metadata/invalid_track.m4a" "Test Title" "Test Artist" "Test Album" "one" "three"

# 7. Only title
create_test_file "libmp3lame" "mp3/bad_metadata/only_title.mp3" "Only Title" "" "" "" ""
create_test_file "aac" "m4b/bad_metadata/only_title.m4a" "Only Title" "" "" "" ""

# 8. Only artist
create_test_file "libmp3lame" "mp3/bad_metadata/only_artist.mp3" "" "Only Artist" "" "" ""
create_test_file "aac" "m4b/bad_metadata/only_artist.m4a" "" "Only Artist" "" "" ""

# 9. Only album
create_test_file "libmp3lame" "mp3/bad_metadata/only_album.mp3" "" "" "Only Album" "" ""
create_test_file "aac" "m4b/bad_metadata/only_album.m4a" "" "" "Only Album" "" ""

echo "Created test files with bad metadata in mp3/bad_metadata/ and m4b/bad_metadata/"
