# ASCII Art Patterns

## Compact CLI Banner

Use for command help where vertical space matters:

```text
  AUDIOBOOK
  =========
  -- organizer --
```

## Wordmark With Rule

```text
  A U D I O B O O K
  -----------------
      organizer
```

## Framed Label

```text
+----------------------+
|      AUDIOBOOK       |
|     -- organizer --  |
+----------------------+
```

## Line-Art Logo

```text
     ___        ___
    / _ \ _   _/ _ \
   | |_| | | | | | |
    \___/ \_,_|\___/
       -- organizer --
```

Use only when the shape is still readable at the target width.

## Review Checklist

- No tabs.
- No ANSI escapes.
- No Unicode unless allowed.
- Every line fits the target width.
- The art survives Markdown fenced-code rendering.
- The art is meaningful when printed in CI logs or copied into a bug report.
- If used as fallback for an ANSI/logo tier, it is a separate plain-text design,
  not just stripped ANSI.
