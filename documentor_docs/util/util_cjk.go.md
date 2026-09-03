# Documentation: `util/util_cjk.go`

## Overview

The `util/util_cjk.go` file provides utility functions and custom Unicode range tables for classifying and handling Chinese, Japanese, Korean (CJK), Yi, and other East Asian scripts/characters.

Its main capabilities include:
1. Identifying whether a rune is considered an **East Asian Wide** character.
2. Identifying whether a rune belongs to the **Space-Discarding Unicode Character Set** (following W3C CSS Text specifications).
3. Classifying a rune's **East Asian Width** property (Fullwidth, Halfwidth, Wide, Narrow, Ambiguous, or Neutral) in accordance with Unicode Standard Annex #11 (UAX #11).

---

## Private Constants & Internal Range Tables

The file defines 29 package-private `unicode.RangeTable` variables. These define specific Unicode blocks using either 16-bit (`R16`) or 32-bit (`R32`) ranges. They are primarily used by `IsSpaceDiscardingUnicodeRune`.

### 16-Bit Range Tables (`R16`)
* `cjkRadicalsSupplement` (`0x2E80`–`0x2EFF`)
* `kangxiRadicals` (`0x2F00`–`0x2FDF`)
* `ideographicDescriptionCharacters` (`0x2FF0`–`0x2FFF`)
* `cjkSymbolsAndPunctuation` (`0x3000`–`0x303F`)
* `hiragana` (`0x3040`–`0x309F`)
* `katakana` (`0x30A0`–`0x30FF`)
* `kanbun` (`0x3130`–`0x318F`, `0x3190`–`0x319F`)
* `cjkStrokes` (`0x31C0`–`0x31EF`)
* `katakanaPhoneticExtensions` (`0x31F0`–`0x31FF`)
* `cjkCompatibility` (`0x3300`–`0x33FF`)
* `cjkUnifiedIdeographsExtensionA` (`0x3400`–`0x4DBF`)
* `cjkUnifiedIdeographs` (`0x4E00`–`0x9FFF`)
* `yiSyllables` (`0xA000`–`0xA48F`)
* `yiRadicals` (`0xA490`–`0xA4CF`)
* `cjkCompatibilityIdeographs` (`0xF900`–`0xFAFF`)
* `verticalForms` (`0xFE10`–`0xFE1F`)
* `cjkCompatibilityForms` (`0xFE30`–`0xFE4F`)
* `smallFormVariants` (`0xFE50`–`0xFE6F`)
* `halfwidthAndFullwidthForms` (`0xFF00`–`0xFFEF`)

### 32-Bit Range Tables (`R32`)
* `kanaSupplement` (`0x1B000`–`0x1B0FF`)
* `kanaExtendedA` (`0x1B100`–`0x1B12F`)
* `smallKanaExtension` (`0x1B130`–`0x1B16F`)
* `cjkUnifiedIdeographsExtensionB` (`0x20000`–`0x2A6DF`)
* `cjkUnifiedIdeographsExtensionC` (`0x2A700`–`0x2B73F`)
* `cjkUnifiedIdeographsExtensionD` (`0x2B740`–`0x2B81F`)
* `cjkUnifiedIdeographsExtensionE` (`0x2B820`–`0x2CEAF`)
* `cjkUnifiedIdeographsExtensionF` (`0x2CEB0`–`0x2EBEF`)
* `cjkCompatibilityIdeographsSupplement` (`0x2F800`–`0x2FA1F`)
* `cjkUnifiedIdeographsExtensionG` (`0x30000`–`0x3134F`)

---

## Functions

### `IsEastAsianWideRune`

```go
func IsEastAsianWideRune(r rune) bool
```

#### Description
Determines whether a given rune `r` is classified as an East Asian wide character.

#### Logic
Returns `true` if `r` belongs to any of the following standard Go `unicode` tables or internal tables:
* `unicode.Hiragana`
* `unicode.Katakana`
* `unicode.Han`
* `unicode.Lm` (Letter, modifier)
* `unicode.Hangul`
* `cjkSymbolsAndPunctuation`

Otherwise, returns `false`.

---

### `IsSpaceDiscardingUnicodeRune`

```go
func IsSpaceDiscardingUnicodeRune(r rune) bool
```

#### Description
Determines whether a given rune `r` belongs to the space-discarding Unicode character set, as specified in the [W3C CSS Text Module Level 3 Editor's Draft](https://www.w3.org/TR/2020/WD-css-text-3-20200429/#space-discard-set).

#### Logic
Evaluates whether the rune `r` falls inside any of the 29 internal `unicode.RangeTable` variables defined in this file (e.g., CJK Radicals, Ideographs, Hiragana, Katakana, Yi Syllables, vertical/compatibility forms, and Extensions A through G).

Returns `true` if matched, or `false` otherwise.

---

### `EastAsianWidth`

```go
func EastAsianWidth(r rune) string
```

#### Description
Computes and returns the East Asian Width property string for a given rune `r`, as defined by [Unicode Standard Annex #11 (TR11-36)](https://www.unicode.org/reports/tr11/tr11-36.html).

#### Return Values
The function evaluates code point ranges using a `switch` statement and returns one of six possible string codes:

| String Return | East Asian Width Category | Description |
| :--- | :--- | :--- |
| `"F"` | Fullwidth | Characters that are explicitly fullwidth variants of narrow/halfwidth characters (e.g., fullwidth punctuation, `0x3000` ideographic space). |
| `"H"` | Halfwidth | Characters that are explicit halfwidth variants of fullwidth characters (e.g., halfwidth Katakana). |
| `"W"` | Wide | Characters that are wide (occupying 2 display cells) and do not have narrow variants (e.g., CJK Unified Ideographs, Kana, Yi). |
| `"Na"` | Narrow | Characters that are always narrow (e.g., standard ASCII printable characters `0x0020`–`0x007E`, specific punctuation). |
| `"A"` | Ambiguous | Characters whose width depends on context or system display settings (e.g., Greek, Cyrillic, certain symbols, enclosed alphanumerics). |
| `"N"` | Neutral | All other characters that do not fall into any East Asian context (default case). |