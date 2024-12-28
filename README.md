# GitHub Skyline (gh-skyline)

A GitHub CLI extension that generates 3D-printable STL files of your GitHub contribution graph.

## Features

- Generate a Binary STL file from GitHub contribution data for 3D printing
- Customizable year selection (single year and multi-year)
- Automatic authentication via GitHub CLI or specify a user
- ASCII art loading preview of contribution data unique to each user and year

| 3D Print                                                                                                   | ASCII Art                                                                                                                               |
| ---------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| ![Example GitHub Skyline](https://github.com/user-attachments/assets/ed0fe34e-6825-4eb2-91d7-a0834966dc3a) | ![Example of gh-skyline and ASCII art in the terminal](https://github.com/user-attachments/assets/8ddda088-ac5a-4020-8ae0-ef0d4825f6a1) |

## Usage

### Prerequisites

The extension requires the [`gh` CLI](https://cli.github.com/) to be installed and in the `PATH`. The extension also requires the user to have authenticated via `gh auth`.

### Installing

This project is a GitHub CLI extension. After installing the `gh` CLI, from a command-line run:

```bash
gh extension install github/gh-skyline
```

### Extension Flags

You can run the `gh skyline` command with the following flags:

- `-d`, `--debug`: Enable debug logging for more detailed output.
  - Example: `gh skyline --debug`
- `-h`, `--help`: Show help for the command.
  - Example: `gh skyline --help`
- `-f`, `--full`: Generate the contribution graph from the user's join year to the current year.
  - Example: `gh skyline --full`
- `-o`, `--output`: Specify the output filename. If not provided, the default is `{username}-{year}-github-skyline.stl`.
  - Example: `gh skyline --output my-skyline.stl`
- `-u`, `--user`: Specify the GitHub username. If not provided, the authenticated user is used.
  - Example: `gh skyline --user mona`
- `-y`, `--year`: Specify the year or range of years for the skyline. Must be between 2008 and the current year.
  - Examples: `gh skyline --year 2020`, `gh skyline --year 2014-2024`
- `-w`, `--web`: Open the GitHub profile for the authenticated or specified user.
  - Example: `gh skyline --web`, `gh skyline --user mona --web`

### Examples

Generate a skyline STL file that defaults to the current year for the authenticated user:

```bash
gh skyline
```

Generate a skyline for a specific year for the authenticated user:

```bash
gh skyline --year 2023
```

Generate a skyline for a specific user and year:

```bash
gh skyline --user mona --year 2023
```

Generate a skyline for a range of years for the authenticated user:

```bash
gh skyline --year 2014-2024
```

Generate a skyline from the user's join year to the current year:

```bash
gh skyline --full
```

Enable debug logging:

```bash
gh skyline --debug
```

By default, the CLI will create a `{username}-{year}-github-skyline.stl` file in your current directory. You can specify a different filename using the `--output` flag.

```bash
gh skyline --output my-skyline.stl
```

Open the GitHub profile for the authenticated user:

```bash
gh skyline --web
```

Open the GitHub profile for a specific user:

```bash
gh skyline --user mona --web
```

## ASCII Art

The extension generates ASCII art in terminal while loading, a unique and fun way to vizualise your contribution data while you wait! Each column represents one week. Days within each week are reordered vertically to create a "building" effect, with empty spaces (no contributions) at the top.

- `' '` Empty/Sky: No contributions
- `'.'` Future dates: What contributions could you make?
- `'░'` Low level: Light contribution activity
- `'▒'` Medium level: Moderate contribution activity
- `'▓'` High level: Heavy contribution activity
- `'╻┃╽'` Top level: Last block with contributions in the week (Low, Medium, High)

## Visualizing your Skyline

Once you have generated your STL file, you can visualize it using 3D modeling or 3D printing software. But did you know that you can upload your STL file to a GitHub repository and view your Skyline there? For example, take a look at [@chrisreddington's GitHub Skyline from 2011 - 2024](https://github.com/chrisreddington/chrisreddington/blob/master/chrisreddington-11-24-github-skyline.stl).

## Project Structure

```text
├── ascii/
│   ├── block.go: ASCII block character definitions for contribution levels
│   ├── block_test.go: Block character unit tests
│   ├── generator.go: Contribution visualization ASCII art generation
│   ├── generator_test.go: ASCII generation tests
│   ├── text.go: ASCII text formatting utilities
│   └── text_test.go: Text formatting unit tests
├── errors/
│   ├── errors.go: Custom error types and domain-specific error handling
│   └── errors_test.go: Error handling unit tests
├── github/
│   ├── client.go: GitHub API client for fetching contribution data
│   └── client_test.go: API client unit tests
├── logger/
│   ├── logger.go: Thread-safe logging with severity levels
│   └── logger_test.go: Logger unit tests
├── stl/
│   ├── generator.go: STL 3D model generation from contribution data
│   ├── generator_test.go: Model generation unit tests
│   ├── stl.go: STL binary file format implementation
│   ├── stl_test.go: STL file generation tests
│   └── geometry/
│       ├── geometry.go: 3D geometry calculations and transformations
│       ├── geometry_test.go: Geometry unit tests
│       ├── shapes.go: Basic 3D primitive shape definitions
│       ├── text.go: 3D text geometry generation
│       └── text_test.go: Text geometry unit tests
├── types/
│   ├── types.go: Shared data structures and interfaces
│   └── types_test.go: Data structure unit tests
└── main.go: CLI application entry point
```

## Contributing

To contribute to the project, please read the instructions and contributing guidelines in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is licensed under the terms of the MIT open source license. Please refer to [MIT](./LICENSE) for the full terms.

Mona Sans is licensed under the [SIL Open Font License v1.1](https://scripts.sil.org/OFL). Find more details at [github/mona-sans](https://github.com/github/mona-sans).

[golang/freetype](https://github.com/golang/freetype) is used as a dependency. Portions of this software are copyright © 2024 The FreeType Project ([www.freetype.org](https://www.freetype.org)). All rights reserved.

## Acknowledgements

- The Invertocat is subject to [GitHub Logos and Usage guidelines](https://github.com/logos).
- The [Mona Sans](https://github.com/github/mona-sans) typeface.
- The [GitHub CLI](https://cli.github.com/) team for the CLI and extension framework.
- [golang/freetype](https://github.com/golang/freetype).
