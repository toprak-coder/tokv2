# tokV2

A high-performance, URL-aware tokenizer designed to extract meaningful keywords from lists of URLs for security assessments.

## Philosophy

In security testing, particularly during reconnaissance and content discovery, the quality of your wordlists is paramount. While generic wordlists are a good starting point, target-specific keywords often yield the most interesting results.

The original `tok` by [TomNomNom](https://github.com/tomnomnom/hacks/tree/master/tok) pioneered the idea of generating these keywords by tokenizing URLs. `tokV2` builds on this foundation with a more powerful, URL-aware parsing engine. Instead of just splitting strings, `tokV2` understands the structure of a URL, allowing it to intelligently separate subdomains, path segments, parameter names, and their values. This leads to cleaner, more contextually relevant wordlists, saving you time and helping you discover hidden endpoints and parameters more effectively.

## Features

-   **Context-Aware Tokenization:** Intelligently dissects URLs into their core components (subdomains, paths, parameters, fragments) before tokenizing.
-   **Advanced Filtering Engine:** Drill down to the exact keywords you need with substring (`-f`), regex (`-r`), length (`-min`, `-max`), and alphanumeric (`-alpha-num-only`) filters.
-   **Categorized File Output:** Automatically sort your findings by saving path tokens (`-o`) and parameter names (`-op`) into dedicated files.
-   **Pipeline-Friendly:** Designed to be a core part of your command-line workflow. It reads from `stdin` and writes to `stdout`, allowing you to pipe data from and to other tools seamlessly.
-   **Robust and Fast:** Built in Go with concurrency in mind to process large lists of URLs quickly. Includes a fallback mechanism for non-URL strings.
-   **Unique Output:** Automatically deduplicates all generated tokens to ensure your final wordlist is clean and ready to use.

## Installation

> **Note:** This method requires the project to be hosted in a public Git repository (e.g., on GitHub).

With Go installed, you can install `tokv2` directly from its GitHub repository:

```bash
go install github.com/toprak-coder/tokv2@latest
```

This command will download the source, compile it, and place the `tokv2` binary in your Go bin directory (`$HOME/go/bin` by default). Ensure this directory is in your system's `$PATH`.

## Workflow Integration

`tokV2` is most powerful when combined with other tools. Here's a typical workflow:

1.  **Gather URLs:** Use tools like [gau](https://github.com/lc/gau), [waybackurls](https://github.com/tomnomnom/waybackurls), or [gospider](https://github.com/jaeles-project/gospider) to collect URLs for a target domain.
2.  **Generate Keywords:** Pipe the collected URLs into `tokv2` to create a target-specific wordlist.
3.  **Fuzz for Content:** Use the generated wordlist with a fuzzer like [ffuf](https://github.com/ffuf/ffuf) to discover hidden files and directories.

```bash
# Example Workflow
echo "example.com" | gau | tokv2 -o paths.txt
ffuf -w paths.txt -u https://example.com/FUZZ
```

## Usage

```bash
cat urls.txt | tokv2 [flags]
```

### Flags

| Flag             | Description                                                              | Default |
| ---------------- | ------------------------------------------------------------------------ | ------- |
| `-min`           | Minimum length of a token to be included in the output.                  | `1`     |
| `-max`           | Maximum length of a token to be included in the output.                  | `25`    |
| `-alpha-num-only`| Only output tokens that contain at least one letter and one number.      | `false` |
| `-f`             | Filter output to tokens containing the specified string.                 | `""`    |
| `-r`             | Filter output to tokens matching the specified regex pattern.            | `""`    |
| `-o`             | Output file to write **path-related** tokens to.                         | `""`    |
| `-op`            | Output file to write **parameter name** tokens to.                       | `""`    |

## Examples

### Example 1: Basic Tokenization

Given a file `urls.txt` with the content:
```
https://api.example.com/v1/users?id=123&role=guest
```

Running the tool without flags will produce all unique tokens:
```bash
cat urls.txt | tokv2
```

**Output:**
```
api
example
com
v1
users
id
123
role
guest
```

### Example 2: Filtering for Specific Keywords

Find all tokens containing the word "user".

```bash
cat urls.txt | tokv2 -f "user"
```

**Output:**
```
users
```

### Example 3: Extracting Numerical IDs with Regex

Find all tokens that consist only of numbers.

```bash
cat urls.txt | tokv2 -r "^[0-9]+$"
```

**Output:**
```
123
```

### Example 4: Saving Paths and Parameter Names to Files

Process a list of URLs and save all path-related keywords to `paths.txt` and all parameter names to `params.txt`. Subdomain parts and parameter values will be printed to the screen.

```bash
cat urls.txt | tokv2 -o paths.txt -op params.txt
```

**`paths.txt` will contain:**
```
v1
users
```

**`params.txt` will contain:**
```
id
role
```

**Standard Output will show:**
```
api
example
com
123
guest
```
