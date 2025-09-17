> [! Warning]
> This project is still not at an usable stage.

# Teka Finance

Teka is a personal finance tracking app built on top of [hledger](https://hledger.org).  
It reads your hledger reports to generate clean reports and visualization charts, accessible through a GUI/web interface.

Teka also provides some quality-of-life commands to help manage your hledger journals more easily.

Teka reads data directly from hledger’s generated reports, so you can use your existing journals without modification (except for multi-currency transactions, which must follow a specific format, documentation on that coming soon).

---

## 📑 Table of Contents

- [🚧 Project Status](#-project-status)
- [⚠️ Disclaimer](#-disclaimer)
- [🚀 Getting Started](#-getting-started)
    - [1. Build the frontend](#1-build-the-frontend)
    - [2. Build the binary](#2-build-the-binary)
        - [Linux / macOS](#linux--macos)
        - [Windows](#windows)
- [📖 Usage](#-usage)
    - [Help](#help)
    - [Journal File](#journal-file)
    - [Serve](#serve)
    - [Add](#add-command)
- [⚙️ Configuration](#️-configuration)

---

## 🚧 Project Status

This project is still **under active development**.  
Many features are not yet implemented, and bugs may be present. It is not yet at a usable stage. Expect breaking changes until the first stable release.

## ⚠️ Disclaimer

Teka was originally built to visualize **my own finances** recorded with hledger, so it is **heavily opinionated**. Your mileage may vary.

- If your journals contain only single-currency transactions, everything will **probably** work out of the box.
- Multi-currency transactions require a specific format (to be documented soon).

Teka is safe to try because it only reads your hledger reports and does not modify your journals, unless you use the `add` command.

## 🚀 Getting Started

### 0. Install hledger

Teka requires hledger to be installed on your system. Follow hledger's [installation instructions](https://hledger.org/install.html) before proceeding if you don't have it installed already.

Pre-built binaries will be provided once the project is released.  
For now, you’ll need to build manually.

### 1. Build the frontend

```bash
cd frontend
pnpm build
```

### 2. Build the binary

#### Linux / macOS

```bash
go build -o teka
chmod +x teka
```

#### Windows

```powershell
go build -o teka.exe
```

You can then move the executable anywhere and run it.

## 📖 Usage

### Help

Show all available commands and flags:

```bash
./teka -h
```

### Journal File

By default, Teka uses the journal file defined in the `LEDGER_FILE` environment variable (same as hledger).

To specify a custom journal file, use the `--file` (`-f`) flag:

```bash
./teka serve --file /path/to/file.journal
```

### Serve

Start the web server and open the web interface:

```bash
./teka serve
```

By default, the server runs on port `8080`, so you can access the UI at:

```
http://127.0.0.1:8080
```

### Add Command

The `add` command provides an interactive way to append transactions to your journal.

#### Initiating

```bash
teka add
```

This adds the transaction to your default ledger (`$LEDGER_FILE`).

You can also specify a file:

```bash
teka add -f /path/to/file.journal
```

#### Interactive Workflow

You will be prompted step by step to enter the transaction:

```
Date? 2025-01-01
Note? Received salary
Account? assets:bank
Amount? 3000 USD
Account? income:salary
Amount? -3000 USD
Account? 
```

- **Finish transaction**: press Enter on an empty account field.
- **Amount**: Anything you type in amount prompt is directly added to the transaction so you can add cost notation (`100 EUR @ 1.2 EUR`) or comments to the amount field as usual. You can also keep amount field empty.
- **Validation**: after saving, Teka runs `hledger check` on the journal and lets you keep or revert changes if errors occur.

#### Adding Comments

You can insert comments during date/account prompts:

- Type only `;` or `#` in the account/date prompt to enter a comment.
- If you use `;` in the account prompt it adds an indented comment.
- Using `#` always adds an unindented comment.

Examples:

```
Date? ;
Comment? ; This is a comment 
Note? Salary ; inline comment
Account? ;
Comment? ; indented comment 
Account? #
Comment? ; unindented comment
Account? assets:cash
Amount? 100 USD ; inline comment
```

Teka does not add `;` or `#` automatically you must type them in the comment prompt again. This makes it possible to include any text inline, not just comments.

#### The Mighty Dot

Shortcuts using `.` are available in multiple fields:

- **Date shortcuts**
    - `.` for today
    - `.y` for yesterday
    - `.t` for tomorrow
- **Search**
    - At account or note prompt: `.search term` searches accounts and lets you select from the search results
- **Balance auto-fill**
    - At **amount prompt**: `.` fills remaining amount to balance the transaction, then closes it
    - Works only for single-currency postings

#### Currency Conversion with FX Gain

If you follow the multi-currency workflow (documentation coming soon), Teka can automatically generate foreign exchange gain and equity postings.

Mark foreign currency account with a `$`:

Example: converting EUR to USD

```
Account? $assets:bank:EU bank
Amount? -100 EUR
Account? assets:bank:US bank
Amount? 120 USD
```

Produces:

```
assets:bank:EU bank    -100 EUR @@ 110 USD
assets:bank:US bank     120 USD
income:fx gain         -10 USD
equity:conversion      100 EUR
equity:conversion     -110 USD
```

Teka will calculate your gain/loss based on average cost. If you have multiple files use the `--mainfile` flag. Teka will do its calculation from the main file and then add the transaction to the file passed through `--file`.

## ⚙️ Configuration

When you first run Teka, it will create a configuration file in the OS config path and print its location in the terminal.

Default config paths:

- **Linux:** `~/.config/teka/tekaconf.yaml`
- **Windows:** `C:\Users\<YourUsername>\AppData\Local\teka\tekaconf.yaml`

