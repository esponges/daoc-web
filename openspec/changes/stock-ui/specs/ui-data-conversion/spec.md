## ADDED Requirements

### Requirement: UI layouts convert to JSON
The converter SHALL read `ui/uimain.xml` and its includes for a chosen skin and write one JSON file per window with its templates (windows, buttons, image areas, text), sizes, positions and texture references, resolving includes and named templates.

#### Scenario: Chat window
- **WHEN** the converter runs for the `atlantis` skin
- **THEN** the output holds the chat window with its size, minimum size and the textures it names

#### Scenario: Unknown element
- **WHEN** a layout has an element the converter does not know
- **THEN** it keeps going, and the report names the element and file

### Requirement: Art and fonts convert to web formats
The converter SHALL read the DDS, TGA and BMP textures and the bitmap fonts that converted layouts reference, and write PNG atlases and font glyph tables.

#### Scenario: Atlas
- **WHEN** a layout names `atlantis/atlantis_01.tga`
- **THEN** a PNG of the same size exists in the output and the layout points to it

#### Scenario: Missing texture
- **WHEN** a referenced texture is not in the install
- **THEN** it is listed in the report and the layout is still written

### Requirement: Icon sheets are converted and found by number
The converter SHALL write the install's icon sheets (10 by 10 grids of 32-pixel icons, one file per hundred numbers) as PNG and list them, so that icon n is found on the sheet whose first number is n - n mod 100, at column n mod 10 and row (n mod 100) div 10.

#### Scenario: A combat icon
- **WHEN** icon 634 is looked up
- **THEN** it is on the sheet starting at 600, column 4, row 3

#### Scenario: A number with no sheet
- **WHEN** an icon number falls in a range the install has no sheet for
- **THEN** no picture is returned and the caller draws a fallback

### Requirement: Classic window art is converted
The converter SHALL write the classic skin's window art (`data/skin1`: the paper doll, vault, merchant and so on) as PNG with its size, for windows the install has no layout file for.

#### Scenario: The paper doll
- **WHEN** the converter has run
- **THEN** `skin1/pperdoll.png` exists at 192 by 448 and its size is in the layout data

### Requirement: Bitmap fonts carry their metrics
The converter SHALL read each bitmap font sheet's glyph widths from the marker row under each row of glyphs, characters running in Windows-1252 order from "!", and SHALL leave out and report a sheet that does not follow that layout.

#### Scenario: Glyph widths
- **WHEN** the font `arial11` is converted
- **THEN** it has a glyph for every character from "!" to "ÿ" and "l" is narrower than "W"

### Requirement: Output stays out of the repository
All converted UI data SHALL be written under the gitignored `web/data/ui/`.

#### Scenario: Clean tree
- **WHEN** the converter has run
- **THEN** `git status` shows no converted UI data
