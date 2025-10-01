# Test Data

This directory contains test data for the audiobook organizer.

## Directory Structure

- `epub/`: Sample EPUB files for testing
- `m4b/`: Sample M4B files for testing
  - `bad_metadata/`: Files with various metadata issues
- `mp3/`: Sample MP3 files for testing
- `mp3-badmetadata/`: MP3 files with various metadata issues
- `mp3flat/`: Flat directory structure MP3 files
- `mp3track/`: MP3 files with track numbers in filenames

## Test Files

### Metadata Test Files

Files in `m4b/bad_metadata/` and `mp3-badmetadata/` test various edge cases:
- `empty.*`: Empty files
- `empty_metadata.*`: Files with empty metadata
- `invalid_track.*`: Files with invalid track numbers
- `long_fields.*`: Files with very long metadata fields
- `missing_*`: Files with specific missing metadata fields
- `no_*`: Files with empty metadata fields
- `only_*`: Files with only one metadata field set
- `special_chars.*`: Files with special characters in metadata

### Integration Tests

Integration tests use these files to verify:
1. Metadata extraction
2. File organization
3. Error handling
4. Edge cases with special characters and formatting

## Adding New Test Files

When adding new test files:
1. Place them in the appropriate directory based on file type
2. Use descriptive names that indicate the test case
3. Document any special test cases in this README
4. Keep file sizes small (use dummy content)
