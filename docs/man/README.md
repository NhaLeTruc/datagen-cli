# Man Pages

This directory contains man pages for datagen-cli commands.

## Generated Man Pages

- `datagen.1` - Main datagen command
- `datagen-generate.1` - Generate command
- `datagen-validate.1` - Validate command
- `datagen-template.1` - Template command
- `datagen-template-list.1` - List templates subcommand
- `datagen-template-show.1` - Show template subcommand
- `datagen-template-export.1` - Export template subcommand
- `datagen-version.1` - Version command

## Viewing Man Pages

### During Development

View a man page directly:

```bash
man docs/man/datagen.1
```

Or use `man` with the full path:

```bash
man ./docs/man/datagen-generate.1
```

### After Installation

If datagen is installed system-wide, man pages should be copied to `/usr/local/share/man/man1/` or similar:

```bash
sudo cp docs/man/*.1 /usr/local/share/man/man1/
sudo mandb  # Update man database
```

Then view with:

```bash
man datagen
man datagen-generate
man datagen-validate
```

## Regenerating Man Pages

Man pages are generated from Cobra command definitions using:

```bash
go run scripts/gen-manpages.go docs/man
```

Run this whenever command descriptions, flags, or structure changes.

## Format

Man pages are in `groff` format, section 1 (User Commands).

They follow standard man page structure:
- **NAME**: Command name and brief description
- **SYNOPSIS**: Usage syntax
- **DESCRIPTION**: Detailed description
- **OPTIONS**: Command-line flags and options
- **SEE ALSO**: Related commands
- **EXAMPLES**: Usage examples (if provided in command descriptions)

## Installation Packaging

When creating distribution packages (deb, rpm, homebrew, etc.), ensure man pages are included:

**Debian/Ubuntu (deb)**:
```
/usr/share/man/man1/datagen.1.gz
```

**Homebrew**:
```ruby
man1.install Dir["docs/man/*.1"]
```

**Docker**:
Man pages are optional in containers but can be included for documentation.
